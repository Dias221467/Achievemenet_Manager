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
)

// UserRepository handles database operations related to users.
type UserRepository struct {
	collection *mongo.Collection
}

// NewUserRepository creates a new instance of UserRepository.
func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

// CreateUser inserts a new user into the database.
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		logrus.WithError(err).Error("Failed to insert user into database")
		return nil, fmt.Errorf("failed to insert user: %v", err)
	}

	// Convert the inserted ID to primitive.ObjectID and assign it.
	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		logrus.Error("Failed to cast inserted ID to ObjectID")
		return nil, fmt.Errorf("failed to cast inserted ID")
	}

	user.ID = insertedID

	logrus.WithField("userID", user.ID.Hex()).Info("User inserted successfully")
	return user, nil
}

// GetUserByEmail retrieves a user by email.
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"email": email,
			"error": err,
		}).Warn("Failed to find user by email")
		return nil, fmt.Errorf("failed to find user by email: %v", err)
	}

	logrus.WithField("userID", user.ID.Hex()).Info("User found by email")
	return &user, nil
}

// GetUserByID retrieves a user by their ID.
func (r *UserRepository) GetUserByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"userID": id.Hex(),
			"error":  err,
		}).Warn("Failed to find user by ID")
		return nil, fmt.Errorf("failed to find user by id: %v", err)
	}

	logrus.WithField("userID", user.ID.Hex()).Info("User found by ID")
	return &user, nil
}

// UpdateUser updates an existing user's details.
func (r *UserRepository) UpdateUser(ctx context.Context, id primitive.ObjectID, user *models.User) (*models.User, error) {
	user.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": user})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"userID": id.Hex(),
			"error":  err,
		}).Error("Failed to update user")
		return nil, fmt.Errorf("failed to update user: %v", err)
	}

	logrus.WithField("userID", id.Hex()).Info("User updated successfully")
	return user, nil
}

// DeleteUser deletes a user from the database.
func (r *UserRepository) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"userID": id.Hex(),
			"error":  err,
		}).Error("Failed to delete user")
		return fmt.Errorf("failed to delete user: %v", err)
	}

	logrus.WithField("userID", id.Hex()).Info("User deleted successfully")
	return nil
}

func (r *UserRepository) AddFriend(ctx context.Context, userID, friendID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$addToSet": bson.M{"friends": friendID}}, // avoid duplicates
	)
	if err != nil {
		return fmt.Errorf("failed to add friend: %v", err)
	}
	return nil
}

// GetFriendIDs returns the list of friends for a user
func (r *UserRepository) GetFriendIDs(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user for friend list: %v", err)
	}
	return user.Friends, nil
}
