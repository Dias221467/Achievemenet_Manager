package cron

import (
	"context"

	"github.com/Dias221467/Achievemenet_Manager/internal/services"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

func StartNotificationCronJobs(notificationService *services.NotificationService) {
	c := cron.New()

	// Inactive user reminders
	c.AddFunc("0 0 * * *", func() {
		err := notificationService.CheckInactiveUsers(context.Background())
		if err != nil {
			logrus.WithError(err).Error("CheckInactiveUsers failed")
		}
	})

	// Step due soon
	c.AddFunc("@hourly", func() {
		err := notificationService.CheckStepDueSoon(context.Background())
		if err != nil {
			logrus.WithError(err).Error("CheckStepDueSoon failed")
		}
	})

	// Goal due soon
	c.AddFunc("@hourly", func() {
		err := notificationService.CheckGoalDueSoon(context.Background())
		if err != nil {
			logrus.WithError(err).Error("CheckGoalDueSoon failed")
		}
	})

	c.Start()

	// Check substep deadlines hourly
	c.AddFunc("@hourly", func() {
		err := notificationService.CheckSubstepDueSoon(context.Background())
		if err != nil {
			logrus.WithError(err).Error("CheckSubstepDueSoon failed")
		}
	})
}
