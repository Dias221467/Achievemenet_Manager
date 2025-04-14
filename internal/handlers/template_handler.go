package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"github.com/Dias221467/Achievemenet_Manager/internal/services"
	"github.com/Dias221467/Achievemenet_Manager/pkg/logger"
	"github.com/Dias221467/Achievemenet_Manager/pkg/middleware"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TemplateHandler handles HTTP requests related to goal templates.
type TemplateHandler struct {
	Service *services.TemplateService
}

// NewTemplateHandler creates a new instance of TemplateHandler.
func NewTemplateHandler(service *services.TemplateService) *TemplateHandler {
	return &TemplateHandler{Service: service}
}

// CreateTemplateHandler allows a user to create a goal template.
func (h *TemplateHandler) CreateTemplateHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		logger.Log.Warn("Unauthorized attempt to create a template")
		return
	}

	var template models.GoalTemplate
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		logger.Log.Warnf("Failed to decode template: %v", err)
		return
	}
	defer r.Body.Close()

	if template.Title == "" || len(template.Steps) == 0 {
		http.Error(w, "Title and steps are required", http.StatusBadRequest)
		logger.Log.Warn("Missing required template fields")
		return
	}

	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		logger.Log.Errorf("Failed to parse user ID: %v", err)
		return
	}

	template.UserID = userID
	template.CreatedAt = time.Now()

	createdTemplate, err := h.Service.CreateTemplate(r.Context(), &template)
	if err != nil {
		http.Error(w, "Failed to create template", http.StatusInternalServerError)
		logger.Log.Errorf("Error creating template: %v", err)
		return
	}

	logger.Log.Infof("User %s created template %s", claims.UserID, createdTemplate.ID.Hex())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdTemplate)
}

// GetTemplatesHandler allows a user to fetch their own templates.
func (h *TemplateHandler) GetTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		logger.Log.Warn("Unauthorized attempt to fetch templates")
		return
	}

	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		logger.Log.Errorf("Failed to parse user ID: %v", err)
		return
	}

	templates, err := h.Service.GetTemplatesByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to fetch templates", http.StatusInternalServerError)
		logger.Log.Errorf("Error fetching templates for user %s: %v", claims.UserID, err)
		return
	}

	logger.Log.Infof("Fetched %d templates for user %s", len(templates), claims.UserID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}
