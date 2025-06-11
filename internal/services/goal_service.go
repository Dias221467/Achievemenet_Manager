package services

import (
	"context"
	"fmt"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"github.com/Dias221467/Achievemenet_Manager/internal/repository"
	"github.com/Dias221467/Achievemenet_Manager/pkg/logger"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GoalService encapsulates the business logic for goals.
type GoalService struct {
	repo                *repository.GoalRepository
	userRepo            *repository.UserRepository
	NotificationService *NotificationService
}

// NewGoalService creates a new instance of GoalService.
func NewGoalService(repo *repository.GoalRepository, userRepo *repository.UserRepository, notificationService *NotificationService) *GoalService {
	return &GoalService{
		repo:                repo,
		userRepo:            userRepo,
		NotificationService: notificationService,
	}
}

// CreateGoal processes the goal creation logic and stores it in the database.
func (s *GoalService) CreateGoal(ctx context.Context, goal *models.Goal) (*models.Goal, error) {
	if goal.Name == "" {
		logger.Log.Warn("Goal name is empty during creation")
		return nil, fmt.Errorf("goal name is required")
	}

	createdGoal, err := s.repo.CreateGoal(ctx, goal)
	if err != nil {
		logger.Log.WithError(err).Error("Service failed to create goal")
		return nil, fmt.Errorf("failed to create goal: %v", err)
	}

	logger.Log.WithField("goal_id", createdGoal.ID.Hex()).Info("Goal created in service layer")
	return createdGoal, nil
}

// GetGoal retrieves a goal by its ID.
func (s *GoalService) GetGoal(ctx context.Context, id string) (*models.Goal, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.Log.WithField("goal_id", id).WithError(err).Warn("Invalid goal ID in GetGoal")
		return nil, fmt.Errorf("invalid goal ID: %v", err)
	}

	goal, err := s.repo.GetGoalByID(ctx, objID)
	if err != nil {
		logger.Log.WithField("goal_id", id).WithError(err).Error("Failed to get goal from repository")
		return nil, fmt.Errorf("failed to get goal: %v", err)
	}

	logger.Log.WithField("goal_id", id).Info("Goal retrieved successfully in service layer")
	return goal, nil
}

// UpdateGoal updates an existing goal.
func (s *GoalService) UpdateGoal(ctx context.Context, id string, updatedGoal *models.Goal) (*models.Goal, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.Log.WithField("goal_id", id).WithError(err).Warn("Invalid goal ID in UpdateGoal")
		return nil, fmt.Errorf("invalid goal ID: %v", err)
	}

	goal, err := s.repo.UpdateGoal(ctx, objID, updatedGoal)
	if err != nil {
		logger.Log.WithField("goal_id", id).WithError(err).Error("Failed to update goal")
		return nil, fmt.Errorf("failed to update goal: %v", err)
	}

	if goal.Status == "completed" {
		go func() {
			err := s.NotificationService.CreateNotification(
				ctx,
				goal.UserID,
				"goal_completed",
				"🎉 Goal Completed",
				fmt.Sprintf("You’ve successfully completed your goal: \"%s\"!", goal.Name),
				&goal.ID,
			)
			if err != nil {
				logrus.WithError(err).Warn("Failed to send goal completed notification")
			}
		}()
	}

	logger.Log.WithField("goal_id", id).Info("Goal updated successfully in service layer")
	return goal, nil
}

// DeleteGoal removes a goal from the database.
func (s *GoalService) DeleteGoal(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.Log.WithField("goal_id", id).WithError(err).Warn("Invalid goal ID in DeleteGoal")
		return fmt.Errorf("invalid goal ID: %v", err)
	}

	if err := s.repo.DeleteGoal(ctx, objID); err != nil {
		logger.Log.WithField("goal_id", id).WithError(err).Error("Failed to delete goal")
		return fmt.Errorf("failed to delete goal: %v", err)
	}

	logger.Log.WithField("goal_id", id).Info("Goal deleted successfully in service layer")
	return nil
}

// GetAllGoals retrieves a list of goals with an optional limit.
func (s *GoalService) GetAllGoals(ctx context.Context, limit int64) ([]models.Goal, error) {
	goals, err := s.repo.GetAllGoals(ctx, limit)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to fetch all goals")
		return nil, fmt.Errorf("failed to fetch goals: %v", err)
	}

	logger.Log.WithField("count", len(goals)).Info("All goals fetched successfully in service layer")
	return goals, nil
}

func (s *GoalService) GetGoals(ctx context.Context, userID primitive.ObjectID, category string) ([]models.Goal, error) {
	goals, err := s.repo.GetGoals(ctx, userID, category)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"user_id":  userID.Hex(),
			"category": category,
		}).WithError(err).Error("Failed to get filtered goals in service")
		return nil, err
	}

	logger.Log.WithFields(map[string]interface{}{
		"user_id":  userID.Hex(),
		"category": category,
		"count":    len(goals),
	}).Info("Filtered goals fetched in service layer")
	return goals, nil
}

// InviteCollaborator adds a user as a collaborator to a goal if the requester is the owner.
func (s *GoalService) InviteCollaborator(ctx context.Context, goalID string, requesterID, collaboratorID primitive.ObjectID) error {
	objID, err := primitive.ObjectIDFromHex(goalID)
	if err != nil {
		return fmt.Errorf("invalid goal ID: %v", err)
	}

	goal, err := s.repo.GetGoalByID(ctx, objID)
	if err != nil {
		return fmt.Errorf("goal not found: %v", err)
	}

	// Only the owner can invite collaborators
	if goal.UserID != requesterID {
		return fmt.Errorf("only the owner can invite collaborators")
	}

	// Prevent inviting self or duplicate
	if collaboratorID == requesterID {
		return fmt.Errorf("you cannot invite yourself")
	}
	for _, existing := range goal.Collaborators {
		if existing == collaboratorID {
			return fmt.Errorf("user is already a collaborator")
		}
	}

	//Check if they are friends (important!)
	friendIDs, err := s.userRepo.GetFriendIDs(ctx, requesterID)
	if err != nil {
		return fmt.Errorf("failed to fetch friend list: %v", err)
	}

	isFriend := false
	for _, id := range friendIDs {
		if id == collaboratorID {
			isFriend = true
			break
		}
	}
	if !isFriend {
		return fmt.Errorf("you can only invite your friends")
	}

	return s.repo.AddCollaborator(ctx, objID, collaboratorID)
}
