package models

import (
	"time"
	"go.mongodb.org/mongo-driver/v2/bson"
)
type Supplier struct {
	ID              bson.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	SupplierID      string        `json:"supplier_id" bson:"supplier_id"`
	UserID          string        `json:"user_id" bson:"user_id"` 
	BusinessName    string        `json:"business_name" bson:"business_name"`
	BusinessLicense string        `json:"business_license,omitempty" bson:"business_license,omitempty"`
	ContactPerson   string        `json:"contact_person" bson:"contact_person"`
	Phone           string        `json:"phone" bson:"phone"`
	Email           string        `json:"email" bson:"email"`
	Address         string        `json:"address" bson:"address"`
	Latitude        float64       `json:"latitude" bson:"latitude"`
	Longitude       float64       `json:"longitude" bson:"longitude"`
	LocationRadius  int           `json:"location_radius" bson:"location_radius"` 
	IsActive        bool          `json:"is_active" bson:"is_active"`
	IsVerified      bool          `json:"is_verified" bson:"is_verified"` 
	BankAccount     string        `json:"bank_account,omitempty" bson:"bank_account,omitempty"`
	TaxID           string        `json:"tax_id,omitempty" bson:"tax_id,omitempty"`
	Notes           string        `json:"notes,omitempty" bson:"notes,omitempty"` 
	CreatedByAdminID string       `json:"created_by_admin_id,omitempty" bson:"created_by_admin_id,omitempty"`
	CreatedAt       time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" bson:"updated_at"`
}
type CreateSupplierRequest struct {
	UserID          string  `json:"user_id" binding:"required"`
	BusinessName    string  `json:"business_name" binding:"required,min=2"`
	BusinessLicense string  `json:"business_license"`
	ContactPerson   string  `json:"contact_person" binding:"required"`
	Phone           string  `json:"phone" binding:"required"`
	Email           string  `json:"email" binding:"required,email"`
	Address         string  `json:"address" binding:"required"`
	Latitude        float64 `json:"latitude" binding:"required"`
	Longitude       float64 `json:"longitude" binding:"required"`
	LocationRadius  int     `json:"location_radius"` 
	BankAccount     string  `json:"bank_account"`
	TaxID           string  `json:"tax_id"`
	Notes           string  `json:"notes"`
}

type UpdateSupplierRequest struct {
	BusinessName    string  `json:"business_name"`
	BusinessLicense string  `json:"business_license"`
	ContactPerson   string  `json:"contact_person"`
	Phone           string  `json:"phone"`
	Address         string  `json:"address"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	LocationRadius  int     `json:"location_radius"`
	BankAccount     string  `json:"bank_account"`
	TaxID           string  `json:"tax_id"`
	Notes           string  `json:"notes"`
}

// SupplierResponse - Output for API responses
type SupplierResponse struct {
	SupplierID      string    `json:"supplier_id"`
	BusinessName    string    `json:"business_name"`
	ContactPerson   string    `json:"contact_person"`
	Phone           string    `json:"phone"`
	Email           string    `json:"email"`
	Address         string    `json:"address"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	LocationRadius  int       `json:"location_radius"`
	IsActive        bool      `json:"is_active"`
	IsVerified      bool      `json:"is_verified"`
	CreatedAt       time.Time `json:"created_at"`
}

// SupplierTotalsResponse - Supplier earnings/stats
type SupplierTotalsResponse struct {
	SupplierID         string  `json:"supplier_id"`
	BusinessName       string  `json:"business_name"`
	TotalTransactions  int     `json:"total_transactions"`
	TotalCoupons       int     `json:"total_coupons"`
	TotalAmount        float64 `json:"total_amount"`
	PendingTransactions int    `json:"pending_transactions"`
	CompletedToday     int     `json:"completed_today"`
	CompletedThisMonth int     `json:"completed_this_month"`
	EarningsToday      float64 `json:"earnings_today"`
	EarningsThisMonth  float64 `json:"earnings_this_month"`
}