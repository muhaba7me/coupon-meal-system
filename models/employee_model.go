package models

import (
	"time"
	"go.mongodb.org/mongo-driver/v2/bson"
)
type Employee struct {
    ID                    bson.ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
    EmployeeID            string         `json:"employee_id" bson:"employee_id"`
    UserID                string         `json:"user_id" bson:"user_id"`
    EmployeeCode          string         `json:"employee_code" bson:"employee_code"`
    Name                  string         `json:"name" bson:"name"`
    Email                 string         `json:"email" bson:"email"`
    Phone                 string         `json:"phone,omitempty" bson:"phone,omitempty"`
    Status                string         `json:"status" bson:"status"`
    MonthlyAllocation     int            `json:"monthly_coupon_allocation" bson:"monthly_coupon_allocation"`
    CurrentBalance        int            `json:"current_coupon_balance" bson:"current_coupon_balance"`
    LastAllocationDate    *time.Time     `json:"last_allocation_date,omitempty" bson:"last_allocation_date,omitempty"`
    HireDate              time.Time      `json:"hire_date" bson:"hire_date"`
    TerminationDate       *time.Time     `json:"termination_date,omitempty" bson:"termination_date,omitempty"`
    CreatedByAdminID      string         `json:"created_by_admin_id,omitempty" bson:"created_by_admin_id,omitempty"`
    LastLogin             *time.Time     `json:"last_login,omitempty" bson:"last_login,omitempty"`
    IsVerified            bool           `json:"is_verified" bson:"is_verified"`
    Notes                 string         `json:"notes,omitempty" bson:"notes,omitempty"`
    CreatedAt             time.Time      `json:"created_at" bson:"created_at"`
    UpdatedAt             time.Time      `json:"updated_at" bson:"updated_at"`
}

type CreateEmployeeRequest struct {
    UserID       string `json:"user_id" binding:"required"`
    EmployeeCode string `json:"employee_code" binding:"required"`
    Name         string `json:"name" binding:"required,min=2"`
    Email        string `json:"email" binding:"required,email"`
    Phone        string `json:"phone"`
    HireDate     string `json:"hire_date" binding:"required"` 
    Notes        string `json:"notes"`
}

type UpdateEmployeeRequest struct {
    Name   string `json:"name"`
    Phone  string `json:"phone"`
    Status string `json:"status"` 
    Notes  string `json:"notes"`
}
type EmployeeResponse struct {
    EmployeeID         string     `json:"employee_id"`
    EmployeeCode       string     `json:"employee_code"`
    Name               string     `json:"name"`
    Email              string     `json:"email"`
    Phone              string     `json:"phone"`
    Status             string     `json:"status"`
    MonthlyAllocation  int        `json:"monthly_coupon_allocation"`
    CurrentBalance     int        `json:"current_coupon_balance"`
    LastAllocationDate *time.Time `json:"last_allocation_date,omitempty"`
    HireDate           time.Time  `json:"hire_date"`
    IsVerified         bool       `json:"is_verified"`
    CreatedAt          time.Time  `json:"created_at"`
}