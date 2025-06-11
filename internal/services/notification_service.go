package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"github.com/Dias221467/Achievemenet_Manager/internal/repository"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationService struct {
	repo     *repository.NotificationRepository
	userRepo *repository.UserRepository
	goalRepo *repository.GoalRepository
}

func NewNotificationService(repo *repository.NotificationRepository, userrepo *repository.UserRepository, goalrepo *repository.GoalRepository) *NotificationService {
	return &NotificationService{
		repo:     repo,
		userRepo: userrepo,
		goalRepo: goalrepo,
	}
}

// CreateNotification logs a new notification for a user
func (s *NotificationService) CreateNotification(ctx context.Context, userID primitive.ObjectID, notifType, title, message string, targetID *primitive.ObjectID) error {
	notif := &models.Notification{
		UserID:   userID,
		Type:     notifType,
		Title:    title,
		Message:  message,
		Read:     false,
		TargetID: targetID,
	}
	return s.repo.CreateNotification(ctx, notif)
}

// GetUserNotifications returns all notifications for a user
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID primitive.ObjectID) ([]models.Notification, error) {
	return s.repo.GetUserNotifications(ctx, userID)
}

// MarkNotificationAsRead sets the "read" status of a notification to true
func (s *NotificationService) MarkNotificationAsRead(ctx context.Context, notifID primitive.ObjectID) error {
	return s.repo.MarkAsRead(ctx, notifID)
}

// DeleteNotification deletes a specific notification
func (s *NotificationService) DeleteNotification(ctx context.Context, notifID primitive.ObjectID) error {
	return s.repo.DeleteNotification(ctx, notifID)
}

// CleanupExpiredNotifications could be called periodically (e.g. by cron) to delete old ones
func (s *NotificationService) CleanupExpiredNotifications(ctx context.Context) error {
	// Optional to implement later
	return nil
}

func (s *NotificationService) CheckInactiveUsers(ctx context.Context) error {
	users, err := s.userRepo.GetAllUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	now := time.Now()
	for _, user := range users {
		if user.LastActiveAt.IsZero() || now.Sub(user.LastActiveAt) >= 3*24*time.Hour {
			// Check if they already got a recent inactivity notification
			existing, err := s.repo.GetLatestNotificationByType(ctx, user.ID, "user_inactive")
			if err == nil && existing != nil && now.Sub(existing.CreatedAt) < 3*24*time.Hour {
				continue // skip duplicate notification
			}

			err = s.CreateNotification(ctx, user.ID, "user_inactive",
				"We miss you!",
				"You haven't been active for a few days. Come back and make progress on your goals!",
				nil,
			)
			if err != nil {
				logrus.WithError(err).Warnf("Failed to send inactivity notification to user %s", user.ID.Hex())
			}
		}
	}

	return nil
}

func (s *NotificationService) DeleteExpiredNotifications(ctx context.Context) error {
	return s.repo.DeleteExpiredNotifications(ctx)
}

func (s *NotificationService) CheckGoalDueSoon(ctx context.Context) error {
	goals, err := s.goalRepo.GetAllGoals(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to fetch goals: %w", err)
	}

	now := time.Now()
	for _, goal := range goals {
		// Пропустить уже завершённые цели или те, у кого нет дедлайна
		if goal.Status == "completed" || goal.DueDate.IsZero() {
			continue
		}

		// В пределах следующих 24 часов?
		timeLeft := goal.DueDate.Sub(now)
		if timeLeft > 0 && timeLeft <= 24*time.Hour {
			// Проверим, уже ли есть похожее уведомление
			existing, err := s.repo.GetLatestNotificationByType(ctx, goal.UserID, "goal_due_soon")
			if err == nil && existing != nil && existing.TargetID != nil && *existing.TargetID == goal.ID {
				continue // уже есть активное уведомление
			}

			message := fmt.Sprintf("Goal \"%s\" is due soon! Don't forget to complete it.", goal.Name)
			err = s.CreateNotification(ctx, goal.UserID, "goal_due_soon", "⏰ Goal Due Soon", message, &goal.ID)
			if err != nil {
				logrus.WithError(err).Warnf("Failed to send goal due soon notification for goal %s", goal.ID.Hex())
			}
		}
	}

	return nil
}

func (s *NotificationService) CheckStepDueSoon(ctx context.Context) error {
	goals, err := s.goalRepo.GetAllGoals(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to fetch goals: %w", err)
	}

	now := time.Now()
	for _, goal := range goals {
		// Пропускаем завершённые цели
		if goal.Status == "completed" {
			continue
		}

		for _, step := range goal.Steps {
			if step.Completed || step.DueDate.IsZero() {
				continue
			}

			timeLeft := step.DueDate.Sub(now)
			if timeLeft > 0 && timeLeft <= 24*time.Hour {
				// Проверим, есть ли уже уведомление
				existing, err := s.repo.GetLatestNotificationByType(ctx, goal.UserID, "step_due_soon")
				if err == nil && existing != nil && existing.Title == step.Name && existing.TargetID != nil && *existing.TargetID == goal.ID {
					continue // уведомление уже есть
				}

				message := fmt.Sprintf("Step \"%s\" of goal \"%s\" is due soon!", step.Name, goal.Name)
				err = s.CreateNotification(ctx, goal.UserID, "step_due_soon", step.Name, message, &goal.ID)
				if err != nil {
					logrus.WithError(err).Warnf("Failed to send step due soon notification for goal %s", goal.ID.Hex())
				}
			}
		}
	}

	return nil
}

func (s *NotificationService) CheckSubstepDueSoon(ctx context.Context) error {
	goals, err := s.goalRepo.GetAllGoals(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to fetch goals: %w", err)
	}

	now := time.Now()
	for _, goal := range goals {
		for _, step := range goal.Steps {
			for i, sub := range step.Substeps {
				if sub.Done || sub.DueDate.IsZero() {
					continue
				}
				if sub.DueDate.After(now) && sub.DueDate.Before(now.Add(24*time.Hour)) {
					// Create unique key per substep (avoid spam)
					key := fmt.Sprintf("substep_due_%s_%d", goal.ID.Hex(), i)
					existing, _ := s.repo.GetLatestNotificationByType(ctx, goal.UserID, key)
					if existing != nil && now.Sub(existing.CreatedAt) < 12*time.Hour {
						continue
					}

					// Send notification
					err := s.CreateNotification(
						ctx,
						goal.UserID,
						key, // Use unique type to avoid repeats
						"📌 Substep Deadline Approaching",
						fmt.Sprintf("Your substep '%s' in step '%s' of goal '%s' is due soon!", sub.Title, step.Name, goal.Name),
						&goal.ID,
					)
					if err != nil {
						logrus.WithError(err).Warn("Failed to send substep due notification")
					}
				}
			}
		}
	}

	return nil
}
