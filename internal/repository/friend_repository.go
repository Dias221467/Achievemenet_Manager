package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FriendRepository struct {
	collection *mongo.Collection
}

func NewFriendRepository(db *mongo.Database) *FriendRepository {
	return &FriendRepository{
		collection: db.Collection("friend_requests"),
	}
}

func (r *FriendRepository) CreateRequest(ctx context.Context, req *models.FriendRequest) (*models.FriendRequest, error) {
	req.CreatedAt = time.Now()
	req.Status = "pending"

	result, err := r.collection.InsertOne(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to send friend request: %v", err)
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("failed to cast inserted ID")
	}
	req.ID = insertedID

	return req, nil
}

func (r *FriendRepository) GetRequestsByReceiver(ctx context.Context, receiverID primitive.ObjectID) ([]models.FriendRequest, error) {
	filter := bson.M{"receiver_id": receiverID, "status": "pending"}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find friend requests: %v", err)
	}
	defer cursor.Close(ctx)

	var requests []models.FriendRequest
	for cursor.Next(ctx) {
		var req models.FriendRequest
		if err := cursor.Decode(&req); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, nil
}

func (r *FriendRepository) UpdateRequestStatus(ctx context.Context, id primitive.ObjectID, status string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": status}},
	)
	if err != nil {
		return fmt.Errorf("failed to update request status: %v", err)
	}
	return nil
}

func (r *FriendRepository) GetFriends(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"sender_id": userID, "status": "accepted"},
			{"receiver_id": userID, "status": "accepted"},
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve friends: %v", err)
	}
	defer cursor.Close(ctx)

	var friends []primitive.ObjectID
	for cursor.Next(ctx) {
		var req models.FriendRequest
		if err := cursor.Decode(&req); err != nil {
			return nil, err
		}

		if req.SenderID == userID {
			friends = append(friends, req.ReceiverID)
		} else {
			friends = append(friends, req.SenderID)
		}
	}

	return friends, nil
}

func (r *FriendRepository) GetRequestByID(ctx context.Context, id primitive.ObjectID) (*models.FriendRequest, error) {
	var request models.FriendRequest
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&request)
	if err != nil {
		return nil, fmt.Errorf("failed to find friend request: %v", err)
	}
	return &request, nil
}
