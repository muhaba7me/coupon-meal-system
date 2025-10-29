package main

import (
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/muhaba7me/coupon-meal-system/database"
	routes "github.com/muhaba7me/coupon-meal-system/routes"
)

func main() {
	// Initialize Gin router
	router := gin.Default()

	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: unable to find .env file")
	}

	// Connect to MongoDB
	client := database.Connect()

	// Enable CORS
	router.Use(cors.Default())

	// Test route
	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello, Coupon meal system!")
	})

	// Setup routes
	routes.SetupUnProtectedRoutes(router, client)
	// routes.SetupProtectedRoutes(router, client)

	// Start server
	if err := router.Run(":8080"); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
