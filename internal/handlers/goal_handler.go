package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"github.com/Dias221467/Achievemenet_Manager/internal/services"
	"github.com/Dias221467/Achievemenet_Manager/pkg/logger"
	"github.com/Dias221467/Achievemenet_Manager/pkg/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GoalHandler handles HTTP requests related to goals.
type GoalHandler struct {
	Service             *services.GoalService
	ActivityService     *services.ActivityService
	NotificationService *services.NotificationService
}

// NewGoalHandler creates a new instance of GoalHandler.
func NewGoalHandler(goalService *services.GoalService, activityService *services.ActivityService, notificationService *services.NotificationService) *GoalHandler {
	return &GoalHandler{
		Service:             goalService,
		ActivityService:     activityService,
		NotificationService: notificationService,
	}
}

// CreateGoalHandler handles the creation of a new goal.
func (h *GoalHandler) CreateGoalHandler(w http.ResponseWriter, r *http.Request) {
	// Get the logged-in user from JWT token
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		logrus.Warn("Unauthorized access attempt during goal creation")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode request body
	var goal models.Goal
	if err := json.NewDecoder(r.Body).Decode(&goal); err != nil {
		logrus.WithError(err).Warn("Invalid request payload during goal creation")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Convert UserID to ObjectID
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to convert user ID")
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}
	goal.UserID = userID
	goal.CreatedAt = time.Now()
	goal.UpdatedAt = time.Now()
	goal.Status = "in_progress"

	//  Validate & Parse Due Date (Optional)
	if !goal.DueDate.IsZero() && goal.DueDate.Before(time.Now()) {
		logrus.Warn("Attempt to set a past due date for goal")
		http.Error(w, "Due date cannot be in the past", http.StatusBadRequest)
		return
	}

	//  Validate & Set Category (Optional)
	if goal.Category != "" {
		if _, exists := models.AllowedCategories[goal.Category]; !exists {
			logrus.Warn("Invalid category provided: ", goal.Category)
			http.Error(w, "Invalid category", http.StatusBadRequest)
			return
		}
	}

	// Auto-calculate completion state of each step
	for i := range goal.Steps {
		allDone := true
		for _, sub := range goal.Steps[i].Substeps {
			if !sub.Done {
				allDone = false
				break
			}
		}
		goal.Steps[i].Completed = allDone
	}

	// Save to DB
	createdGoal, err := h.Service.CreateGoal(r.Context(), &goal)
	if err != nil {
		logrus.WithError(err).Error("Failed to create goal")
		http.Error(w, "Failed to create goal", http.StatusInternalServerError)
		return
	}

	// Log activity
	_ = h.ActivityService.LogActivity(r.Context(), userID, "goal_created", createdGoal.ID, fmt.Sprintf("Created goal: %s", createdGoal.Name))

	logrus.WithFields(logrus.Fields{
		"userID": claims.UserID,
		"goalID": createdGoal.ID.Hex(),
	}).Info("Goal successfully created")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdGoal)
}

// GetGoalHandler handles fetching a single goal by its ID.
func (h *GoalHandler) GetGoalHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	goalID := vars["id"]

	// Get the logged-in user
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		logrus.Warn("Unauthorized goal fetch attempt")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch the goal from DB
	goal, err := h.Service.GetGoal(r.Context(), goalID)
	if err != nil || goal == nil {
		logrus.WithField("goalID", goalID).Warn("Goal not found")
		http.Error(w, "Goal not found", http.StatusNotFound)
		return
	}

	//  Ensure the logged-in user is the owner of the goal
	if goal.UserID.Hex() != claims.UserID && !isCollaborator(goal.Collaborators, claims.UserID) {
		logrus.WithFields(logrus.Fields{
			"userID": claims.UserID,
			"goalID": goalID,
		}).Warn("Forbidden: User tried to access goal without permission")
		http.Error(w, "Forbidden: You can only view your own or shared goals", http.StatusForbidden)
		return
	}

	//  Check if the goal is overdue
	if !goal.DueDate.IsZero() && goal.DueDate.Before(time.Now()) {
		goal.Status = "expired"
	}

	logrus.WithFields(logrus.Fields{
		"userID": claims.UserID,
		"goalID": goalID,
	}).Info("Goal successfully fetched")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goal)
}

