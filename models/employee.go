package models

import (
	"time"


	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	StatusActive     = "active"
	StatusInactive   = "inactive"
	StatusTerminated = "terminated"
)

type Employee struct {
	ID                 primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID             primitive.ObjectID  `bson:"user_id" json:"user_id" binding:"required"`
	EmployeeCode       string              `bson:"employee_code" json:"employee_code" binding:"required"`
	Name               string              `bson:"name" json:"name" binding:"required"`
	Email              string              `bson:"email" json:"email" binding:"required,email"`
	Phone              string              `bson:"phone,omitempty" json:"phone,omitempty"`
	Status             string              `bson:"status,omitempty" json:"status,omitempty"`
	MonthlyAllocation  int                 `bson:"monthly_coupon_allocation" json:"monthly_coupon_allocation"`
	CurrentBalance     int                 `bson:"current_coupon_balance" json:"current_coupon_balance"`
	LastAllocationDate *time.Time          `bson:"last_allocation_date,omitempty" json:"last_allocation_date,omitempty"`
	HireDate           time.Time           `bson:"hire_date" json:"hire_date"`
	TerminationDate    *time.Time          `bson:"termination_date,omitempty" json:"termination_date,omitempty"`
	CreatedByAdminID   *primitive.ObjectID `bson:"created_by_admin_id,omitempty" json:"created_by_admin_id,omitempty"`
	LastLogin          *time.Time          `bson:"last_login,omitempty" json:"last_login,omitempty"`
	IsVerified         bool                `bson:"is_verified" json:"is_verified"`
	Notes              string              `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt          time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time           `bson:"updated_at" json:"updated_at"`
}
