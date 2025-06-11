package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/Dias221467/Achievemenet_Manager/internal/services"
	"github.com/sirupsen/logrus"
)

type DeadlineNotifier struct {
	GoalService         *services.GoalService
	NotificationService *services.NotificationService
}

// NewDeadlineNotifier creates a new instance of DeadlineNotifier
func NewDeadlineNotifier(goalService *services.GoalService, notifService *services.NotificationService) *DeadlineNotifier {
	return &DeadlineNotifier{
		GoalService:         goalService,
		NotificationService: notifService,
	}
}

// RunDailyScan checks for goals, steps and suvsteps due in next 24h and sends reminders
func (d *DeadlineNotifier) RunDailyScan(ctx context.Context) error {
	goals, err := d.GoalService.GetAllGoals(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to fetch goals: %v", err)
	}

	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)

	for _, goal := range goals {
		//  Goal due soon
		if goal.Status != "completed" && goal.DueDate.After(now) && goal.DueDate.Before(tomorrow) {
			_ = d.NotificationService.CreateNotification(
				ctx,
				goal.UserID,
				"goal_due_soon",
				"Goal Due Soon",
				fmt.Sprintf("Your goal \"%s\" is due by %s.", goal.Name, goal.DueDate.Format("Jan 2")),
				&goal.ID,
			)
		}

		for _, step := range goal.Steps {
			//  Step due soon
			if !step.Completed && step.DueDate.After(now) && step.DueDate.Before(tomorrow) {
				_ = d.NotificationService.CreateNotification(
					ctx,
					goal.UserID,
					"step_due_soon",
					"Step Due Soon",
					fmt.Sprintf("Step \"%s\" in goal \"%s\" is due soon.", step.Name, goal.Name),
					&goal.ID,
				)
			}

			for _, substep := range step.Substeps {
				//  Substep due soon
				if !substep.Done && substep.DueDate.After(now) && substep.DueDate.Before(tomorrow) {
					_ = d.NotificationService.CreateNotification(
						ctx,
						goal.UserID,
						"substep_due",
						"Substep Due Soon",
						fmt.Sprintf("Substep \"%s\" in goal \"%s\" is due soon.", substep.Title, goal.Name),
						&goal.ID,
					)
				}
			}
		}
	}

	logrus.Info(" Deadline scan completed: goal/step/substep")
	return nil
}
