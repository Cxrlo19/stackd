package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Cxrlo19/stackd/api/internal/db"
	"github.com/Cxrlo19/stackd/api/internal/models"
	"github.com/Cxrlo19/stackd/api/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GenerateWeeklyInsight(c *gin.Context) {
	userID, _ := c.Get("userID")

	// Get last 7 days of transactions
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	type CategorySpend struct {
		Category string
		Total    float64
		Count    int
	}

	var categorySpends []CategorySpend
	db.DB.Model(&models.Transaction{}).
		Select("category, SUM(amount) as total, COUNT(*) as count").
		Where("user_id = ? AND date >= ? AND amount > 0", userID, sevenDaysAgo).
		Group("category").
		Order("total DESC").
		Scan(&categorySpends)

	// Get total spending
	var totalSpending float64
	for _, s := range categorySpends {
		totalSpending += s.Total
	}

	// Get budgets for context
	var budgets []models.Budget
	db.DB.Where("user_id = ?", userID).Find(&budgets)

	// Build prompt
	prompt := fmt.Sprintf(`Here is my spending data for the last 7 days:
Total spent: $%.2f

Spending by category:`, totalSpending)

	for _, s := range categorySpends {
		prompt += fmt.Sprintf("\n- %s: $%.2f (%d transactions)", s.Category, s.Total, s.Count)
	}

	if len(budgets) > 0 {
		prompt += "\n\nMy monthly budgets:"
		for _, b := range budgets {
			prompt += fmt.Sprintf("\n- %s: $%.2f", b.Category, b.Amount)
		}
	}

	prompt += "\n\nPlease provide a brief analysis of my spending and 2-3 specific actionable tips."

	// Generate insight with Groq
	content, err := services.GenerateInsight(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate insight"})
		return
	}

	// Store insight in database
	insight := models.Insight{
		ID:      uuid.New(),
		UserID:  userID.(uuid.UUID),
		Content: content,
		Type:    "weekly",
	}

	db.DB.Create(&insight)

	c.JSON(http.StatusOK, gin.H{
		"insight": insight,
		"data": gin.H{
			"total_spending":  totalSpending,
			"category_spends": categorySpends,
			"period":          "last 7 days",
		},
	})
}

func GetInsights(c *gin.Context) {
	userID, _ := c.Get("userID")

	var insights []models.Insight
	db.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(10).
		Find(&insights)

	c.JSON(http.StatusOK, gin.H{"insights": insights})
}

func GenerateBudgetInsight(c *gin.Context) {
	userID, _ := c.Get("userID")

	// Get current month spending by category
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	type CategorySpend struct {
		Category string
		Total    float64
	}

	var spends []CategorySpend
	db.DB.Model(&models.Transaction{}).
		Select("category, SUM(amount) as total").
		Where("user_id = ? AND date >= ? AND amount > 0", userID, startOfMonth).
		Group("category").
		Scan(&spends)

	var budgets []models.Budget
	db.DB.Where("user_id = ?", userID).Find(&budgets)

	if len(budgets) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No budgets set. Set budgets to get personalized insights."})
		return
	}

	// Build prompt with budget vs actual comparison
	prompt := fmt.Sprintf("It's day %d of the month. Here is my budget vs actual spending:\n", now.Day())

	for _, budget := range budgets {
		spent := 0.0
		for _, s := range spends {
			if s.Category == budget.Category {
				spent = s.Total
				break
			}
		}
		percentage := 0.0
		if budget.Amount > 0 {
			percentage = (spent / budget.Amount) * 100
		}
		prompt += fmt.Sprintf("\n- %s: spent $%.2f of $%.2f budget (%.0f%%)",
			budget.Category, spent, budget.Amount, percentage)
	}

	prompt += "\n\nGive me specific advice on which budgets need attention and how to stay on track for the rest of the month."

	content, err := services.GenerateInsight(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate insight"})
		return
	}

	insight := models.Insight{
		ID:      uuid.New(),
		UserID:  userID.(uuid.UUID),
		Content: content,
		Type:    "monthly",
	}

	db.DB.Create(&insight)

	c.JSON(http.StatusOK, gin.H{"insight": insight})
}
