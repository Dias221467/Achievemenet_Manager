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
	"github.com/Dias221467/Achievemenet_Manager/pkg/middleware"
	"github.com/gorilla/mux"
)

func main() {
	// Load configuration from .env file
	cfg := config.LoadConfig()

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

	// Initialize Gorilla Mux router
	router := mux.NewRouter()

	// Register Goal routes
	router.HandleFunc("/goals", goalHandler.CreateGoalHandler).Methods("POST")
	router.HandleFunc("/goals", goalHandler.GetAllGoalsHandler).Methods("GET")
	router.HandleFunc("/goals/{id}", goalHandler.GetGoalHandler).Methods("GET")
	router.HandleFunc("/goals/{id}", goalHandler.UpdateGoalHandler).Methods("PUT")
	router.HandleFunc("/goals/{id}", goalHandler.DeleteGoalHandler).Methods("DELETE")

	// Register User routes
	router.HandleFunc("/users/register", userHandler.RegisterUserHandler).Methods("POST")
	router.HandleFunc("/users/login", userHandler.LoginUserHandler).Methods("POST")
	router.HandleFunc("/users/{id}", userHandler.GetUserHandler).Methods("GET")

	// Apply middleware for logging
	router.Use(middleware.LoggingMiddleware)

	// Start the HTTP server
	port := cfg.Port
	fmt.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
