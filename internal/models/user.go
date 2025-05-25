package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user account in the Achievement Manager system.
type User struct {
	ID             primitive.ObjectID   `bson:"_id,omitempty"`
	Friends        []primitive.ObjectID `json:"friends,omitempty" bson:"friends,omitempty"`
	Username       string               `bson:"username"`
	Email          string               `bson:"email"`
	HashedPassword string               `json:"hashed_password"`
	Role           string               `bson:"role" json:"role"`
	CreatedAt      time.Time            `bson:"created_at"`
	UpdatedAt      time.Time            `bson:"updated_at"`
}

type PublicUser struct {
	ID       primitive.ObjectID `json:"id"`
	Username string             `json:"username"`
	Email    string             `json:"email"`
}