// UpdateGoalHandler handles updating an existing goal.
func (h *GoalHandler) UpdateGoalHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	goalID := vars["id"]

	// Get the logged-in user
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		logrus.Warn("Unauthorized update attempt")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert goalID to ObjectID
	objID, err := primitive.ObjectIDFromHex(goalID)
	if err != nil {
		logrus.WithError(err).Warn("Invalid goal ID format during update")
		http.Error(w, "Invalid goal ID", http.StatusBadRequest)
		return
	}

	// Fetch the existing goal
	existingGoal, err := h.Service.GetGoal(r.Context(), goalID)
	if err != nil || existingGoal == nil {
		logrus.WithField("goalID", goalID).Warn("Goal not found during update")
		http.Error(w, "Goal not found", http.StatusNotFound)
		return
	}

	// Ensure the logged-in user is the owner of the goal
	if existingGoal.UserID.Hex() != claims.UserID && !isCollaborator(existingGoal.Collaborators, claims.UserID) {
		logrus.WithFields(logrus.Fields{
			"userID": claims.UserID,
			"goalID": goalID,
		}).Warn("Forbidden: Update attempt by non-owner and non-collaborator")
		http.Error(w, "Forbidden: Only owner or collaborators can update the goal", http.StatusForbidden)
		return
	}

	// Decode request body
	var updatedGoal models.Goal
	if err := json.NewDecoder(r.Body).Decode(&updatedGoal); err != nil {
		logrus.WithError(err).Warn("Invalid update payload")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	//  Validate & Parse Due Date (Optional)
	if !updatedGoal.DueDate.IsZero() && updatedGoal.DueDate.Before(time.Now()) {
		http.Error(w, "Due date cannot be in the past", http.StatusBadRequest)
		return
	}

	//  Validate & Set Category (Optional)
	if updatedGoal.Category != "" {
		if _, exists := models.AllowedCategories[updatedGoal.Category]; !exists {
			http.Error(w, "Invalid category", http.StatusBadRequest)
			return
		}
	}

	// Auto-complete parent step when all substeps are done
	for i := range updatedGoal.Steps {
		step := &updatedGoal.Steps[i]
		allSubstepsDone := true
		for _, sub := range step.Substeps {
			if !sub.Done {
				allSubstepsDone = false
				break
			}
		}
		step.Completed = allSubstepsDone
	}

	// Auto-update goal status based on steps
	allStepsDone := true
	for _, step := range updatedGoal.Steps {
		if !step.Completed {
			allStepsDone = false
			break
		}
	}
	if allStepsDone {
		updatedGoal.Status = "completed"
	} else {
		updatedGoal.Status = "in_progress"
	}

	//  Assign updated values
	updatedGoal.ID = objID
	updatedGoal.UserID = existingGoal.UserID
	updatedGoal.Collaborators = existingGoal.Collaborators
	updatedGoal.CreatedAt = existingGoal.CreatedAt
	updatedGoal.UpdatedAt = time.Now()

	// Save the updated goal
	updatedGoalData, err := h.Service.UpdateGoal(r.Context(), goalID, &updatedGoal)
	if err != nil {
		logrus.WithError(err).Error("Failed to update goal")
		http.Error(w, "Failed to update goal", http.StatusInternalServerError)
		return
	}

	_ = h.ActivityService.LogActivity(r.Context(), existingGoal.UserID, "goal_updated", updatedGoal.ID, fmt.Sprintf("Updated goal: %s", updatedGoal.Name))

	logrus.WithFields(logrus.Fields{
		"userID": claims.UserID,
		"goalID": goalID,
	}).Info("Goal successfully updated")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedGoalData)
}

