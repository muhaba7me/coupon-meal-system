
package controllers

import (
	"context"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/muhaba7me/coupon-meal-system/database"
	"github.com/muhaba7me/coupon-meal-system/models"
	"github.com/muhaba7me/coupon-meal-system/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func InitiateTransaction(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.InitiateTransactionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
			return
		}

		// Get supplier user ID from context
		supplierUserID, err := utils.GetUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		// 1. Get Supplier Profile
		supplierCollection := database.OpenCollection("suppliers", client)
		var supplier models.Supplier
		err = supplierCollection.FindOne(ctx, bson.D{{Key: "user_id", Value: supplierUserID}}).Decode(&supplier)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Supplier profile not found. Please contact admin."})
			return
		}

		// 2. Validate Supplier Status
		if !supplier.IsActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "Your supplier account has been deactivated. Please contact admin."})
			return
		}

		if !supplier.IsVerified {
			c.JSON(http.StatusForbidden, gin.H{"error": "Your supplier account is pending verification by admin."})
			return
		}

		// 3. Validate QR Code
		qrCollection := database.OpenCollection("qr_codes", client)
		var qrCode models.QRCode
		err = qrCollection.FindOne(ctx, bson.D{{Key: "code", Value: req.QRCode}}).Decode(&qrCode)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid QR code"})
			return
		}

		if qrCode.IsUsed {
			c.JSON(http.StatusBadRequest, gin.H{"error": "QR code has already been used"})
			return
		}

		if time.Now().After(qrCode.ExpiresAt) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "QR code has expired. Please ask employee to generate a new one.",
				"expired_at": qrCode.ExpiresAt,
			})
			return
		}

		// 4. Get Employee
		employeeCollection := database.OpenCollection("employees", client)
		var employee models.Employee
		err = employeeCollection.FindOne(ctx, bson.D{{Key: "employee_id", Value: qrCode.EmployeeID}}).Decode(&employee)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}

		// 5. Validate Employee Status
		if employee.Status == "terminated" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Employee account has been terminated"})
			return
		}

		if employee.Status == "suspended" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Employee account is currently suspended"})
			return
		}

		// Allow "active" and "on_leave" status
		if employee.Status != "active" && employee.Status != "on_leave" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Employee account status does not allow transactions"})
			return
		}

		// 6. Location Validation (if coordinates provided)
		if req.Latitude != 0 && req.Longitude != 0 {
			isWithinRadius := utils.ValidateLocation(
				supplier.Latitude,
				supplier.Longitude,
				req.Latitude,
				req.Longitude,
				supplier.LocationRadius,
			)

			if !isWithinRadius {
				distance := utils.CalculateDistance(
					supplier.Latitude,
					supplier.Longitude,
					req.Latitude,
					req.Longitude,
				)

				c.JSON(http.StatusForbidden, gin.H{
					"error": "Employee is outside the allowed location radius",
					"details": gin.H{
						"distance_meters": int(distance),
						"allowed_radius_meters": supplier.LocationRadius,
						"supplier_name": supplier.BusinessName,
						"supplier_address": supplier.Address,
					},
					"message": "Employee must be within " + string(rune(supplier.LocationRadius)) + " meters of your location",
				})
				return
			}
		}

		// 7. Validate Coupons Range
		if req.CouponsUsed < 1 || req.CouponsUsed > 3 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid coupon amount",
				"details": "You can only charge between 1 and 3 coupons per transaction",
			})
			return
		}

		// 8. Check Employee Balance
		if employee.CurrentBalance < req.CouponsUsed {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Insufficient coupon balance",
				"employee_balance": employee.CurrentBalance,
				"requested_coupons": req.CouponsUsed,
				"message": "Employee only has " + string(rune(employee.CurrentBalance)) + " coupons available",
			})
			return
		}

		// 9. Calculate Amount
		couponValue := 45.0
		totalAmount := float64(req.CouponsUsed) * couponValue

		// 10. Create Transaction Record
		transactionID := uuid.New().String()
		transaction := models.Transaction{
			TransactionID:     transactionID,
			EmployeeID:        employee.EmployeeID,
			SupplierID:        supplier.SupplierID,
			QRCodeID:          qrCode.QRCodeID,
			CouponsUsed:       req.CouponsUsed,
			TotalAmount:       totalAmount,
			EmployeeLatitude:  req.Latitude,
			EmployeeLongitude: req.Longitude,
			Status:            "pending",
			Notes:             req.Notes,
			ProcessedAt:       time.Now(),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		transactionCollection := database.OpenCollection("transactions", client)
		_, err = transactionCollection.InsertOne(ctx, transaction)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
			return
		}

		// 11. Return Success Response
		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "Transaction initiated successfully. Waiting for employee approval.",
			"transaction_id": transactionID,
			"employee": gin.H{
				"name": employee.Name,
				"code": employee.EmployeeCode,
				"current_balance": employee.CurrentBalance,
				"new_balance": employee.CurrentBalance - req.CouponsUsed,
			},
			"supplier": gin.H{
				"name": supplier.BusinessName,
				"address": supplier.Address,
			},
			"transaction": gin.H{
				"coupons_used": req.CouponsUsed,
				"total_amount": totalAmount,
				"status": "pending",
			},
			"requires_approval": true,
		})
	}
}

