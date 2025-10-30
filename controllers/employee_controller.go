package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/muhaba7me/coupon-meal-system/database"
	"github.com/muhaba7me/coupon-meal-system/models"
	"github.com/muhaba7me/coupon-meal-system/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func CreateEmployee(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
	var req models.CreateEmployeeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
			return
		}

		// Validate
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
			return
		}

		// Get admin user ID from context
		adminUserID, err := utils.GetUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}


		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		// Check if employee code already exists
		employeeCollection := database.OpenCollection("employees", client)
		count, err := employeeCollection.CountDocuments(ctx, bson.D{{Key: "employee_code", Value: req.EmployeeCode}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check employee code"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Employee code already exists"})
			return
		}
			// Check if user exists
		userCollection := database.OpenCollection("users", client)
		var user models.User
		err = userCollection.FindOne(ctx, bson.D{{Key: "user_id", Value: req.UserID}}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		// Parse hire date
		hireDate, err := time.Parse("2006-01-02", req.HireDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hire date format. Use YYYY-MM-DD"})
			return
		}
	// Get monthly allocation from env or default to 26
		monthlyAllocation := 26
		if alloc := utils.GetEnvAsInt("MONTHLY_ALLOCATION", 26); alloc > 0 {
			monthlyAllocation = alloc
		}

		// Create employee
		now := time.Now()
		employee := models.Employee{
			EmployeeID:         bson.NewObjectID().Hex(),
			UserID:             req.UserID,
			EmployeeCode:       req.EmployeeCode,
			Name:               req.Name,
			Email:              req.Email,
			Phone:              req.Phone,
			Status:             "active",
			MonthlyAllocation:  monthlyAllocation,
			CurrentBalance:     monthlyAllocation, 
			LastAllocationDate: &now,
			HireDate:           hireDate,
			CreatedByAdminID:   adminUserID,
			IsVerified:         true,
			Notes:              req.Notes,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
	result, err := employeeCollection.InsertOne(ctx, employee)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create employee"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":     "Employee created successfully",
			"employee_id": employee.EmployeeID,
			"result":      result,
		})
	}
}