package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Cxrlo19/stackd/api/internal/db"
	"github.com/Cxrlo19/stackd/api/internal/handlers"
	"github.com/Cxrlo19/stackd/api/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db.Connect()

	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://stackd-nu.vercel.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "stackd-api"})
	})

	// Auth routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
	}

	// Protected routes
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("/me", handlers.Me)

		// Accounts routes
		api.POST("/accounts", handlers.CreateAccount)
		api.GET("/accounts", handlers.GetAccounts)
		api.GET("/accounts/:id", handlers.GetAccount)
		api.DELETE("/accounts/:id", handlers.DeleteAccount)

		// Transactions routes
		api.POST("/accounts/:id/transactions", handlers.CreateTransaction)
		api.GET("/accounts/:id/transactions", handlers.GetTransactions)
		api.GET("/transactions/summary", handlers.GetSpendingSummary)

		// Budget routes
		api.POST("/budgets", handlers.CreateBudget)
		api.GET("/budgets", handlers.GetBudgets)
		api.GET("/budgets/alerts", handlers.GetBudgetAlerts)
		api.DELETE("/budgets/:id", handlers.DeleteBudget)

		// Insights routes
		api.POST("/insights/weekly", handlers.GenerateWeeklyInsight)
		api.POST("/insights/budget", handlers.GenerateBudgetInsight)
		api.GET("/insights", handlers.GetInsights)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Stackd API running on port %s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}
