package repository

import (
	"context"
	"time"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ChatRepository struct {
	collection *mongo.Collection
}

func NewChatRepository(db *mongo.Database) *ChatRepository {
	return &ChatRepository{collection: db.Collection("messages")}
}

func (r *ChatRepository) SendMessage(ctx context.Context, msg *models.Message) (*models.Message, error) {
	msg.CreatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, msg)
	if err != nil {
		return nil, err
	}
	msg.ID = result.InsertedID.(primitive.ObjectID)
	return msg, nil
}

func (r *ChatRepository) GetChat(ctx context.Context, userID, friendID primitive.ObjectID) ([]models.Message, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"sender_id": userID, "receiver_id": friendID},
			{"sender_id": friendID, "receiver_id": userID},
		},
	}
	opts := options.Find().SetSort(bson.D{{"created_at", 1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []models.Message
	for cursor.Next(ctx) {
		var msg models.Message
		if err := cursor.Decode(&msg); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
