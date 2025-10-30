package controllers

import (
	"context"
	"fmt"
	"net/http"
		"encoding/base64"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/muhaba7me/coupon-meal-system/database"
	"github.com/muhaba7me/coupon-meal-system/models"
	"github.com/muhaba7me/coupon-meal-system/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	qrcode "github.com/skip2/go-qrcode"
)

func GenerateQrCode(client *mongo.Client) gin.HandlerFunc{
	return func(c *gin.Context){
		//get employee ID from JWT token 

		employeeUserID, err := utils.GetUserIdFromContext(c)
		if err !=nil{
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Unauthorized"})
			return 
		}
		var ctx, cancel = context.WithTimeout(c, 100*time.Second)

		defer cancel()
		//get employee details 
		employeeCollection := database.OpenCollection("employees", client)
		var employee models.Employee
			err = employeeCollection.FindOne(ctx, bson.D{{Key: "user_id", Value: employeeUserID}}).Decode(&employee)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}
		// Check if employee can use coupons
		if employee.Status != "active" && employee.Status != "on_leave" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Employee account is not active"})
			return
		}
		if employee.CurrentBalance <= 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "No coupons available"})
			return
		}
			// Generate unique QR code
		qrCodeUUID := uuid.New().String()
		
		// Get expiry minutes from env or default to 15
		expiryMinutes := 15
		if expiryStr := os.Getenv("QR_EXPIRY_MINUTES"); expiryStr != "" {
			if minutes, err := strconv.Atoi(expiryStr); err == nil {
				expiryMinutes = minutes
			}
		}
		expiresAt := time.Now().Add(time.Duration(expiryMinutes) * time.Minute)
		// Create QR code record
		qrCodeRecord := models.QRCode{
			QRCodeID:   bson.NewObjectID().Hex(),
			Code:       qrCodeUUID,
			EmployeeID: employee.EmployeeID,
			ExpiresAt:  expiresAt,
			IsUsed:     false,
			CreatedAt:  time.Now(),
		}

		
	// Save to database
		qrCollection := database.OpenCollection("qr_codes", client)
		_, err = qrCollection.InsertOne(ctx, qrCodeRecord)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate QR code"})
			return
		}

		// Generate QR code image (base64)
		qrCodeData := fmt.Sprintf("COUPON-%s-%s", employee.EmployeeCode, qrCodeUUID)
		qrImage, err := qrcode.Encode(qrCodeData, qrcode.Medium, 256)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate QR image"})
			return
		
	}
	// Convert to base64
		base64Image := base64.StdEncoding.EncodeToString(qrImage)

		// Return response
		c.JSON(http.StatusOK, models.QRCodeResponse{
			QRCodeID:         qrCodeRecord.QRCodeID,
			Code:             qrCodeRecord.Code,
			QRCodeImage:      "data:image/png;base64," + base64Image,
			ExpiresAt:        expiresAt,
			ExpiresInMinutes: expiryMinutes,
			EmployeeBalance:  employee.CurrentBalance,
		})
}
}