package services

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"github.com/Dias221467/Achievemenet_Manager/internal/repository"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserService encapsulates the business logic for user operations.
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService creates a new instance of UserService.
func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// RegisterUser registers a new user after hashing their password.
func (s *UserService) RegisterUser(ctx context.Context, user *models.User) (*models.User, error) {
	logrus.Info("Registering new user")

	if user.Email == "" || user.Username == "" || user.HashedPassword == "" {
		logrus.Warn("Missing required fields during registration")
		return nil, fmt.Errorf("missing required user fields")
	}

	// Check if the email is already registered
	existingUser, _ := s.repo.GetUserByEmail(ctx, user.Email)
	if existingUser != nil {
		logrus.WithField("email", user.Email).Warn("Email already in use")
		return nil, fmt.Errorf("email already in use")
	}

	// Hash the user's password.
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(user.HashedPassword), bcrypt.DefaultCost)
	if err != nil {
		logrus.WithError(err).Error("Password hashing failed")
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	user.HashedPassword = string(hashedPwd)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Create the user in the repository.
	createdUser, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		logrus.WithError(err).Error("User registration failed")
		return nil, fmt.Errorf("failed to register user: %v", err)
	}

	logrus.WithField("userID", createdUser.ID.Hex()).Info("User registered successfully")
	return createdUser, nil
}

// AuthenticateUser verifies the email and password and returns the user if credentials are valid.
func (s *UserService) AuthenticateUser(ctx context.Context, email, password string) (*models.User, error) {
	logrus.WithField("email", email).Info("Authenticating user")

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		logrus.WithField("email", email).Warn("User not found")
		return nil, fmt.Errorf("user not found")
	}

	// Compare the provided password with the hashed password.
	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
		logrus.WithField("email", email).Warn("Invalid credentials")
		return nil, fmt.Errorf("invalid credentials")
	}

	logrus.WithField("userID", user.ID.Hex()).Info("User authenticated successfully")
	return user, nil
}

// GetUser retrieves a user by their ID.
func (s *UserService) GetUser(ctx context.Context, id string) (*models.User, error) {
	logrus.WithField("userID", id).Info("Fetching user")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logrus.WithError(err).Warn("Invalid user ID")
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	user, err := s.repo.GetUserByID(ctx, objID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to retrieve user")
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	logrus.WithField("userID", user.ID.Hex()).Info("User retrieved successfully")
	return user, nil
}

// UpdateUser updates an existing user's details.
func (s *UserService) UpdateUser(ctx context.Context, id string, updatedUser *models.User) (*models.User, error) {
	logrus.WithField("userID", id).Info("Updating user")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logrus.WithError(err).Warn("Invalid user ID")
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	updatedUser.UpdatedAt = time.Now()

	user, err := s.repo.UpdateUser(ctx, objID, updatedUser)
	if err != nil {
		logrus.WithError(err).Error("Failed to update user")
		return nil, fmt.Errorf("failed to update user: %v", err)
	}

	logrus.WithField("userID", user.ID.Hex()).Info("User updated successfully")
	return user, nil
}

// DeleteUser deletes a user by their ID.
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	logrus.WithField("userID", id).Info("Deleting user")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logrus.WithError(err).Warn("Invalid user ID")
		return fmt.Errorf("invalid user ID: %v", err)
	}

	if err := s.repo.DeleteUser(ctx, objID); err != nil {
		logrus.WithError(err).Error("Failed to delete user")
		return err
	}

	logrus.WithField("userID", id).Info("User deleted successfully")
	return nil
}
