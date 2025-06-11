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
	IsVerified     bool                 `bson:"is_verified" json:"is_verified"`
	VerifyToken    string               `bson:"verify_token,omitempty" json:"-"`
	ResetToken     string               `bson:"reset_token,omitempty" json:"-"`
	ResetTokenExp  time.Time            `bson:"reset_token_exp,omitempty" json:"-"`
	CreatedAt      time.Time            `bson:"created_at"`
	UpdatedAt      time.Time            `bson:"updated_at"`
	LastActiveAt   time.Time            `bson:"last_active_at,omitempty" json:"last_active_at,omitempty"`
}

type PublicUser struct {
	ID       primitive.ObjectID `json:"id"`
	Username string             `json:"username"`
	Email    string             `json:"email"`
}
