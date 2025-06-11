package repository

import (
	"context"
	"fmt"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ActivityRepository struct {
	collection *mongo.Collection
}

func NewActivityRepository(db *mongo.Database) *ActivityRepository {
	return &ActivityRepository{
		collection: db.Collection("activities"),
	}
}

// CreateActivity inserts a new activity log
func (r *ActivityRepository) CreateActivity(ctx context.Context, activity *models.Activity) error {
	_, err := r.collection.InsertOne(ctx, activity)
	if err != nil {
		logrus.WithError(err).Error("Failed to insert activity")
		return fmt.Errorf("failed to insert activity: %v", err)
	}
	return nil
}

// GetUserActivities fetches recent activities of a specific user
func (r *ActivityRepository) GetUserActivities(ctx context.Context, userID primitive.ObjectID, limit int) ([]models.Activity, error) {
	filter := bson.M{"user_id": userID}
	sort := bson.D{{Key: "timestamp", Value: -1}}

	opts := options.Find().SetSort(sort).SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch activities: %v", err)
	}
	defer cursor.Close(ctx)

	var activities []models.Activity
	if err := cursor.All(ctx, &activities); err != nil {
		return nil, fmt.Errorf("failed to decode activities: %v", err)
	}
	return activities, nil
}