func (h *GoalHandler) UpdateGoalProgressHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	goalID := vars["id"]
	log := logrus.WithField("goalID", goalID)

	// Get logged-in user
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		log.Warn("Unauthorized access")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch goal from DB
	goal, err := h.Service.GetGoal(r.Context(), goalID)
	if err != nil || goal == nil {
		log.WithError(err).Warn("Goal not found")
		http.Error(w, "Goal not found", http.StatusNotFound)
		return
	}

	// Ensure the logged-in user owns the goal
	if goal.UserID.Hex() != claims.UserID && !isCollaborator(goal.Collaborators, claims.UserID) {
		log.Warn("Forbidden: User is not the owner or a collaborator")
		http.Error(w, "Forbidden: Only owner or collaborators can update progress", http.StatusForbidden)
		return
	}

	// Decode request body
	var progressUpdate struct {
		StepName   string `json:"step"`
		SubstepIdx int    `json:"substep_index"`
		Done       bool   `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&progressUpdate); err != nil {
		log.WithError(err).Warn("Invalid request payload")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Find the step by name
	var stepFound bool
	for i := range goal.Steps {
		if goal.Steps[i].Name == progressUpdate.StepName {
			stepFound = true
			// Validate substep index
			if progressUpdate.SubstepIdx < 0 || progressUpdate.SubstepIdx >= len(goal.Steps[i].Substeps) {
				http.Error(w, "Invalid substep index", http.StatusBadRequest)
				return
			}

			// Update the substep's done status
			goal.Steps[i].Substeps[progressUpdate.SubstepIdx].Done = progressUpdate.Done

			// Auto-complete the step if all substeps are done
			allDone := true
			for _, sub := range goal.Steps[i].Substeps {
				if !sub.Done {
					allDone = false
					break
				}
			}
			goal.Steps[i].Completed = allDone
			break
		}
	}

	if !stepFound {
		http.Error(w, "Step not found", http.StatusBadRequest)
		return
	}

	// Check if all steps are completed to set goal status
	allStepsCompleted := true
	for _, step := range goal.Steps {
		if !step.Completed {
			allStepsCompleted = false
			break
		}
	}
	if allStepsCompleted {
		goal.Status = "completed"
	} else {
		goal.Status = "in_progress"
	}

	goal.UpdatedAt = time.Now()

	// Save changes
	updatedGoal, err := h.Service.UpdateGoal(r.Context(), goalID, goal)
	if err != nil {
		log.WithError(err).Error("Failed to update goal progress in DB")
		http.Error(w, "Failed to update progress", http.StatusInternalServerError)
		return
	}

	_ = h.ActivityService.LogActivity(r.Context(), goal.UserID, "goal_progress_updated", goal.ID, fmt.Sprintf("Updated progress for goal: %s", goal.Name))

	log.Info("Goal progress successfully updated")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedGoal)
}

// DeleteGoalHandler handles deleting a goal by its ID.
func (h *GoalHandler) DeleteGoalHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	goalID := vars["id"]
	log := logrus.WithField("goalID", goalID)

	// Get the logged-in user from JWT token
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		log.Warn("Unauthorized access attempt")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch the goal from DB
	goal, err := h.Service.GetGoal(r.Context(), goalID)
	if err != nil || goal == nil {
		log.WithError(err).Warn("Goal not found or fetch failed")
		http.Error(w, "Goal not found", http.StatusNotFound)
		return
	}

	// Check if the logged-in user is the owner
	if goal.UserID.Hex() != claims.UserID {
		log.Warn("Forbidden: User tried to delete another user's goal")
		http.Error(w, "Forbidden: You can only delete your own goals", http.StatusForbidden)
		return
	}

	// Perform delete
	err = h.Service.DeleteGoal(r.Context(), goalID)
	if err != nil {
		log.WithError(err).Error("Failed to delete goal")
		http.Error(w, "Failed to delete goal", http.StatusInternalServerError)
		return
	}

	_ = h.ActivityService.LogActivity(r.Context(), goal.UserID, "goal_deleted", goal.ID, fmt.Sprintf("Deleted goal: %s", goal.Name))

	log.Info("Goal deleted successfully")
	w.WriteHeader(http.StatusNoContent)
}

// GetAllGoalsHandler handles fetching all goals, with an optional limit.

// Its not working right now, we will need it later when we will add admins and their rights with functions
func (h *GoalHandler) GetAllGoalsHandler(w http.ResponseWriter, r *http.Request) {
	limitParam := r.URL.Query().Get("limit")
	var limit int64 = 10 // default limit
	log := logrus.WithField("defaultLimit", limit)

	if limitParam != "" {
		parsed, err := strconv.ParseInt(limitParam, 10, 64)
		if err == nil {
			limit = parsed
			log = log.WithField("parsedLimit", limit)
		} else {
			log.WithError(err).Warn("Invalid limit query param")
		}
	}

	goals, err := h.Service.GetAllGoals(r.Context(), limit)
	if err != nil {
		log.WithError(err).Error("Failed to fetch all goals")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.WithField("goalCount", len(goals)).Info("Successfully fetched all goals")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goals)
}

func (h *GoalHandler) GetGoalProgressHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	goalID := vars["id"]
	log := logrus.WithField("goalID", goalID)

	// Get the logged-in user
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		log.Warn("Unauthorized access")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch the goal from DB
	goal, err := h.Service.GetGoal(r.Context(), goalID)
	if err != nil || goal == nil {
		log.WithError(err).Warn("Goal not found")
		http.Error(w, "Goal not found", http.StatusNotFound)
		return
	}

	// Ensure the logged-in user is the owner of the goal
	if goal.UserID.Hex() != claims.UserID && !isCollaborator(goal.Collaborators, claims.UserID) {
		log.Warn("Forbidden: Not owner or collaborator")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Return steps and substeps as progress data
	response := map[string]interface{}{
		"steps": goal.Steps,
	}

	log.Info("Goal progress fetched successfully")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *GoalHandler) GetGoalsHandler(w http.ResponseWriter, r *http.Request) {
	// Get logged-in user
	claims := middleware.GetUserFromContext(r.Context())
	log := logrus.WithField("userID", claims.UserID)

	if claims == nil {
		log.Warn("Unauthorized access")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert UserID to ObjectID
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		log.WithError(err).Error("Invalid user ID format")
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	// Get category filter from query params (optional)
	category := r.URL.Query().Get("category")
	log = log.WithField("category", category)

	// Fetch goals from DB with optional category filter
	goals, err := h.Service.GetGoals(r.Context(), userID, category)
	if err != nil {
		log.WithError(err).Error("Failed to retrieve user goals")
		http.Error(w, "Failed to retrieve goals", http.StatusInternalServerError)
		return
	}

	log.WithField("goalCount", len(goals)).Info("User goals fetched successfully")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goals)
}

func (h *GoalHandler) InviteCollaboratorHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	goalID := vars["id"]

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		logger.Log.Warn("Unauthorized attempt to invite collaborator")
		return
	}

	requesterID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		logger.Log.Errorf("Invalid user ID format: %v", err)
		return
	}

	// Parse body to get collaboratorID
	var req struct {
		CollaboratorID string `json:"collaborator_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		logger.Log.Warn("Invalid request payload for collaborator invite")
		return
	}
	defer r.Body.Close()

	collaboratorID, err := primitive.ObjectIDFromHex(req.CollaboratorID)
	if err != nil {
		http.Error(w, "Invalid collaborator ID", http.StatusBadRequest)
		logger.Log.Warnf("Invalid collaborator ID: %v", err)
		return
	}

	err = h.Service.InviteCollaborator(r.Context(), goalID, requesterID, collaboratorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logger.Log.Warnf("Failed to invite collaborator: %v", err)
		return
	}

	goal, _ := h.Service.GetGoal(r.Context(), goalID)

	_ = h.ActivityService.LogActivity(r.Context(), requesterID, "collaborator_invited", goal.ID, fmt.Sprintf("Invited user %s to collaborate", collaboratorID))

	// Send notification to invited user
	_ = h.NotificationService.CreateNotification(
		r.Context(),
		collaboratorID,
		"collaborator_added",
		"You’ve been added to a goal",
		fmt.Sprintf("You’ve been invited to collaborate on: %s", goal.Name),
		&goal.ID,
	)

	logger.Log.Infof("User %s invited %s to collaborate on goal %s", claims.UserID, req.CollaboratorID, goalID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Collaborator successfully invited",
	})
}

func isCollaborator(collaborators []primitive.ObjectID, userID string) bool {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false
	}
	for _, c := range collaborators {
		if c == id {
			return true
		}
	}
	return false
}
