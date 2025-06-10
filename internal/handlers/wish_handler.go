package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"github.com/Dias221467/Achievemenet_Manager/internal/services"
	"github.com/Dias221467/Achievemenet_Manager/pkg/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WishHandler struct {
	Service     *services.WishService
	GoalService *services.GoalService
}

func NewWishHandler(service *services.WishService, goalService *services.GoalService) *WishHandler {
	return &WishHandler{
		Service:     service,
		GoalService: goalService,
	}
}

// CreateWishHandler handles creation of a new wish
func (h *WishHandler) CreateWishHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var wish models.Wish
	if err := json.NewDecoder(r.Body).Decode(&wish); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}
	wish.UserID = userID
	wish.CreatedAt = time.Now()
	wish.UpdatedAt = time.Now()

	createdWish, err := h.Service.CreateWish(r.Context(), &wish)
	if err != nil {
		http.Error(w, "Failed to create wish", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdWish)
}

// GetWishByIDHandler retrieves a specific wish by ID
func (h *WishHandler) GetWishByIDHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	wishID := mux.Vars(r)["id"]

	wish, err := h.Service.GetWishByID(r.Context(), wishID)
	if err != nil {
		http.Error(w, "Wish not found", http.StatusNotFound)
		return
	}

	if wish.UserID.Hex() != claims.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wish)
}

// GetWishesHandler returns all wishes of a user
func (h *WishHandler) GetWishesHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	wishes, err := h.Service.GetWishesByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to fetch wishes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wishes)
}

// UpdateWishHandler updates a wish
func (h *WishHandler) UpdateWishHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	wishID := mux.Vars(r)["id"]

	wish, err := h.Service.GetWishByID(r.Context(), wishID)
	if err != nil || wish.UserID.Hex() != claims.UserID {
		http.Error(w, "Forbidden or not found", http.StatusForbidden)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid update payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := h.Service.UpdateWish(r.Context(), wishID, updates); err != nil {
		http.Error(w, "Failed to update wish", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Wish updated successfully"))
}

// DeleteWishHandler removes a wish
func (h *WishHandler) DeleteWishHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	wishID := mux.Vars(r)["id"]
	wish, err := h.Service.GetWishByID(r.Context(), wishID)
	if err != nil || wish.UserID.Hex() != claims.UserID {
		http.Error(w, "Forbidden or not found", http.StatusForbidden)
		return
	}

	if err := h.Service.DeleteWish(r.Context(), wishID); err != nil {
		http.Error(w, "Failed to delete wish", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Wish deleted successfully"))
}

// PromoteWishHandler transforms a wish into a goal
func (h *WishHandler) PromoteWishHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	wishID := vars["id"]

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert user ID
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	// Get the wish by ID
	wish, err := h.Service.GetWishByID(r.Context(), wishID)
	if err != nil {
		http.Error(w, "Wish not found", http.StatusNotFound)
		return
	}

	if wish.UserID != userID {
		http.Error(w, "Forbidden: You can only promote your own wish", http.StatusForbidden)
		return
	}

	// Construct a Goal from the Wish
	goal := &models.Goal{
		Name:        wish.Title,
		Description: wish.Description,
		UserID:      userID,
		Status:      "in_progress",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Optionally carry over any substeps if your wish had them (not implemented yet)
	createdGoal, err := h.GoalService.CreateGoal(r.Context(), goal)
	if err != nil {
		http.Error(w, "Failed to promote wish to goal", http.StatusInternalServerError)
		return
	}

	// Respond with the created goal
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdGoal)
}

// UploadWishImageHandler handles uploading an image for a specific wish.
func (h *WishHandler) UploadWishImageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	wishID := vars["id"]

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form (max size: 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "File too big or invalid format", http.StatusBadRequest)
		return
	}

	// Get the file from form-data
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Missing file in request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check content type
	contentType := header.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" {
		http.Error(w, "Only JPEG and PNG images are allowed", http.StatusBadRequest)
		return
	}

	// Generate unique file name
	ext := filepath.Ext(header.Filename)
	fileName := uuid.NewString() + ext
	savePath := filepath.Join("uploads", fileName)

	// Create folder if not exists
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		http.Error(w, "Failed to create upload folder", http.StatusInternalServerError)
		return
	}

	// Save file to disk
	out, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		return
	}

	// Build file URL (can be changed later to use full domain)
	fileURL := "/uploads/" + fileName

	logrus.WithFields(logrus.Fields{
		"wishID":  wishID,
		"userID":  claims.UserID,
		"fileURL": fileURL,
	}).Info("Attempting to update wish image")

	if err != nil {
		logrus.WithError(err).Error("UpdateWishImage failed")
		http.Error(w, "Failed to update wish with image", http.StatusInternalServerError)
		return
	}

	// Update Wish with image URL
	updated, err := h.Service.UpdateWishImage(r.Context(), wishID, claims.UserID, fileURL)
	if err != nil {
		http.Error(w, "Failed to update wish with image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}
