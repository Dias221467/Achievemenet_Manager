package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TemplateRepository struct {
	collection *mongo.Collection
}

func NewTemplateRepository(db *mongo.Database) *TemplateRepository {
	return &TemplateRepository{
		collection: db.Collection("templates"),
	}
}

func (r *TemplateRepository) CreateTemplate(ctx context.Context, template *models.GoalTemplate) (*models.GoalTemplate, error) {
	template.CreatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, template)
	if err != nil {
		return nil, fmt.Errorf("failed to insert template: %v", err)
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("failed to cast inserted ID")
	}
	template.ID = insertedID

	return template, nil
}

func (r *TemplateRepository) GetAllTemplates(ctx context.Context) ([]models.GoalTemplate, error) {
	var templates []models.GoalTemplate

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch templates: %v", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var template models.GoalTemplate
		if err := cursor.Decode(&template); err != nil {
			return nil, fmt.Errorf("failed to decode template: %v", err)
		}
		templates = append(templates, template)
	}

	return templates, nil
}

func (r *TemplateRepository) GetTemplateByID(ctx context.Context, id primitive.ObjectID) (*models.GoalTemplate, error) {
	var template models.GoalTemplate

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&template)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch template by id: %v", err)
	}

	return &template, nil
}

// GetTemplatesByUser fetches templates created by a specific user.
func (r *TemplateRepository) GetTemplatesByUser(ctx context.Context, userID primitive.ObjectID) ([]models.GoalTemplate, error) {
	var templates []models.GoalTemplate

	// Filter templates by user_id
	filter := bson.M{"user_id": userID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch templates: %v", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var template models.GoalTemplate
		if err := cursor.Decode(&template); err != nil {
			return nil, fmt.Errorf("failed to decode template: %v", err)
		}
		templates = append(templates, template)
	}

	return templates, nil
}

// GetPublicTemplates returns all public templates
func (r *TemplateRepository) GetPublicTemplates(ctx context.Context) ([]models.GoalTemplate, error) {
	var templates []models.GoalTemplate

	filter := bson.M{"public": true}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch public templates: %v", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var template models.GoalTemplate
		if err := cursor.Decode(&template); err != nil {
			return nil, fmt.Errorf("failed to decode template: %v", err)
		}
		templates = append(templates, template)
	}

	return templates, nil
}

// GetPublicTemplatesByUser fetches public templates created by a specific user.
func (r *TemplateRepository) GetPublicTemplatesByUser(ctx context.Context, userID primitive.ObjectID) ([]models.GoalTemplate, error) {
	var templates []models.GoalTemplate
	filter := bson.M{
		"user_id": userID,
		"public":  true,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch public templates for user: %v", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var template models.GoalTemplate
		if err := cursor.Decode(&template); err != nil {
			return nil, fmt.Errorf("failed to decode template: %v", err)
		}
		templates = append(templates, template)
	}

	return templates, nil
}
