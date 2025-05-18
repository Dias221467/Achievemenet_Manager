package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"github.com/Dias221467/Achievemenet_Manager/internal/services"
	"github.com/Dias221467/Achievemenet_Manager/pkg/logger"
	"github.com/Dias221467/Achievemenet_Manager/pkg/middleware"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TemplateHandler handles HTTP requests related to goal templates.
type TemplateHandler struct {
	TemplateService *services.TemplateService
	GoalService     *services.GoalService
}

// NewTemplateHandler creates a new instance of TemplateHandler.
func NewTemplateHandler(templateService *services.TemplateService, goalService *services.GoalService) *TemplateHandler {
	return &TemplateHandler{
		TemplateService: templateService,
		GoalService:     goalService,
	}
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

	createdTemplate, err := h.TemplateService.CreateTemplate(r.Context(), &template)
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

	templates, err := h.TemplateService.GetTemplatesByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to fetch templates", http.StatusInternalServerError)
		logger.Log.Errorf("Error fetching templates for user %s: %v", claims.UserID, err)
		return
	}

	logger.Log.Infof("Fetched %d templates for user %s", len(templates), claims.UserID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

func (h *TemplateHandler) GetTemplateByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID := vars["id"]

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		logger.Log.Warn("Unauthorized access to template by ID")
		return
	}

	// Parse ObjectID
	objID, err := primitive.ObjectIDFromHex(templateID)
	if err != nil {
		http.Error(w, "Invalid template ID", http.StatusBadRequest)
		logger.Log.Warnf("Invalid template ID: %v", err)
		return
	}

	template, err := h.TemplateService.GetTemplateByID(r.Context(), objID.Hex())
	if err != nil {
		http.Error(w, "Template not found", http.StatusNotFound)
		logger.Log.Warnf("Template not found: %v", err)
		return
	}

	// Make sure only the owner can view it (or add sharing logic later)
	if template.UserID.Hex() != claims.UserID {
		http.Error(w, "Forbidden: You can only view your own templates", http.StatusForbidden)
		logger.Log.Warnf("User %s tried to access template %s they do not own", claims.UserID, templateID)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

func (h *TemplateHandler) CopyTemplateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID := vars["id"]

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		logger.Log.Warn("Unauthorized attempt to copy template")
		return
	}

	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		logger.Log.Errorf("Failed to parse user ID: %v", err)
		return
	}

	goal, err := h.TemplateService.CopyTemplateToGoal(r.Context(), templateID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Log.Errorf("Failed to copy template: %v", err)
		return
	}

	logger.Log.Infof("User %s copied template %s into goal %s", claims.UserID, templateID, goal.ID.Hex())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goal)
}

func (h *TemplateHandler) GetPublicTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		logger.Log.Warn("Unauthorized attempt to fetch public templates")
		return
	}

	templates, err := h.TemplateService.GetPublicTemplates(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch public templates", http.StatusInternalServerError)
		logger.Log.Errorf("Error fetching public templates: %v", err)
		return
	}

	logger.Log.Infof("User %s fetched %d public templates", claims.UserID, len(templates))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

func (h *TemplateHandler) GetTemplatesByUserHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		logger.Log.Warn("Unauthorized request to get templates by user")
		return
	}

	vars := mux.Vars(r)
	requestedUserID := vars["id"]
	userID, err := primitive.ObjectIDFromHex(requestedUserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		logger.Log.Warnf("Invalid user ID: %v", err)
		return
	}

	// Parse optional ?public=true query param
	publicOnly := r.URL.Query().Get("public") == "true"

	var templates []models.GoalTemplate

	if publicOnly {
		templates, err = h.TemplateService.GetPublicTemplatesByUser(r.Context(), userID)
	} else if claims.UserID == requestedUserID {
		templates, err = h.TemplateService.GetTemplatesByUser(r.Context(), userID)
	} else {
		http.Error(w, "Forbidden: You can only view your own private templates", http.StatusForbidden)
		logger.Log.Warnf("User %s attempted to access private templates of user %s", claims.UserID, requestedUserID)
		return
	}

	if err != nil {
		http.Error(w, "Failed to retrieve templates", http.StatusInternalServerError)
		logger.Log.Errorf("Failed to get templates for user %s: %v", requestedUserID, err)
		return
	}

	logger.Log.Infof("User %s fetched %d templates for user %s", claims.UserID, len(templates), requestedUserID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

func (h *TemplateHandler) AdminGetAllTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		logger.Log.Warn("Unauthorized attempt to access all templates")
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: Admins only", http.StatusForbidden)
		logger.Log.Warnf("User %s attempted to access admin-only endpoint", claims.UserID)
		return
	}

	templates, err := h.TemplateService.GetAllTemplates(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch templates", http.StatusInternalServerError)
		logger.Log.Errorf("Admin failed to fetch all templates: %v", err)
		return
	}

	logger.Log.Infof("Admin %s fetched %d templates", claims.UserID, len(templates))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}
