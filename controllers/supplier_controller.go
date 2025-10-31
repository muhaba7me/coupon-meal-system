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

func CreateSupplier(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateSupplierRequest
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

		// Check if user exists and has SUPPLIER role
		userCollection := database.OpenCollection("users", client)
		var user models.User
		err = userCollection.FindOne(ctx, bson.D{{Key: "user_id", Value: req.UserID}}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		if user.Role != "SUPPLIER" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User must have SUPPLIER role"})
			return
		}

		// Check if supplier already exists for this user
		supplierCollection := database.OpenCollection("suppliers", client)
		count, _ := supplierCollection.CountDocuments(ctx, bson.D{{Key: "user_id", Value: req.UserID}})
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Supplier profile already exists for this user"})
			return
		}

		locationRadius := req.LocationRadius
		if locationRadius == 0 {
			locationRadius = 500 
		}
		supplier := models.Supplier{
			SupplierID:       bson.NewObjectID().Hex(),
			UserID:           req.UserID,
			BusinessName:     req.BusinessName,
			BusinessLicense:  req.BusinessLicense,
			ContactPerson:    req.ContactPerson,
			Phone:            req.Phone,
			Email:            req.Email,
			Address:          req.Address,
			Latitude:         req.Latitude,
			Longitude:        req.Longitude,
			LocationRadius:   locationRadius,
			IsActive:         true,
			IsVerified:       false, 
			BankAccount:      req.BankAccount,
			TaxID:            req.TaxID,
			Notes:            req.Notes,
			CreatedByAdminID: adminUserID,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		_, err = supplierCollection.InsertOne(ctx, supplier)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create supplier"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":     "Supplier created successfully",
			"supplier_id": supplier.SupplierID,
			"business_name": supplier.BusinessName,
		})
	}
}

func GetAllSuppliers(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		// Query parameters for filtering
		status := c.Query("status")     
		verified := c.Query("verified") 

		// Build filter
		filter := bson.D{}
		if status == "active" {
			filter = append(filter, bson.E{Key: "is_active", Value: true})
		} else if status == "inactive" {
			filter = append(filter, bson.E{Key: "is_active", Value: false})
		}

		if verified == "true" {
			filter = append(filter, bson.E{Key: "is_verified", Value: true})
		} else if verified == "false" {
			filter = append(filter, bson.E{Key: "is_verified", Value: false})
		}

		supplierCollection := database.OpenCollection("suppliers", client)
		cursor, err := supplierCollection.Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suppliers"})
			return
		}
		defer cursor.Close(ctx)

		var suppliers []models.Supplier
		if err = cursor.All(ctx, &suppliers); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode suppliers"})
			return
		}

		// Convert to response format
		var response []models.SupplierResponse
		for _, sup := range suppliers {
			response = append(response, models.SupplierResponse{
				SupplierID:     sup.SupplierID,
				BusinessName:   sup.BusinessName,
				ContactPerson:  sup.ContactPerson,
				Phone:          sup.Phone,
				Email:          sup.Email,
				Address:        sup.Address,
				Latitude:       sup.Latitude,
				Longitude:      sup.Longitude,
				LocationRadius: sup.LocationRadius,
				IsActive:       sup.IsActive,
				IsVerified:     sup.IsVerified,
				CreatedAt:      sup.CreatedAt,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"suppliers": response,
			"total":     len(response),
		})
	}
}


func GetSupplierByID(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		supplierID := c.Param("id")

		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		supplierCollection := database.OpenCollection("suppliers", client)
		var supplier models.Supplier
		err := supplierCollection.FindOne(ctx, bson.D{{Key: "supplier_id", Value: supplierID}}).Decode(&supplier)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Supplier not found"})
			return
		}

		c.JSON(http.StatusOK, supplier)
	}
}

