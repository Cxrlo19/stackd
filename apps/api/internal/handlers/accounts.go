package handlers

import (
	"net/http"
	"time"

	"github.com/Cxrlo19/stackd/api/internal/db"
	"github.com/Cxrlo19/stackd/api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateAccountRequest struct {
	InstitutionName string  `json:"institution_name" binding:"required"`
	AccountName     string  `json:"account_name" binding:"required"`
	AccountType     string  `json:"account_type" binding:"required"`
	Balance         float64 `json:"balance"`
	Currency        string  `json:"currency"`
}

func CreateAccount(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	now := time.Now()
	account := models.BankAccount{
		ID:               uuid.New(),
		UserID:           userID.(uuid.UUID),
		PlaidAccountID:   "manual_" + uuid.New().String(), // placeholder for manual accounts
		PlaidAccessToken: "manual",                        // placeholder
		InstitutionName:  req.InstitutionName,
		AccountName:      req.AccountName,
		AccountType:      req.AccountType,
		Balance:          req.Balance,
		Currency:         currency,
		LastSynced:       &now,
	}

	if err := db.DB.Create(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"account": account})
}

func GetAccounts(c *gin.Context) {
	userID, _ := c.Get("userID")

	var accounts []models.BankAccount
	if err := db.DB.Where("user_id = ?", userID).Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get accounts"})
		return
	}

	// Calculate total balance
	var totalBalance float64
	for _, account := range accounts {
		totalBalance += account.Balance
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts":      accounts,
		"total_balance": totalBalance,
	})
}

func GetAccount(c *gin.Context) {
	userID, _ := c.Get("userID")
	accountID := c.Param("id")

	var account models.BankAccount
	if err := db.DB.Where("id = ? AND user_id = ?", accountID, userID).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"account": account})
}

func DeleteAccount(c *gin.Context) {
	userID, _ := c.Get("userID")
	accountID := c.Param("id")

	result := db.DB.Where("id = ? AND user_id = ?", accountID, userID).Delete(&models.BankAccount{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted"})
}
