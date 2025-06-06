package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GoalTemplate struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description" bson:"description"`
	Steps       []TemplateStep     `json:"steps" bson:"steps"`
	Category    string             `json:"category,omitempty" bson:"category,omitempty"`
	UserID      primitive.ObjectID `json:"user_id" bson:"user_id"`
	Public      bool               `json:"public" bson:"public"` // New: indicates if template is public
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
}

// For use inside templates
type TemplateStep struct {
	Name     string            `bson:"name" json:"name"`
	Substeps []TemplateSubstep `bson:"substeps" json:"substeps"`
}

type TemplateSubstep struct {
	Title string `bson:"title" json:"title"`
}
