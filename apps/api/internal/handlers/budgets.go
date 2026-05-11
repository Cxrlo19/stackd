package handlers

import (
	"net/http"
	"time"

	"github.com/Cxrlo19/stackd/api/internal/db"
	"github.com/Cxrlo19/stackd/api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateBudgetRequest struct {
	Category string  `json:"category" binding:"required"`
	Amount   float64 `json:"amount" binding:"required"`
	Period   string  `json:"period"`
}

func CreateBudget(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	period := req.Period
	if period == "" {
		period = "monthly"
	}

	// Check if budget already exists
	var existing models.Budget
	result := db.DB.Where("user_id = ? AND category = ? AND period = ?",
		userID, req.Category, period).First(&existing)

	if result.Error == nil {
		// Update existing budget
		db.DB.Model(&existing).Update("amount", req.Amount)
		existing.Amount = req.Amount
		c.JSON(http.StatusOK, gin.H{"budget": existing})
		return
	}

	// Create new budget
	budget := models.Budget{
		ID:       uuid.New(),
		UserID:   userID.(uuid.UUID),
		Category: req.Category,
		Amount:   req.Amount,
		Period:   period,
	}

	if err := db.DB.Create(&budget).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create budget"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"budget": budget})
}

func GetBudgets(c *gin.Context) {
	userID, _ := c.Get("userID")

	var budgets []models.Budget
	if err := db.DB.Where("user_id = ?", userID).Find(&budgets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get budgets"})
		return
	}

	// Enrich with current spending for each budget category
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	type BudgetWithSpending struct {
		models.Budget
		Spent      float64 `json:"spent"`
		Remaining  float64 `json:"remaining"`
		Percentage float64 `json:"percentage"`
	}

	var enriched []BudgetWithSpending
	for _, budget := range budgets {
		var spent float64
		db.DB.Model(&models.Transaction{}).
			Select("COALESCE(SUM(amount), 0)").
			Where("user_id = ? AND category = ? AND date >= ? AND amount > 0",
				userID, budget.Category, startOfMonth).
			Scan(&spent)

		remaining := budget.Amount - spent
		percentage := 0.0
		if budget.Amount > 0 {
			percentage = (spent / budget.Amount) * 100
		}

		enriched = append(enriched, BudgetWithSpending{
			Budget:     budget,
			Spent:      spent,
			Remaining:  remaining,
			Percentage: percentage,
		})
	}

	c.JSON(http.StatusOK, gin.H{"budgets": enriched})
}

func DeleteBudget(c *gin.Context) {
	userID, _ := c.Get("userID")
	budgetID := c.Param("id")

	result := db.DB.Where("id = ? AND user_id = ?", budgetID, userID).
		Delete(&models.Budget{})

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Budget deleted"})
}

func GetBudgetAlerts(c *gin.Context) {
	userID, _ := c.Get("userID")

	var budgets []models.Budget
	db.DB.Where("user_id = ?", userID).Find(&budgets)

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	type Alert struct {
		Category   string  `json:"category"`
		Budget     float64 `json:"budget"`
		Spent      float64 `json:"spent"`
		Percentage float64 `json:"percentage"`
		Status     string  `json:"status"` // "ok", "warning", "exceeded"
	}

	var alerts []Alert
	for _, budget := range budgets {
		var spent float64
		db.DB.Model(&models.Transaction{}).
			Select("COALESCE(SUM(amount), 0)").
			Where("user_id = ? AND category = ? AND date >= ? AND amount > 0",
				userID, budget.Category, startOfMonth).
			Scan(&spent)

		percentage := 0.0
		if budget.Amount > 0 {
			percentage = (spent / budget.Amount) * 100
		}

		status := "ok"
		if percentage >= 100 {
			status = "exceeded"
		} else if percentage >= 80 {
			status = "warning"
		}

		// Only include budgets that need attention
		if status != "ok" {
			alerts = append(alerts, Alert{
				Category:   budget.Category,
				Budget:     budget.Amount,
				Spent:      spent,
				Percentage: percentage,
				Status:     status,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"alerts": alerts})
}
