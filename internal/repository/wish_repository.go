package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type WishRepository struct {
	collection *mongo.Collection
}

func NewWishRepository(db *mongo.Database) *WishRepository {
	return &WishRepository{collection: db.Collection("wishes")}
}

func (r *WishRepository) CreateWish(ctx context.Context, wish *models.Wish) (*models.Wish, error) {
	wish.CreatedAt = time.Now()
	wish.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, wish)
	if err != nil {
		return nil, fmt.Errorf("failed to create wish: %v", err)
	}

	wish.ID = result.InsertedID.(primitive.ObjectID)
	return wish, nil
}

func (r *WishRepository) GetWishByID(ctx context.Context, id primitive.ObjectID) (*models.Wish, error) {
	var wish models.Wish
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&wish); err != nil {
		return nil, fmt.Errorf("failed to get wish: %v", err)
	}
	return &wish, nil
}

func (r *WishRepository) GetWishesByUser(ctx context.Context, userID primitive.ObjectID) ([]models.Wish, error) {
	var wishes []models.Wish
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to get wishes: %v", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var wish models.Wish
		if err := cursor.Decode(&wish); err != nil {
			return nil, err
		}
		wishes = append(wishes, wish)
	}

	return wishes, nil
}

func (r *WishRepository) UpdateWish(ctx context.Context, id primitive.ObjectID, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": updates})
	if err != nil {
		return fmt.Errorf("failed to update wish: %v", err)
	}
	return nil
}

func (r *WishRepository) UpdateWishAndReturn(ctx context.Context, id primitive.ObjectID, updates map[string]interface{}) (*models.Wish, error) {
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedWish models.Wish
	err := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": updates}, opts).Decode(&updatedWish)
	if err != nil {
		logrus.WithError(err).Error("Failed to update wish and return updated object")
		return nil, fmt.Errorf("failed to update wish: %v", err)
	}

	return &updatedWish, nil
}

func (r *WishRepository) DeleteWish(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete wish: %v", err)
	}
	return nil
}
