package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Activity struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Type      string             `bson:"type" json:"type"`           // e.g. "goal_created", "wish_updated"
	TargetID  primitive.ObjectID `bson:"target_id" json:"target_id"` // the ID of the goal, wish, etc.
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
	Message   string             `bson:"message" json:"message"`
}
