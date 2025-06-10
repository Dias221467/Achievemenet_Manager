package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Dias221467/Achievemenet_Manager/internal/config"
	"github.com/Dias221467/Achievemenet_Manager/internal/database"
	"github.com/Dias221467/Achievemenet_Manager/internal/handlers"
	"github.com/Dias221467/Achievemenet_Manager/internal/repository"
	"github.com/Dias221467/Achievemenet_Manager/internal/services"
	"github.com/Dias221467/Achievemenet_Manager/pkg/logger"
	"github.com/Dias221467/Achievemenet_Manager/pkg/middleware"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
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

	// --- Services ---
	userService := services.NewUserService(userRepo)
	goalService := services.NewGoalService(goalRepo, userRepo)
	friendService := services.NewFriendService(friendRepo, userRepo)
	templateService := services.NewTemplateService(templateRepo, goalRepo)
	wishService := services.NewWishService(wishRepo, goalRepo)

	// --- Handlers ---
	userHandler := handlers.NewUserHandler(userService, cfg)
	goalHandler := handlers.NewGoalHandler(goalService)
	friendHandler := handlers.NewFriendHandler(friendService)
	templateHandler := handlers.NewTemplateHandler(templateService, goalService)
	wishHandler := handlers.NewWishHandler(wishService, goalService)

	// Initialize Gorilla Mux router
	router := mux.NewRouter()

	// Apply authentication middleware to goal routes
	protectedRoutes := router.PathPrefix("/goals").Subrouter()
	protectedRoutes.Use(middleware.AuthMiddleware(cfg.JWTSecret))
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
	protectedUserRoutes.HandleFunc("/{id}", userHandler.GetUserHandler).Methods("GET")
	protectedUserRoutes.HandleFunc("/{id}", userHandler.UpdateUserHandler).Methods("PATCH")

	// Template-related routes
	protectedTemplateRoutes := router.PathPrefix("/templates").Subrouter()
	protectedTemplateRoutes.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	protectedTemplateRoutes.HandleFunc("", templateHandler.CreateTemplateHandler).Methods("POST")
	protectedTemplateRoutes.HandleFunc("", templateHandler.GetTemplatesHandler).Methods("GET")
	protectedTemplateRoutes.HandleFunc("/{id}", templateHandler.GetTemplateByIDHandler).Methods("GET")
	protectedTemplateRoutes.HandleFunc("/{id}/copy", templateHandler.CopyTemplateHandler).Methods("POST")
	protectedTemplateRoutes.HandleFunc("/public", templateHandler.GetPublicTemplatesHandler).Methods("GET")
	protectedTemplateRoutes.HandleFunc("/user/{id}", templateHandler.GetTemplatesByUserHandler).Methods("GET")

	// Friend routes
	protectedFriendRoutes := router.PathPrefix("/friends").Subrouter()
	protectedFriendRoutes.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	protectedFriendRoutes.HandleFunc("/{id}/request", friendHandler.SendFriendRequestHandler).Methods("POST")
	protectedFriendRoutes.HandleFunc("/requests", friendHandler.GetPendingRequestsHandler).Methods("GET")
	protectedFriendRoutes.HandleFunc("/requests/{id}/respond", friendHandler.RespondToFriendRequestHandler).Methods("POST")
	protectedFriendRoutes.HandleFunc("", friendHandler.GetFriendsHandler).Methods("GET")
	protectedFriendRoutes.HandleFunc("/{id}", friendHandler.RemoveFriendHandler).Methods("DELETE")

	// Wish routes
	protectedWishRoutes := router.PathPrefix("/wishes").Subrouter()
	protectedWishRoutes.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	protectedWishRoutes.HandleFunc("", wishHandler.CreateWishHandler).Methods("POST")
	protectedWishRoutes.HandleFunc("", wishHandler.GetWishesHandler).Methods("GET")
	protectedWishRoutes.HandleFunc("/{id}", wishHandler.GetWishByIDHandler).Methods("GET")
	protectedWishRoutes.HandleFunc("/{id}", wishHandler.UpdateWishHandler).Methods("PUT")
	protectedWishRoutes.HandleFunc("/{id}", wishHandler.DeleteWishHandler).Methods("DELETE")
	protectedWishRoutes.HandleFunc("/{id}/promote", wishHandler.PromoteWishHandler).Methods("POST")

	protectedWishRoutes.HandleFunc("/{id}/upload", wishHandler.UploadWishImageHandler).Methods("POST")
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads/"))))

	// Admin routes
	adminRoutes := router.PathPrefix("/admin").Subrouter()
	adminRoutes.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	adminRoutes.Use(middleware.RequireRole("admin"))
	adminRoutes.HandleFunc("/goals", goalHandler.GetAllGoalsHandler).Methods("GET")
	adminRoutes.HandleFunc("/templates", templateHandler.AdminGetAllTemplatesHandler).Methods("GET")
	adminRoutes.HandleFunc("/users", userHandler.AdminGetAllUsersHandler).Methods("GET")

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

	fmt.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
