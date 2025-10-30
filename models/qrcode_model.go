package models

import (
	"time"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type QRCode struct {
	ID         bson.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	QRCodeID   string        `json:"qr_code_id" bson:"qr_code_id"`
	Code       string        `json:"code" bson:"code"` // UUID string
	EmployeeID string        `json:"employee_id" bson:"employee_id"`
	ExpiresAt  time.Time     `json:"expires_at" bson:"expires_at"`
	IsUsed     bool          `json:"is_used" bson:"is_used"`
	UsedAt     *time.Time    `json:"used_at,omitempty" bson:"used_at,omitempty"`
	CreatedAt  time.Time     `json:"created_at" bson:"created_at"`
}

type QRCodeResponse struct {
	QRCodeID       string    `json:"qr_code_id"`
	Code           string    `json:"code"`
	QRCodeImage    string    `json:"qr_code_image"` // Base64 encoded
	ExpiresAt      time.Time `json:"expires_at"`
	ExpiresInMinutes int     `json:"expires_in_minutes"`
	EmployeeBalance int      `json:"employee_balance"`
}

type ValidateQRRequest struct {
	Code string `json:"code" binding:"required"`
}

type ValidateQRResponse struct {
	Valid           bool      `json:"valid"`
	EmployeeID      string    `json:"employee_id,omitempty"`
	EmployeeName    string    `json:"employee_name,omitempty"`
	EmployeeCode    string    `json:"employee_code,omitempty"`
	CurrentBalance  int       `json:"current_balance,omitempty"`
	QRCodeID        string    `json:"qr_code_id,omitempty"`
	ExpiresAt       time.Time `json:"expires_at,omitempty"`
	Message         string    `json:"message,omitempty"`
}