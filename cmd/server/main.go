package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Dias221467/Achievemenet_Manager/internal/config"
	"github.com/Dias221467/Achievemenet_Manager/internal/database"
	"github.com/Dias221467/Achievemenet_Manager/internal/handlers"
	"github.com/Dias221467/Achievemenet_Manager/internal/jobs"
	"github.com/Dias221467/Achievemenet_Manager/internal/repository"
	"github.com/Dias221467/Achievemenet_Manager/internal/services"
	"github.com/Dias221467/Achievemenet_Manager/pkg/logger"
	"github.com/Dias221467/Achievemenet_Manager/pkg/middleware"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration from .env file
	cfg := config.LoadConfig()

	logger.InitLogger()
	logger.Log.Info("Logger initialized")

	// Connect to MongoDB Atlas
	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}

	// --- Repositories ---
	userRepo := repository.NewUserRepository(db)
	goalRepo := repository.NewGoalRepository(db)
	friendRepo := repository.NewFriendRepository(db)
	templateRepo := repository.NewTemplateRepository(db)
	wishRepo := repository.NewWishRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	// --- Services ---
	userService := services.NewUserService(userRepo)
	goalService := services.NewGoalService(goalRepo, userRepo, services.NewNotificationService(notificationRepo, userRepo, goalRepo))
	friendService := services.NewFriendService(friendRepo, userRepo)
	templateService := services.NewTemplateService(templateRepo, goalRepo)
	wishService := services.NewWishService(wishRepo, goalRepo)
	activityService := services.NewActivityService(activityRepo)
	notificationService := services.NewNotificationService(notificationRepo, userRepo, goalRepo)

	// --- Handlers ---
	userHandler := handlers.NewUserHandler(userService, cfg)
	goalHandler := handlers.NewGoalHandler(goalService, activityService, notificationService)
	friendHandler := handlers.NewFriendHandler(friendService, activityService, notificationService, userService)
	templateHandler := handlers.NewTemplateHandler(templateService, goalService, activityService)
	wishHandler := handlers.NewWishHandler(wishService, goalService, activityService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	// ----deadline_notifier ----
	deadlinRepo := jobs.NewDeadlineNotifier(goalService, notificationService)

	// Initialize Gorilla Mux router
	router := mux.NewRouter()

	// Apply authentication middleware to goal routes
	protectedRoutes := router.PathPrefix("/goals").Subrouter()
	protectedRoutes.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	protectedRoutes.Use(middleware.UpdateLastActiveMiddleware(userService))

	protectedRoutes.HandleFunc("", goalHandler.CreateGoalHandler).Methods("POST")
	protectedRoutes.HandleFunc("/{id}", goalHandler.GetGoalHandler).Methods("GET")
	protectedRoutes.HandleFunc("/{id}", goalHandler.UpdateGoalHandler).Methods("PUT")
	protectedRoutes.HandleFunc("/{id}", goalHandler.DeleteGoalHandler).Methods("DELETE")
	protectedRoutes.HandleFunc("/{id}/progress", goalHandler.UpdateGoalProgressHandler).Methods("PATCH")
	protectedRoutes.HandleFunc("/{id}/progress", goalHandler.GetGoalProgressHandler).Methods("GET")
	protectedRoutes.HandleFunc("", goalHandler.GetGoalsHandler).Methods("GET")
	protectedRoutes.HandleFunc("/{id}/invite", goalHandler.InviteCollaboratorHandler).Methods("POST")

	// Register User routes
	router.HandleFunc("/users/register", userHandler.RegisterUserHandler).Methods("POST")
	router.HandleFunc("/users/login", userHandler.LoginUserHandler).Methods("POST")
	router.HandleFunc("/users/verify", userHandler.VerifyEmailHandler).Methods("GET")

	// Password reset routes
	router.HandleFunc("/users/request-password-reset", userHandler.RequestPasswordResetHandler).Methods("POST")
	router.HandleFunc("/users/reset-password", userHandler.ResetPasswordHandler).Methods("POST")

	// Protected user routes (only authenticated users can access)
	protectedUserRoutes := router.PathPrefix("/users").Subrouter()
	protectedUserRoutes.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	protectedUserRoutes.Use(middleware.UpdateLastActiveMiddleware(userService))

	protectedUserRoutes.HandleFunc("/{id}", userHandler.GetUserHandler).Methods("GET")
	protectedUserRoutes.HandleFunc("/{id}", userHandler.UpdateUserHandler).Methods("PATCH")
	protectedUserRoutes.HandleFunc("", userHandler.GetAllUsersHandler).Methods("GET")

	// Template-related routes
	protectedTemplateRoutes := router.PathPrefix("/templates").Subrouter()
	protectedTemplateRoutes.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	protectedTemplateRoutes.Use(middleware.UpdateLastActiveMiddleware(userService))

	protectedTemplateRoutes.HandleFunc("", templateHandler.CreateTemplateHandler).Methods("POST")
	protectedTemplateRoutes.HandleFunc("", templateHandler.GetTemplatesHandler).Methods("GET")
	protectedTemplateRoutes.HandleFunc("/public", templateHandler.GetPublicTemplatesHandler).Methods("GET")
	protectedTemplateRoutes.HandleFunc("/user/{id}", templateHandler.GetTemplatesByUserHandler).Methods("GET")
	protectedTemplateRoutes.HandleFunc("/{id}", templateHandler.GetTemplateByIDHandler).Methods("GET")
	protectedTemplateRoutes.HandleFunc("/{id}/copy", templateHandler.CopyTemplateHandler).Methods("POST")

	// Friend routes
	protectedFriendRoutes := router.PathPrefix("/friends").Subrouter()
	protectedFriendRoutes.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	protectedFriendRoutes.Use(middleware.UpdateLastActiveMiddleware(userService))

	protectedFriendRoutes.HandleFunc("/{id}/request", friendHandler.SendFriendRequestHandler).Methods("POST")
	protectedFriendRoutes.HandleFunc("/requests", friendHandler.GetPendingRequestsHandler).Methods("GET")
	protectedFriendRoutes.HandleFunc("/requests/{id}/respond", friendHandler.RespondToFriendRequestHandler).Methods("POST")
	protectedFriendRoutes.HandleFunc("", friendHandler.GetFriendsHandler).Methods("GET")
	protectedFriendRoutes.HandleFunc("/{id}", friendHandler.RemoveFriendHandler).Methods("DELETE")

	// Wish routes
	protectedWishRoutes := router.PathPrefix("/wishes").Subrouter()
	protectedWishRoutes.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	protectedWishRoutes.Use(middleware.UpdateLastActiveMiddleware(userService))

	protectedWishRoutes.HandleFunc("", wishHandler.CreateWishHandler).Methods("POST")
	protectedWishRoutes.HandleFunc("", wishHandler.GetWishesHandler).Methods("GET")
	protectedWishRoutes.HandleFunc("/{id}", wishHandler.GetWishByIDHandler).Methods("GET")
	protectedWishRoutes.HandleFunc("/{id}", wishHandler.UpdateWishHandler).Methods("PUT")
	protectedWishRoutes.HandleFunc("/{id}", wishHandler.DeleteWishHandler).Methods("DELETE")
	protectedWishRoutes.HandleFunc("/{id}/promote", wishHandler.PromoteWishHandler).Methods("POST")

	protectedWishRoutes.HandleFunc("/{id}/upload", wishHandler.UploadWishImageHandler).Methods("POST")
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads/"))))

	// Notifications routes
	protectedNotificationRoutes := router.PathPrefix("/notifications").Subrouter()
	protectedNotificationRoutes.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	protectedNotificationRoutes.HandleFunc("", notificationHandler.GetUserNotificationsHandler).Methods("GET")
	protectedNotificationRoutes.HandleFunc("/{id}/read", notificationHandler.MarkAsReadHandler).Methods("POST")
	protectedNotificationRoutes.HandleFunc("/{id}", notificationHandler.DeleteNotificationHandler).Methods("DELETE")

	// Admin routes
	adminRoutes := router.PathPrefix("/admin").Subrouter()
	adminRoutes.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	adminRoutes.Use(middleware.RequireRole("admin"))
	adminRoutes.HandleFunc("/goals", goalHandler.GetAllGoalsHandler).Methods("GET")
	adminRoutes.HandleFunc("/templates", templateHandler.AdminGetAllTemplatesHandler).Methods("GET")

	// Apply middleware for logging
	router.Use(middleware.LoggingMiddleware)

	// Start the HTTP server
	port := cfg.Port
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // adjust to frontend origin
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	notifier := jobs.NewDeadlineNotifier(goalService, notificationService)
	go func() {
		for {
			notifier.RunDailyScan(context.Background())
			time.Sleep(24 * time.Hour)
		}
	}()

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		for range ticker.C {
			ctx := context.Background()
			if err := notificationService.CheckInactiveUsers(ctx); err != nil {
				logrus.WithError(err).Error("Failed to run inactive user check")
			}
		}
	}()

	go deadlinRepo.RunDailyScan(context.Background())

	fmt.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
