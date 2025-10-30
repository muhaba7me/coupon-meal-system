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

func GetAllEmployees(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		//get parameters for filtering
		status := c.Query("status")
		search := c.Query("search")

		//build filter
		filter := bson.D{}
		if status != "" {
			filter = append(filter, bson.E{Key: "status", Value: status})
		}
		if search != "" {
			filter = append(filter, bson.E{Key: "$or", Value: bson.A{
				bson.D{{Key: "name", Value: bson.D{{Key: "$regex", Value: search}, {Key: "$options", Value: "i"}}}},
				bson.D{{Key: "employee_code", Value: bson.D{{Key: "$regex", Value: search}, {Key: "$options", Value: "i"}}}},
				bson.D{{Key: "email", Value: bson.D{{Key: "$regex", Value: search}, {Key: "$options", Value: "i"}}}},
			}})
		}
		employeeCollection := database.OpenCollection("employees", client)
		cursor, err := employeeCollection.Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Filed to fetch employees"})
		}

		defer cursor.Close(ctx)
		var employees []models.Employee
		if err = cursor.All(ctx, &employees); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode employee coupon"})
			return
		}

		// convert response format
		var response []models.EmployeeResponse
		for _, emp := range employees {
			response = append(response, models.EmployeeResponse{
				EmployeeID:         emp.EmployeeID,
				EmployeeCode:       emp.EmployeeCode,
				Name:               emp.Name,
				Email:              emp.Email,
				Phone:              emp.Phone,
				Status:             emp.Status,
				MonthlyAllocation:  emp.MonthlyAllocation,
				CurrentBalance:     emp.CurrentBalance,
				LastAllocationDate: emp.LastAllocationDate,
				HireDate:           emp.HireDate,
				IsVerified:         emp.IsVerified,
				CreatedAt:          emp.CreatedAt,
			})
		}
		c.JSON(http.StatusOK, gin.H{
			"employees": response,
			"total":     len(response),
		})

	}
}

// GetEmployeeByID - Get single employee details
func GetEmployeeByID(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		employeeID := c.Param("id")
		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		employeeCollection := database.OpenCollection("employees", client)
		var employee models.Employee
		err := employeeCollection.FindOne(ctx, bson.D{{Key: "employee_id", Value: employeeID}}).Decode(&employee)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}

		c.JSON(http.StatusOK, employee)
	}
}

// Get employeeybycode- get employee by employee code
func GetEmployeeByCode(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		employeeCode := c.Param("code")
		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		employeeCollection := database.OpenCollection("employees", client)
		var employee models.Employee
		err := employeeCollection.FindOne(ctx, bson.D{{Key: "employee_code", Value: employeeCode}}).Decode(&employee)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}

		c.JSON(http.StatusOK, employee)
	}
}

// UpdateEmployee - Update employee details\

func UpdateEmployee(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		employeeID := c.Param("id")
		var req models.UpdateEmployeeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invaild data input"})
			return
		}
		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		// Build update document
		updateData := bson.M{
			"updated_at": time.Now(),
		}

		if req.Name != "" {
			updateData["name"] = req.Name
		}
		if req.Phone != "" {
			updateData["phone"] = req.Phone
		}
		if req.Status != "" {
			// Validate status
			validStatuses := []string{"active", "on_leave", "suspended", "terminated"}
			isValid := false
			for _, status := range validStatuses {
				if req.Status == status {
					isValid = true
					break
				}
			}
			if !isValid {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Use: active, on_leave, suspended, terminated"})
				return
			}
			updateData["status"] = req.Status

			// If terminated, set termination date
			if req.Status == "terminated" {
				now := time.Now()
				updateData["termination_date"] = now
			}
		}
		if req.Notes != "" {
			updateData["notes"] = req.Notes
		}

		employeeCollection := database.OpenCollection("employees", client)
		result, err := employeeCollection.UpdateOne(
			ctx,
			bson.D{{Key: "employee_id", Value: employeeID}},
			bson.D{{Key: "$set", Value: updateData}},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update employee"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Employee updated successfully",
			"updated": result.ModifiedCount,
		})
	}

}

// GetMyProfile - Employee gets their own profile
func GetMyProfile(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := utils.GetUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		employeeCollection := database.OpenCollection("employees", client)
		var employee models.Employee
		err = employeeCollection.FindOne(ctx, bson.D{{Key: "user_id", Value: userID}}).Decode(&employee)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee profile not found"})
			return
		}

		// Update last login
		now := time.Now()
		employeeCollection.UpdateOne(
			ctx,
			bson.D{{Key: "employee_id", Value: employee.EmployeeID}},
			bson.D{{Key: "$set", Value: bson.M{"last_login": now}}},
		)

		c.JSON(http.StatusOK, employee)
	}
}

// GetMyBalance - Employee checks their coupon balance
func GetMyBalance(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := utils.GetUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		employeeCollection := database.OpenCollection("employees", client)
		var employee models.Employee
		err = employeeCollection.FindOne(ctx, bson.D{{Key: "user_id", Value: userID}}).Decode(&employee)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"employee_code":        employee.EmployeeCode,
			"name":                 employee.Name,
			"current_balance":      employee.CurrentBalance,
			"monthly_allocation":   employee.MonthlyAllocation,
			"last_allocation_date": employee.LastAllocationDate,
			"status":               employee.Status,
		})
	}
}
