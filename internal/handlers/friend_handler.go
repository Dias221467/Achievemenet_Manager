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

// FriendHandler manages HTTP endpoints related to friend requests.
type FriendHandler struct {
	Service *services.FriendService
}

// NewFriendHandler initializes a new FriendHandler.
func NewFriendHandler(service *services.FriendService) *FriendHandler {
	return &FriendHandler{Service: service}
}

// SendFriendRequestHandler allows a user to send a friend request.
func (h *FriendHandler) SendFriendRequestHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		logger.Log.Warn("Unauthorized attempt to send friend request")
		return
	}

	vars := mux.Vars(r)
	receiverIDHex := vars["id"]
	receiverID, err := primitive.ObjectIDFromHex(receiverIDHex)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		logger.Log.Warnf("Invalid receiver ID: %v", err)
		return
	}

	senderID, _ := primitive.ObjectIDFromHex(claims.UserID)

	request, err := h.Service.SendFriendRequest(r.Context(), senderID, receiverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logger.Log.Warnf("Failed to send friend request: %v", err)
		return
	}

	logger.Log.Infof("User %s sent a friend request to %s", claims.UserID, receiverIDHex)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(request)
}

// GetPendingRequestsHandler shows all incoming friend requests.
func (h *FriendHandler) GetPendingRequestsHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		logger.Log.Warn("Unauthorized attempt to get pending requests")
		return
	}

	userID, _ := primitive.ObjectIDFromHex(claims.UserID)

	requests, err := h.Service.GetPendingRequests(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get requests", http.StatusInternalServerError)
		logger.Log.Errorf("Failed to get pending requests: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

// RespondToFriendRequestHandler allows accepting or rejecting a friend request.
func (h *FriendHandler) RespondToFriendRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestIDHex := vars["id"]

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		logger.Log.Warn("Unauthorized request to respond to a friend request")
		return
	}

	requestID, err := primitive.ObjectIDFromHex(requestIDHex)
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		logger.Log.Warnf("Invalid friend request ID: %v", err)
		return
	}

	var body struct {
		Accept bool `json:"accept"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		logger.Log.Warnf("Failed to decode response body: %v", err)
		return
	}
	defer r.Body.Close()

	// Handle the friend request response
	err = h.Service.RespondToRequest(r.Context(), requestID, body.Accept)
	if err != nil {
		http.Error(w, "Failed to respond to request", http.StatusInternalServerError)
		logger.Log.Errorf("Failed to respond to friend request %s: %v", requestIDHex, err)
		return
	}

	logger.Log.Infof("User %s responded to friend request %s (accepted: %v)", claims.UserID, requestIDHex, body.Accept)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Friend request response recorded",
	})
}

// GetFriendsHandler returns a list of user’s friends.
func (h *FriendHandler) GetFriendsHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		logger.Log.Warn("Unauthorized attempt to get friends")
		return
	}

	userID, _ := primitive.ObjectIDFromHex(claims.UserID)

	friends, err := h.Service.GetFriends(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get friends", http.StatusInternalServerError)
		logger.Log.Errorf("Failed to fetch friends for user %s: %v", claims.UserID, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(friends)
}
