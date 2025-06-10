package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"github.com/Dias221467/Achievemenet_Manager/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WishService struct {
	repo     *repository.WishRepository
	goalRepo *repository.GoalRepository
}

func NewWishService(repo *repository.WishRepository, goalRepo *repository.GoalRepository) *WishService {
	return &WishService{
		repo:     repo,
		goalRepo: goalRepo,
	}
}

func (s *WishService) CreateWish(ctx context.Context, wish *models.Wish) (*models.Wish, error) {
	if wish.Title == "" {
		return nil, fmt.Errorf("wish must have a title")
	}
	return s.repo.CreateWish(ctx, wish)
}

func (s *WishService) GetWishByID(ctx context.Context, id string) (*models.Wish, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid wish ID")
	}
	return s.repo.GetWishByID(ctx, objID)
}

func (s *WishService) GetWishesByUser(ctx context.Context, userID primitive.ObjectID) ([]models.Wish, error) {
	return s.repo.GetWishesByUser(ctx, userID)
}

func (s *WishService) UpdateWish(ctx context.Context, id string, updates map[string]interface{}) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid wish ID")
	}
	return s.repo.UpdateWish(ctx, objID, updates)
}

func (s *WishService) DeleteWish(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid wish ID")
	}
	return s.repo.DeleteWish(ctx, objID)
}

func (s *WishService) PromoteWishToGoal(ctx context.Context, id string, userID primitive.ObjectID) (*models.Goal, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid wish ID")
	}

	wish, err := s.repo.GetWishByID(ctx, objID)
	if err != nil {
		return nil, fmt.Errorf("wish not found")
	}

	goal := &models.Goal{
		Name:          wish.Title,
		Description:   wish.Description,
		UserID:        userID,
		Steps:         []models.Step{},
		Collaborators: []primitive.ObjectID{},
		Status:        "in_progress",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	return s.goalRepo.CreateGoal(ctx, goal)
}

func (s *WishService) UpdateWishImage(ctx context.Context, wishID string, userID string, imageURL string) (*models.Wish, error) {
	objID, err := primitive.ObjectIDFromHex(wishID)
	if err != nil {
		return nil, fmt.Errorf("invalid wish ID")
	}

	ownerID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	// Ensure the wish exists and belongs to the user
	wish, err := s.repo.GetWishByID(ctx, objID)
	if err != nil {
		return nil, fmt.Errorf("wish not found: %v", err)
	}
	if wish.UserID != ownerID {
		return nil, fmt.Errorf("forbidden: cannot update someone else's wish")
	}

	// Append the new image URL to existing list
	updatedImages := append(wish.Images, imageURL)

	update := map[string]interface{}{
		"images":     updatedImages,
		"updated_at": time.Now(),
	}

	updatedWish, err := s.repo.UpdateWishAndReturn(ctx, objID, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update wish with image: %v", err)
	}

	return updatedWish, nil
}