// ApproveTransaction - Employee approves or rejects transaction
func ApproveTransaction(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.ApproveTransactionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
			return
		}

		employeeUserID, err := utils.GetUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		// 1. Get Transaction
		transactionCollection := database.OpenCollection("transactions", client)
		var transaction models.Transaction
		err = transactionCollection.FindOne(ctx, bson.D{{Key: "transaction_id", Value: req.TransactionID}}).Decode(&transaction)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}

		// 2. Verify Employee Ownership
		employeeCollection := database.OpenCollection("employees", client)
		var employee models.Employee
		err = employeeCollection.FindOne(ctx, bson.D{
			{Key: "user_id", Value: employeeUserID},
			{Key: "employee_id", Value: transaction.EmployeeID},
		}).Decode(&employee)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only approve your own transactions"})
			return
		}

		// 3. Check Transaction Status
		if transaction.Status != "pending" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Transaction has already been processed",
				"current_status": transaction.Status,
			})
			return
		}

		// 4. Get Supplier Info (for response)
		supplierCollection := database.OpenCollection("suppliers", client)
		var supplier models.Supplier
		supplierCollection.FindOne(ctx, bson.D{{Key: "supplier_id", Value: transaction.SupplierID}}).Decode(&supplier)

		// 5. Process Based on Approval
		if req.Approved {
			if employee.CurrentBalance < transaction.CouponsUsed {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Insufficient balance",
					"current_balance": employee.CurrentBalance,
					"required": transaction.CouponsUsed,
				})
				return
			}

			// Use MongoDB transaction for atomicity
			session, err := client.StartSession()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start database session"})
				return
			}
			defer session.EndSession(ctx)

			_, err = session.WithTransaction(ctx, func(sessCtx context.Context) (interface{}, error) {
				// Deduct employee balance
				_, err := employeeCollection.UpdateOne(
					sessCtx,
					bson.D{{Key: "employee_id", Value: employee.EmployeeID}},
					bson.D{{Key: "$inc", Value: bson.D{{Key: "current_coupon_balance", Value: -transaction.CouponsUsed}}}},
				)
				if err != nil {
					return nil, err
				}

				// Mark QR code as used
				qrCollection := database.OpenCollection("qr_codes", client)
				now := time.Now()
				_, err = qrCollection.UpdateOne(
					sessCtx,
					bson.D{{Key: "qr_code_id", Value: transaction.QRCodeID}},
					bson.D{{Key: "$set", Value: bson.D{
						{Key: "is_used", Value: true},
						{Key: "used_at", Value: now},
					}}},
				)
				if err != nil {
					return nil, err
				}

				// Update transaction status
				_, err = transactionCollection.UpdateOne(
					sessCtx,
					bson.D{{Key: "transaction_id", Value: req.TransactionID}},
					bson.D{{Key: "$set", Value: bson.D{
						{Key: "status", Value: "completed"},
						{Key: "updated_at", Value: now},
					}}},
				)
				if err != nil {
					return nil, err
				}

				return nil, nil
			})

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process transaction"})
				return
			}

			// Success Response
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Transaction approved and processed successfully",
				"transaction_id": req.TransactionID,
				"employee": gin.H{
					"name": employee.Name,
					"previous_balance": employee.CurrentBalance,
					"new_balance": employee.CurrentBalance - transaction.CouponsUsed,
					"coupons_deducted": transaction.CouponsUsed,
				},
				"supplier": gin.H{
					"name": supplier.BusinessName,
					"address": supplier.Address,
				},
				"transaction": gin.H{
					"amount": transaction.TotalAmount,
					"status": "completed",
				},
			})

		} else {

			now := time.Now()
			notes := req.Reason
			if notes == "" {
				notes = "Rejected by employee"
			}

			_, err = transactionCollection.UpdateOne(
				ctx,
				bson.D{{Key: "transaction_id", Value: req.TransactionID}},
				bson.D{{Key: "$set", Value: bson.D{
					{Key: "status", Value: "rejected"},
					{Key: "notes", Value: notes},
					{Key: "updated_at", Value: now},
				}}},
			)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject transaction"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Transaction rejected",
				"transaction_id": req.TransactionID,
				"status": "rejected",
				"reason": notes,
			})
		}
	}
}

