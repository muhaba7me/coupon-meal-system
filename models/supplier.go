package models

import (
	"time"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Supplier struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       bson.ObjectID `bson:"user_id" json:"user_id"`
	BusinessName   string             `bson:"business_name" json:"business_name" binding:"required"`
	BusinessLicense string            `bson:"business_license,omitempty" json:"business_license,omitempty"`
	Address        string             `bson:"address" json:"address" binding:"required"`
	Latitude       float64            `bson:"latitude" json:"latitude" binding:"required"`
	Longitude      float64            `bson:"longitude" json:"longitude" binding:"required"`
	LocationRadius int                `bson:"location_radius" json:"location_radius"` 
	IsActive       bool               `bson:"is_active" json:"is_active"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}