package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Cxrlo19/stackd/api/internal/db"
	"github.com/Cxrlo19/stackd/api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateTransactionRequest struct {
	Name     string  `json:"name" binding:"required"`
	Amount   float64 `json:"amount" binding:"required"`
	Category string  `json:"category"`
	Date     string  `json:"date" binding:"required"` // YYYY-MM-DD
	Notes    string  `json:"notes"`
}

func CreateTransaction(c *gin.Context) {
	userID, _ := c.Get("userID")
	accountID := c.Param("id")

	// Verify account belongs to user
	var account models.BankAccount
	if err := db.DB.Where("id = ? AND user_id = ?", accountID, userID).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	// Build transaction
	transaction := models.Transaction{
		ID:        uuid.New(),
		UserID:    userID.(uuid.UUID),
		AccountID: account.ID,
		Name:      req.Name,
		Amount:    req.Amount,
		Category:  req.Category,
		Date:      date,
	}

	if req.Notes != "" {
		transaction.Notes = &req.Notes
	}

	if err := db.DB.Create(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	// Update account balance
	// Negative amount = expense, positive = income
	db.DB.Model(&account).Update("balance", account.Balance-req.Amount)

	c.JSON(http.StatusCreated, gin.H{"transaction": transaction})
}

func GetTransactions(c *gin.Context) {
	userID, _ := c.Get("userID")
	accountID := c.Param("id")

	// Verify account belongs to user
	var account models.BankAccount
	if err := db.DB.Where("id = ? AND user_id = ?", accountID, userID).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	// Filtering
	category := c.Query("category")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Build query
	query := db.DB.Where("account_id = ?", accountID)

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}

	// Get total count
	var total int64
	query.Model(&models.Transaction{}).Count(&total)

	// Get paginated results
	var transactions []models.Transaction
	query.Order("date DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions)

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func GetSpendingSummary(c *gin.Context) {
	userID, _ := c.Get("userID")

	// Default to current month
	now := time.Now()
	startDate := c.DefaultQuery("start_date", now.Format("2006-01-02")[:8]+"01")
	endDate := c.DefaultQuery("end_date", now.Format("2006-01-02"))

	type CategorySummary struct {
		Category string  `json:"category"`
		Total    float64 `json:"total"`
		Count    int     `json:"count"`
	}

	var summary []CategorySummary
	db.DB.Model(&models.Transaction{}).
		Select("category, SUM(amount) as total, COUNT(*) as count").
		Where("user_id = ? AND date >= ? AND date <= ? AND amount > 0", userID, startDate, endDate).
		Group("category").
		Order("total DESC").
		Scan(&summary)

	// Total spending
	var totalSpending float64
	for _, s := range summary {
		totalSpending += s.Total
	}

	c.JSON(http.StatusOK, gin.H{
		"summary":        summary,
		"total_spending": totalSpending,
		"period": gin.H{
			"start": startDate,
			"end":   endDate,
		},
	})
}
