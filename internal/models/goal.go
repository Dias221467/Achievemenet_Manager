package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Predefined categories (optional, for validation)
var AllowedCategories = map[string]bool{
	"Health":        true,
	"Career":        true,
	"Education":     true,
	"Personal":      true,
	"Finance":       true,
	"Hobby":         true,
	"Relationships": true,
}

// Goal represents a user's goal.
type Goal struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	UserID        primitive.ObjectID   `bson:"user_id" json:"user_id"`
	Name          string               `bson:"name" json:"name"`
	Description   string               `bson:"description" json:"description"`
	Category      string               `bson:"category,omitempty" json:"category,omitempty"` // New Field
	Steps         []Step               `bson:"steps" json:"steps"`
	Status        string               `bson:"status" json:"status"`
	DueDate       time.Time            `bson:"due_date,omitempty" json:"due_date,omitempty"`
	Collaborators []primitive.ObjectID `bson:"collaborators,omitempty" json:"collaborators,omitempty"`
	CreatedAt     time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time            `bson:"updated_at" json:"updated_at"`
}

type Step struct {
	Name      string    `bson:"name" json:"name"`
	DueDate   time.Time `bson:"due_date,omitempty" json:"due_date,omitempty"`
	Substeps  []Substep `bson:"substeps" json:"substeps"`
	Completed bool      `bson:"completed" json:"completed"`
}

type Substep struct {
	Title   string    `bson:"title" json:"title"`
	DueDate time.Time `bson:"due_date,omitempty" json:"due_date,omitempty"`
	Done    bool      `bson:"done" json:"done"`
}
