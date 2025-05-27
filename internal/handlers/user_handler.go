package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Dias221467/Achievemenet_Manager/internal/config"
	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"github.com/Dias221467/Achievemenet_Manager/internal/services"
	jwtutil "github.com/Dias221467/Achievemenet_Manager/pkg/jwt"
	"github.com/Dias221467/Achievemenet_Manager/pkg/logger"
	"github.com/Dias221467/Achievemenet_Manager/pkg/middleware"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// UserHandler handles HTTP requests related to user operations.
type UserHandler struct {
	Service *services.UserService
	Config  *config.Config
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(service *services.UserService, cfg *config.Config) *UserHandler {
	return &UserHandler{
		Service: service,
		Config:  cfg,
	}
}

// RegisterUserHandler handles user registration.
func (h *UserHandler) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("RegisterUserHandler called")
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.WithError(err).Warn("Failed to decode user registration request")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	createdUser, err := h.Service.RegisterUser(r.Context(), &user)
	if err != nil {
		log.WithError(err).Error("Failed to register user")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.WithField("userID", createdUser.ID.Hex()).Info("User registered successfully")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdUser)
}

// LoginUserHandler handles user login.
func (h *UserHandler) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	// Define a simple struct to receive login credentials.
	log.Info("LoginUserHandler called")
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		log.WithError(err).Warn("Failed to decode login request")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	user, err := h.Service.AuthenticateUser(r.Context(), credentials.Email, credentials.Password)
	if err != nil {
		log.WithFields(log.Fields{
			"email": credentials.Email,
			"error": err,
		}).Warn("Authentication failed")
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Generate a JWT token
	token, err := jwtutil.GenerateToken(user.ID.Hex(), user.Email, user.Role, h.Config.JWTSecret, h.Config.TokenExpiry)
	if err != nil {
		log.WithError(err).Error("Failed to generate JWT token")
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	log.WithField("userID", user.ID.Hex()).Info("User logged in successfully")

	// Return the token and user details
	response := map[string]interface{}{
		"token": token,
		"user":  user,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUserHandler handles fetching a user by ID.
func (h *UserHandler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("GetUserHandler called")
	vars := mux.Vars(r)
	requestedUserID := vars["id"]

	// Get the logged-in user from the request context
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		log.Warn("Unauthorized access attempt to GetUserHandler")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Ensure that the requested user ID matches the logged-in user’s ID
	if requestedUserID != claims.UserID {
		log.WithFields(log.Fields{
			"requestedUserID": requestedUserID,
			"loggedInUserID":  claims.UserID,
		}).Warn("Forbidden access attempt")
		http.Error(w, "Forbidden: You can only access your own profile", http.StatusForbidden)
		return
	}

	// Fetch the user from the database
	user, err := h.Service.GetUser(r.Context(), requestedUserID)
	if err != nil {
		log.WithField("userID", requestedUserID).WithError(err).Warn("User not found")
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	log.WithField("userID", user.ID.Hex()).Info("User profile fetched")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// UpdateUserHandler handles updating a user profile.
func (h *UserHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("UpdateUserHandler called")
	vars := mux.Vars(r)
	requestedUserID := vars["id"]

	// Get logged-in user
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		log.Warn("Unauthorized access attempt to UpdateUserHandler")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Ensure only the logged-in user can update their own profile
	if requestedUserID != claims.UserID {
		log.WithFields(log.Fields{
			"requestedUserID": requestedUserID,
			"loggedInUserID":  claims.UserID,
		}).Warn("Forbidden update attempt")
		http.Error(w, "Forbidden: You can only update your own profile", http.StatusForbidden)
		return
	}

	// Decode request body
	var updatedUser models.User
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		log.WithError(err).Warn("Failed to decode update request")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Update user in DB
	updatedUserData, err := h.Service.UpdateUser(r.Context(), requestedUserID, &updatedUser)
	if err != nil {
		log.WithFields(log.Fields{
			"userID": requestedUserID,
			"error":  err,
		}).Error("Failed to update user")
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	log.WithField("userID", updatedUserData.ID.Hex()).Info("User updated successfully")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUserData)
}

func (h *UserHandler) AdminGetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		logger.Log.Warn("Unauthorized attempt to fetch all users")
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: Admins only", http.StatusForbidden)
		logger.Log.Warnf("User %s attempted to access admin-only user list", claims.UserID)
		return
	}

	users, err := h.Service.GetAllUsers(r.Context())
	if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		logger.Log.Errorf("Admin %s failed to fetch users: %v", claims.UserID, err)
		return
	}

	logger.Log.Infof("Admin %s fetched %d users", claims.UserID, len(users))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}


// ========================================================= By Alibek ==================================================================//
func (h *UserHandler) GetMeHandler(w http.ResponseWriter, r *http.Request) {
    claims := middleware.GetUserFromContext(r.Context())
    if claims == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    user, err := h.Service.GetUser(r.Context(), claims.UserID)
    if err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) SearchUsersHandler(w http.ResponseWriter, r *http.Request) {
    claims := middleware.GetUserFromContext(r.Context())
    if claims == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    query := r.URL.Query().Get("query")
    users, err := h.Service.SearchUsers(r.Context(), query)
    if err != nil {
        http.Error(w, "Failed to search users", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

