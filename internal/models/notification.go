package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	ID        primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID  `bson:"user_id" json:"user_id"`
	Type      string              `bson:"type" json:"type"`                               // e.g. "goal_completed", "substep_due"
	Title     string              `bson:"title" json:"title"`                             // Short headline
	Message   string              `bson:"message" json:"message"`                         // Descriptive content
	Read      bool                `bson:"read" json:"read"`                               // True if user viewed it
	TargetID  *primitive.ObjectID `bson:"target_id,omitempty" json:"target_id,omitempty"` // Optional reference to goal/wish/etc.
	CreatedAt time.Time           `bson:"created_at" json:"created_at"`
	ExpiresAt time.Time           `bson:"expires_at" json:"expires_at"` // For auto-deletion after 7 days
}
