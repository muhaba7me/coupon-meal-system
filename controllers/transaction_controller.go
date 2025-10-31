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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invaild input data"})
			return
		}
		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		//validate QR code
		qrCollection := database.OpenCollection("qr_codes", client)
		var qrCode models.QRCode

		err := qrCollection.FindOne(ctx, bson.D{{Key: "code", Value: req.QRCode}}).Decode(&qrCode)

		if err != nil {
			c.JSON(http.StatusFound, gin.H{"error": "Invalid QR code"})
		}
		//check if QR code already used
		if qrCode.IsUsed {
			c.JSON(http.StatusBadRequest, gin.H{"error": "QR code already used"})
			return
		}

		if time.Now().After(qrCode.ExpiresAt) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "QR code expired"})
			return
		}
		// get employee

		employeeCollection := database.OpenCollection("employees", client)
		var employee models.Employee
		err = employeeCollection.FindOne(ctx, bson.D{{Key: "employee_id", Value: employee.EmployeeID}}).Decode(&employee)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}

		//Validate Employee status
		if employee.Status != "active" && employee.Status != "on_leave" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Employee accoun is not active"})
			return
		}

		//Check balance
		if employee.CurrentBalance < req.CouponsUsed {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":           "Insufficient coupon balance",
				"current_balance": employee.CurrentBalance,
				"requested":       req.CouponsUsed,
			})
		}
		//validate coupons (1-3 only)
		if req.CouponsUsed < 1 || req.CouponsUsed > 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Coupon must be in between 1 and 3"})
		}

		couponValue := 45.0
		totalAmount := float64(req.CouponsUsed) * couponValue

		transactionID := uuid.New().String()

		transaction := models.Transaction{
			TransactionID:     transactionID,
			EmployeeID:        employee.EmployeeID,
			SupplierID:        req.SupplierID,
			QRCodeID:          req.QRCode,
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
		c.JSON(http.StatusCreated, models.TransactionInitiatedResponse{
			TransactionID:    transactionID,
			EmployeeName:     employee.Name,
			EmployeeCode:     employee.EmployeeCode,
			CurrentBalance:   employee.CurrentBalance,
			CouponsToDeduct:  req.CouponsUsed,
			TotalAmount:      totalAmount,
			NewBalance:       employee.CurrentBalance - req.CouponsUsed,
			Status:           "pending",
			Message:          "Transaction initiated. Waiting for employee approval.",
			RequiresApproval: true,
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
		//Get employeeUserID user from context
		employeeUserID, err := utils.GetUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()
		//Get transaction
		transactionCollection := database.OpenCollection("transactions", client)
		var transaction models.Transaction
		err = transactionCollection.FindOne(ctx, bson.D{{Key: "transaction_id", Value: transaction.TransactionID}}).Decode(&transaction)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}

		//varify this employee owns the transaction

		employeeCollection := database.OpenCollection("employees", client)
		var employee models.Employee
		err = employeeCollection.FindOne(ctx,
			bson.D{{Key: "user_id", Value: employeeUserID},
				{Key: "employee_id", Value: transaction.EmployeeID},
			}).Decode(&employee)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only approve own transaction"})
			return
		}
		// check if already processed
		if transaction.Status != "pending" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction already processed"})
			return
		}

		//process aproval or rejection
		// 4. Process Approval or Rejection
		if req.Approved {
			// APPROVED - Process transaction

			// Check balance again (might have changed)
			if employee.CurrentBalance < transaction.CouponsUsed {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
				return
			}

			// Deduct coupons using MongoDB transaction for atomicity
			session, err := client.StartSession()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start session"})
				return
			}
			defer session.EndSession(ctx)

			_, err = session.WithTransaction(ctx, func(sessCtx context.Context) (interface{}, error) {
				// Update employee balance
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

			c.JSON(http.StatusOK, gin.H{
				"message":          "Transaction approved and processed successfully",
				"transaction_id":   req.TransactionID,
				"coupons_deducted": transaction.CouponsUsed,
				"new_balance":      employee.CurrentBalance - transaction.CouponsUsed,
				"status":           "completed",
			})

		} else {
			// REJECTED - Cancel transaction
			now := time.Now()
			_, err = transactionCollection.UpdateOne(
				ctx,
				bson.D{{Key: "transaction_id", Value: req.TransactionID}},
				bson.D{{Key: "$set", Value: bson.D{
					{Key: "status", Value: "rejected"},
					{Key: "notes", Value: req.Reason},
					{Key: "updated_at", Value: now},
				}}},
			)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject transaction"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":        "Transaction rejected",
				"transaction_id": req.TransactionID,
				"status":         "rejected",
			})
		}
	}
}

func GetMyTransactions(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		employeeUserID, err := utils.GetUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		//Get employee
		employeeCollection := database.OpenCollection("employees", client)
		var employee models.Employee

		err = employeeCollection.FindOne(ctx, bson.D{{Key: "user_id", Value: employeeUserID}}).Decode(&employee)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}
		//Get transactions
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

		c.JSON(http.StatusOK, gin.H{
			"transactions": transactions,
			"total":        len(transactions),
		})
	}
}
