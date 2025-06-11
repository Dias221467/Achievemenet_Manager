package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Dias221467/Achievemenet_Manager/internal/services"
	"github.com/Dias221467/Achievemenet_Manager/pkg/logger"
	"github.com/Dias221467/Achievemenet_Manager/pkg/middleware"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationHandler struct {
	Service *services.NotificationService
}

func NewNotificationHandler(service *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{Service: service}
}

// GET /notifications
func (h *NotificationHandler) GetUserNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, _ := primitive.ObjectIDFromHex(claims.UserID)
	notifications, err := h.Service.GetUserNotifications(r.Context(), userID)
	if err != nil {
		logger.Log.Errorf("Failed to fetch notifications: %v", err)
		http.Error(w, "Failed to get notifications", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(notifications)
}

// POST /notifications/{id}/read
func (h *NotificationHandler) MarkAsReadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	notifID, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	if err := h.Service.MarkNotificationAsRead(r.Context(), notifID); err != nil {
		logger.Log.Errorf("Failed to mark notification as read: %v", err)
		http.Error(w, "Failed to mark as read", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Notification marked as read"})
}

// DELETE /notifications/{id}
func (h *NotificationHandler) DeleteNotificationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	notifID, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	if err := h.Service.DeleteNotification(r.Context(), notifID); err != nil {
		logger.Log.Errorf("Failed to delete notification: %v", err)
		http.Error(w, "Failed to delete notification", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Notification deleted"})
}
