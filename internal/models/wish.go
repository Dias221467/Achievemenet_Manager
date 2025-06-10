package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Wish struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Images      []string           `bson:"images,omitempty" json:"images,omitempty"` // ← updated
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}
