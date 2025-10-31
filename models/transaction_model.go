package models

import (
	"time"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Transaction struct {
	ID               bson.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	TransactionID    string        `json:"transaction_id" bson:"transaction_id"`
	EmployeeID       string        `json:"employee_id" bson:"employee_id"`
	SupplierID       string        `json:"supplier_id" bson:"supplier_id"`
	QRCodeID         string        `json:"qr_code_id" bson:"qr_code_id"`
	CouponsUsed      int           `json:"coupons_used" bson:"coupons_used"` // 1-3
	TotalAmount      float64       `json:"total_amount" bson:"total_amount"` // CouponsUsed × 45
	EmployeeLatitude  float64      `json:"employee_latitude,omitempty" bson:"employee_latitude,omitempty"`
	EmployeeLongitude float64      `json:"employee_longitude,omitempty" bson:"employee_longitude,omitempty"`
	Status           string        `json:"status" bson:"status"`
	Notes            string        `json:"notes,omitempty" bson:"notes,omitempty"`
	ProcessedAt      time.Time     `json:"processed_at" bson:"processed_at"`
	CreatedAt        time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at" bson:"updated_at"`
}

// Request models
type InitiateTransactionRequest struct {
	QRCode      string  `json:"qr_code" binding:"required"`
	CouponsUsed int     `json:"coupons_used" binding:"required,min=1,max=3"`
	SupplierID  string  `json:"supplier_id" binding:"required"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	Notes       string  `json:"notes,omitempty"`
}

type ApproveTransactionRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
	Approved      bool   `json:"approved" binding:"required"`
	Reason        string `json:"reason,omitempty"` 
}


type TransactionResponse struct {
	TransactionID    string    `json:"transaction_id"`
	EmployeeName     string    `json:"employee_name"`
	EmployeeCode     string    `json:"employee_code"`
	SupplierName     string    `json:"supplier_name"`
	CouponsUsed      int       `json:"coupons_used"`
	TotalAmount      float64   `json:"total_amount"`
	Status           string    `json:"status"`
	ProcessedAt      time.Time `json:"processed_at"`
	CreatedAt        time.Time `json:"created_at"`
}


type TransactionInitiatedResponse struct {
	TransactionID      string  `json:"transaction_id"`
	EmployeeName       string  `json:"employee_name"`
	EmployeeCode       string  `json:"employee_code"`
	CurrentBalance     int     `json:"current_balance"`
	CouponsToDeduct    int     `json:"coupons_to_deduct"`
	TotalAmount        float64 `json:"total_amount"`
	NewBalance         int     `json:"new_balance"`
	Status             string  `json:"status"`
	Message            string  `json:"message"`
	RequiresApproval   bool    `json:"requires_approval"`
}