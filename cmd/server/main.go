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

	// Initialize repositories, services, and handlers for goals
	goalRepo := repository.NewGoalRepository(db)
	goalService := services.NewGoalService(goalRepo)
	goalHandler := handlers.NewGoalHandler(goalService)

	// Initialize repositories, services, and handlers for users
	userRepo := repository.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService, cfg)

	// Initialize repository and service for templates
	templateRepo := repository.NewTemplateRepository(db)
	templateService := services.NewTemplateService(templateRepo, goalRepo) // <- inject goalRepo
	templateHandler := handlers.NewTemplateHandler(templateService, goalService)

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

	// Register User routes
	router.HandleFunc("/users/register", userHandler.RegisterUserHandler).Methods("POST")
	router.HandleFunc("/users/login", userHandler.LoginUserHandler).Methods("POST")

	// Protected user routes (only authenticated users can access)
	protectedUserRoutes := router.PathPrefix("/users").Subrouter()
	protectedUserRoutes.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	protectedUserRoutes.HandleFunc("/{id}", userHandler.GetUserHandler).Methods("GET")
	protectedUserRoutes.HandleFunc("/{id}", userHandler.UpdateUserHandler).Methods("PUT")

	// Template-related routes
	protectedTemplateRoutes := router.PathPrefix("/templates").Subrouter()
	protectedTemplateRoutes.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	protectedTemplateRoutes.HandleFunc("", templateHandler.CreateTemplateHandler).Methods("POST")
	protectedTemplateRoutes.HandleFunc("", templateHandler.GetTemplatesHandler).Methods("GET")
	protectedTemplateRoutes.HandleFunc("/{id}/copy", templateHandler.CopyTemplateHandler).Methods("POST")

	// Apply middleware for logging
	router.Use(middleware.LoggingMiddleware)

	// Start the HTTP server
	port := cfg.Port
	fmt.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