func UpdateSupplier(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		supplierID := c.Param("id")

		var req models.UpdateSupplierRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
			return
		}

		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		// Build update document
		updateData := bson.M{"updated_at": time.Now()}

		if req.BusinessName != "" {
			updateData["business_name"] = req.BusinessName
		}
		if req.ContactPerson != "" {
			updateData["contact_person"] = req.ContactPerson
		}
		if req.Phone != "" {
			updateData["phone"] = req.Phone
		}
		if req.Address != "" {
			updateData["address"] = req.Address
		}
		if req.Latitude != 0 {
			updateData["latitude"] = req.Latitude
		}
		if req.Longitude != 0 {
			updateData["longitude"] = req.Longitude
		}
		if req.LocationRadius > 0 {
			updateData["location_radius"] = req.LocationRadius
		}
		if req.BankAccount != "" {
			updateData["bank_account"] = req.BankAccount
		}
		if req.TaxID != "" {
			updateData["tax_id"] = req.TaxID
		}
		if req.Notes != "" {
			updateData["notes"] = req.Notes
		}

		supplierCollection := database.OpenCollection("suppliers", client)
		result, err := supplierCollection.UpdateOne(
			ctx,
			bson.D{{Key: "supplier_id", Value: supplierID}},
			bson.D{{Key: "$set", Value: updateData}},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update supplier"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Supplier not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Supplier updated successfully",
			"updated": result.ModifiedCount,
		})
	}
}


func ActivateSupplier(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		supplierID := c.Param("id")

		var req struct {
			IsActive bool `json:"is_active" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		supplierCollection := database.OpenCollection("suppliers", client)
		result, err := supplierCollection.UpdateOne(
			ctx,
			bson.D{{Key: "supplier_id", Value: supplierID}},
			bson.D{{Key: "$set", Value: bson.M{
				"is_active":  req.IsActive,
				"updated_at": time.Now(),
			}}},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update supplier"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Supplier not found"})
			return
		}

		status := "deactivated"
		if req.IsActive {
			status = "activated"
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Supplier " + status + " successfully",
		})
	}
}

func GetMySupplierProfile(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := utils.GetUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		supplierCollection := database.OpenCollection("suppliers", client)
		var supplier models.Supplier
		err = supplierCollection.FindOne(ctx, bson.D{{Key: "user_id", Value: userID}}).Decode(&supplier)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Supplier profile not found"})
			return
		}

		c.JSON(http.StatusOK, supplier)
	}
}

func GetMyTotals(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := utils.GetUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		// Get supplier
		supplierCollection := database.OpenCollection("suppliers", client)
		var supplier models.Supplier
		err = supplierCollection.FindOne(ctx, bson.D{{Key: "user_id", Value: userID}}).Decode(&supplier)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Supplier not found"})
			return
		}

		// Calculate totals from transactions
		transactionCollection := database.OpenCollection("transactions", client)

		// All completed transactions
		cursor, _ := transactionCollection.Find(ctx, bson.D{
			{Key: "supplier_id", Value: supplier.SupplierID},
			{Key: "status", Value: "completed"},
		})
		defer cursor.Close(ctx)

		var transactions []models.Transaction
		cursor.All(ctx, &transactions)

		totalCoupons := 0
		totalAmount := 0.0
		completedToday := 0
		completedThisMonth := 0
		earningsToday := 0.0
		earningsThisMonth := 0.0

		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

		for _, tx := range transactions {
			totalCoupons += tx.CouponsUsed
			totalAmount += tx.TotalAmount

			if tx.ProcessedAt.After(today) {
				completedToday++
				earningsToday += tx.TotalAmount
			}

			if tx.ProcessedAt.After(monthStart) {
				completedThisMonth++
				earningsThisMonth += tx.TotalAmount
			}
		}

		// Count pending transactions
		pendingCount, _ := transactionCollection.CountDocuments(ctx, bson.D{
			{Key: "supplier_id", Value: supplier.SupplierID},
			{Key: "status", Value: "pending"},
		})

		c.JSON(http.StatusOK, models.SupplierTotalsResponse{
			SupplierID:          supplier.SupplierID,
			BusinessName:        supplier.BusinessName,
			TotalTransactions:   len(transactions),
			TotalCoupons:        totalCoupons,
			TotalAmount:         totalAmount,
			PendingTransactions: int(pendingCount),
			CompletedToday:      completedToday,
			CompletedThisMonth:  completedThisMonth,
			EarningsToday:       earningsToday,
			EarningsThisMonth:   earningsThisMonth,
		})
	}
}