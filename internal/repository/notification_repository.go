package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NotificationRepository struct {
	collection *mongo.Collection
}

func NewNotificationRepository(db *mongo.Database) *NotificationRepository {
	return &NotificationRepository{
		collection: db.Collection("notifications"),
	}
}

// CreateNotification inserts a new notification
func (r *NotificationRepository) CreateNotification(ctx context.Context, notif *models.Notification) error {
	notif.CreatedAt = time.Now()
	notif.ExpiresAt = notif.CreatedAt.Add(7 * 24 * time.Hour)

	_, err := r.collection.InsertOne(ctx, notif)
	if err != nil {
		logrus.WithError(err).Error("Failed to insert notification")
		return fmt.Errorf("failed to create notification: %v", err)
	}
	return nil
}

// GetUserNotifications returns all notifications for a user
func (r *NotificationRepository) GetUserNotifications(ctx context.Context, userID primitive.ObjectID) ([]models.Notification, error) {
	filter := bson.M{
		"user_id":    userID,
		"expires_at": bson.M{"$gt": time.Now()},
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch notifications: %v", err)
	}
	defer cursor.Close(ctx)

	var notifications []models.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, fmt.Errorf("failed to decode notifications: %v", err)
	}
	return notifications, nil
}

// MarkAsRead sets notification's Read to true
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"read": true}})
	return err
}

// DeleteNotification deletes a notification
func (r *NotificationRepository) DeleteNotification(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *NotificationRepository) GetLatestNotificationByType(ctx context.Context, userID primitive.ObjectID, notifType string) (*models.Notification, error) {
	filter := bson.M{
		"user_id": userID,
		"type":    notifType,
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})

	var notif models.Notification
	err := r.collection.FindOne(ctx, filter, opts).Decode(&notif)
	if err != nil {
		return nil, err
	}
	return &notif, nil
}

// DeleteExpiredNotifications удаляет уведомления, у которых истёк срок
func (r *NotificationRepository) DeleteExpiredNotifications(ctx context.Context) error {
	filter := bson.M{"expires_at": bson.M{"$lte": time.Now()}}
	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete expired notifications: %v", err)
	}
	logrus.Infof("Deleted %d expired notifications", result.DeletedCount)
	return nil
}
