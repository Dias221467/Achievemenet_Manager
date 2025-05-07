package repository

import (
	"context"
	"time"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"github.com/Dias221467/Achievemenet_Manager/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GoalRepository struct handles database operations related to goals
type GoalRepository struct {
	collection *mongo.Collection
}

// NewGoalRepository creates a new instance of GoalRepository
func NewGoalRepository(db *mongo.Database) *GoalRepository {
	return &GoalRepository{
		collection: db.Collection("goals"),
	}
}

// CreateGoal creates a new goal in the database
func (r *GoalRepository) CreateGoal(ctx context.Context, goal *models.Goal) (*models.Goal, error) {
	goal.CreatedAt = time.Now()
	goal.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, goal)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to insert goal")
		return nil, err
	}

	// Cast the inserted ID and assign it to the goal object
	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		logger.Log.Error("Failed to cast inserted ID")
		return nil, err
	}
	goal.ID = insertedID

	logger.Log.WithField("goal_id", goal.ID.Hex()).Info("Goal created successfully")
	return goal, nil
}

// GetGoalByID fetches a goal by its ID
func (r *GoalRepository) GetGoalByID(ctx context.Context, id primitive.ObjectID) (*models.Goal, error) {
	var goal models.Goal

	// Find the goal by its ID
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&goal)
	if err != nil {
		logger.Log.WithError(err).WithField("goal_id", id.Hex()).Error("Failed to find goal by ID")
		return nil, err
	}

	logger.Log.WithField("goal_id", id.Hex()).Info("Goal fetched successfully")
	return &goal, nil
}

// UpdateGoal updates an existing goal in the database
func (r *GoalRepository) UpdateGoal(ctx context.Context, id primitive.ObjectID, goal *models.Goal) (*models.Goal, error) {
	goal.UpdatedAt = time.Now()

	// Update the goal in the database
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": goal},
	)
	if err != nil {
		logger.Log.WithError(err).WithField("goal_id", id.Hex()).Error("Failed to update goal")
		return nil, err
	}

	logger.Log.WithField("goal_id", id.Hex()).Info("Goal updated successfully")
	return goal, nil
}

// DeleteGoal deletes a goal from the database by its ID
func (r *GoalRepository) DeleteGoal(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		logger.Log.WithError(err).WithField("goal_id", id.Hex()).Error("Failed to delete goal")
		return err
	}

	logger.Log.WithField("goal_id", id.Hex()).Info("Goal deleted successfully")
	return nil
}

// GetAllGoals fetches all goals from the database
func (r *GoalRepository) GetAllGoals(ctx context.Context, limit int64) ([]models.Goal, error) {
	var goals []models.Goal

	findOptions := options.Find().SetLimit(limit)
	cursor, err := r.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to fetch all goals")
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var goal models.Goal
		if err := cursor.Decode(&goal); err != nil {
			logger.Log.WithError(err).Error("Failed to decode goal")
			return nil, err
		}
		goals = append(goals, goal)
	}

	logger.Log.WithField("count", len(goals)).Info("All goals fetched successfully")
	return goals, nil
}

// GetGoals fetches goals for a specific user with an optional category filter
func (r *GoalRepository) GetGoals(ctx context.Context, userID primitive.ObjectID, category string) ([]models.Goal, error) {
	var goals []models.Goal

	// Build the filter for MongoDB query
	filter := bson.M{"user_id": userID}
	if category != "" {
		filter["category"] = category
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		logger.Log.WithError(err).WithField("user_id", userID.Hex()).Error("Failed to fetch filtered goals")
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var goal models.Goal
		if err := cursor.Decode(&goal); err != nil {
			logger.Log.WithError(err).Error("Failed to decode filtered goal")
			return nil, err
		}
		goals = append(goals, goal)
	}

	logger.Log.WithFields(map[string]interface{}{
		"user_id": userID.Hex(),
		"count":   len(goals),
	}).Info("Filtered goals fetched successfully")

	return goals, nil
}

// AddCollaborator adds a collaborator to a goal by updating the collaborators array.
func (r *GoalRepository) AddCollaborator(ctx context.Context, goalID, collaboratorID primitive.ObjectID) error {
	filter := bson.M{"_id": goalID}
	update := bson.M{
		"$addToSet": bson.M{"collaborators": collaboratorID}, // Prevents duplicates
		"$set":      bson.M{"updated_at": time.Now()},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Log.WithError(err).WithFields(map[string]interface{}{
			"goal_id":         goalID.Hex(),
			"collaborator_id": collaboratorID.Hex(),
		}).Error("Failed to add collaborator to goal")
		return err
	}

	logger.Log.WithFields(map[string]interface{}{
		"goal_id":         goalID.Hex(),
		"collaborator_id": collaboratorID.Hex(),
	}).Info("Collaborator successfully added to goal")

	return nil
}
