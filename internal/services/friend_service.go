package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"github.com/Dias221467/Achievemenet_Manager/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FriendService handles business logic for managing friendships.
type FriendService struct {
	friendRepo *repository.FriendRepository
	userRepo   *repository.UserRepository
}

// NewFriendService creates a new FriendService.
func NewFriendService(friendRepo *repository.FriendRepository, userRepo *repository.UserRepository) *FriendService {
	return &FriendService{
		friendRepo: friendRepo,
		userRepo:   userRepo,
	}
}

// SendFriendRequest creates a new friend request.
func (s *FriendService) SendFriendRequest(ctx context.Context, senderID, receiverID primitive.ObjectID) (*models.FriendRequest, error) {
	if senderID == receiverID {
		return nil, fmt.Errorf("cannot send a friend request to yourself")
	}

	request := &models.FriendRequest{
		SenderID:   senderID,
		ReceiverID: receiverID,
		CreatedAt:  time.Now(),
		Status:     "pending",
	}

	return s.friendRepo.CreateRequest(ctx, request)
}

// GetPendingRequests fetches all pending requests for the receiver.
func (s *FriendService) GetPendingRequests(ctx context.Context, receiverID primitive.ObjectID) ([]models.FriendRequest, error) {
	return s.friendRepo.GetRequestsByReceiver(ctx, receiverID)
}

// RespondToRequest updates a friend request's status and updates user friend lists if accepted.
func (s *FriendService) RespondToRequest(ctx context.Context, requestID primitive.ObjectID, accept bool) error {
	request, err := s.friendRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("could not find request: %v", err)
	}

	if request.Status != "pending" {
		return fmt.Errorf("request already responded to")
	}

	status := "rejected"
	if accept {
		status = "accepted"
	}

	// Update the status of the request
	if err := s.friendRepo.UpdateRequestStatus(ctx, requestID, status); err != nil {
		return err
	}

	if accept {
		// Update both users' friend lists
		if err := s.userRepo.AddFriend(ctx, request.SenderID, request.ReceiverID); err != nil {
			return fmt.Errorf("failed to add friend to sender: %v", err)
		}
		if err := s.userRepo.AddFriend(ctx, request.ReceiverID, request.SenderID); err != nil {
			return fmt.Errorf("failed to add friend to receiver: %v", err)
		}
	}

	return nil
}

func (s *FriendService) GetFriends(ctx context.Context, userID primitive.ObjectID) ([]models.PublicUser, error) {
	friendIDs, err := s.userRepo.GetFriendIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get friend IDs: %v", err)
	}

	if len(friendIDs) == 0 {
		return []models.PublicUser{}, nil
	}

	users, err := s.userRepo.GetUsersByIDs(ctx, friendIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %v", err)
	}

	publicFriends := make([]models.PublicUser, 0, len(users))
	for _, user := range users {
		publicFriends = append(publicFriends, models.PublicUser{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		})
	}

	return publicFriends, nil
}

func (s *FriendService) RemoveFriend(ctx context.Context, userID, friendID primitive.ObjectID) error {
	return s.userRepo.RemoveFriend(ctx, userID, friendID)
}