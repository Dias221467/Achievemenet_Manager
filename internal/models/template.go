package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GoalTemplate struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description" bson:"description"`
	Steps       []string           `json:"steps" bson:"steps"`
	Category    string             `json:"category,omitempty" bson:"category,omitempty"`
	UserID      primitive.ObjectID `json:"user_id" bson:"user_id"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
}
