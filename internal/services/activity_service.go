package services

import (
	"context"
	"time"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"github.com/Dias221467/Achievemenet_Manager/internal/repository"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ActivityService struct {
	repo *repository.ActivityRepository
}

func NewActivityService(repo *repository.ActivityRepository) *ActivityService {
	return &ActivityService{repo: repo}
}

// LogActivity logs a user activity
func (s *ActivityService) LogActivity(
	ctx context.Context,
	userID primitive.ObjectID,
	actionType string,
	targetID primitive.ObjectID,
	message string,
) error {
	activity := &models.Activity{
		UserID:    userID,
		Type:      actionType,
		TargetID:  targetID,
		Message:   message,
		Timestamp: time.Now(),
	}

	err := s.repo.CreateActivity(ctx, activity)
	if err != nil {
		logrus.WithError(err).Error("Failed to log activity in service")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"user_id":     userID.Hex(),
		"action_type": actionType,
	}).Info("Activity logged successfully")

	return nil
}

// GetRecentActivities returns recent actions performed by a user
func (s *ActivityService) GetRecentActivities(ctx context.Context, userID primitive.ObjectID, limit int) ([]models.Activity, error) {
	return s.repo.GetUserActivities(ctx, userID, limit)
}