// GetMyTransactions - Employee views transaction history
func GetMyTransactions(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		employeeUserID, err := utils.GetUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		// Get employee
		employeeCollection := database.OpenCollection("employees", client)
		var employee models.Employee
		err = employeeCollection.FindOne(ctx, bson.D{{Key: "user_id", Value: employeeUserID}}).Decode(&employee)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}

		// Get transactions
		transactionCollection := database.OpenCollection("transactions", client)
		cursor, err := transactionCollection.Find(ctx, bson.D{{Key: "employee_id", Value: employee.EmployeeID}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
			return
		}
		defer cursor.Close(ctx)

		var transactions []models.Transaction
		if err = cursor.All(ctx, &transactions); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode transactions"})
			return
		}

		// Enrich with supplier info
		supplierCollection := database.OpenCollection("suppliers", client)
		type EnrichedTransaction struct {
			models.Transaction
			SupplierName    string `json:"supplier_name"`
			SupplierAddress string `json:"supplier_address"`
		}

		var enriched []EnrichedTransaction
		for _, tx := range transactions {
			var supplier models.Supplier
			supplierCollection.FindOne(ctx, bson.D{{Key: "supplier_id", Value: tx.SupplierID}}).Decode(&supplier)

			enriched = append(enriched, EnrichedTransaction{
				Transaction:     tx,
				SupplierName:    supplier.BusinessName,
				SupplierAddress: supplier.Address,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"transactions": enriched,
			"total": len(enriched),
		})
	}
}

// GetSupplierTransactions - Supplier views their transactions
func GetSupplierTransactions(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		supplierUserID, err := utils.GetUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		// Get supplier
		supplierCollection := database.OpenCollection("suppliers", client)
		var supplier models.Supplier
		err = supplierCollection.FindOne(ctx, bson.D{{Key: "user_id", Value: supplierUserID}}).Decode(&supplier)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Supplier not found"})
			return
		}

		// Get transactions
		transactionCollection := database.OpenCollection("transactions", client)
		cursor, err := transactionCollection.Find(ctx, bson.D{{Key: "supplier_id", Value: supplier.SupplierID}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
			return
		}
		defer cursor.Close(ctx)

		var transactions []models.Transaction
		cursor.All(ctx, &transactions)

		// Calculate statistics
		totalCoupons := 0
		totalAmount := 0.0
		completedCount := 0
		pendingCount := 0

		for _, tx := range transactions {
			switch tx.Status {
case "completed":
				totalCoupons += tx.CouponsUsed
				totalAmount += tx.TotalAmount
				completedCount++
			case "pending":
				pendingCount++
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"transactions": transactions,
			"statistics": gin.H{
				"total_transactions": len(transactions),
				"completed": completedCount,
				"pending": pendingCount,
				"total_coupons": totalCoupons,
				"total_amount": totalAmount,
			},
		})
	}
}

// GetPendingTransactions - Employee sees pending approvals
func GetPendingTransactions(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		employeeUserID, err := utils.GetUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		employeeCollection := database.OpenCollection("employees", client)
		var employee models.Employee
		err = employeeCollection.FindOne(ctx, bson.D{{Key: "user_id", Value: employeeUserID}}).Decode(&employee)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}

		transactionCollection := database.OpenCollection("transactions", client)
		cursor, err := transactionCollection.Find(ctx, bson.D{
			{Key: "employee_id", Value: employee.EmployeeID},
			{Key: "status", Value: "pending"},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
			return
		}
		defer cursor.Close(ctx)

		var transactions []models.Transaction
		cursor.All(ctx, &transactions)

		c.JSON(http.StatusOK, gin.H{
			"pending_transactions": transactions,
			"count": len(transactions),
		})
	}
}
