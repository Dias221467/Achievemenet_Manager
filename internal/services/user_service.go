package services

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Dias221467/Achievemenet_Manager/internal/models"
	"github.com/Dias221467/Achievemenet_Manager/internal/repository"
	"github.com/Dias221467/Achievemenet_Manager/pkg/email"
	"github.com/google/uuid"
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

	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	if user.Email == "" || user.Username == "" || user.HashedPassword == "" {
		logrus.Warn("Missing required fields during registration")
		return nil, fmt.Errorf("missing required user fields")
	}

	if !emailRegex.MatchString(user.Email) {
		logrus.WithField("email", user.Email).Warn("Invalid email format during registration")
		return nil, fmt.Errorf("invalid email format")
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

	if user.Role == "" {
		user.Role = "user"
	}

	verificationToken := uuid.NewString()
	user.VerifyToken = verificationToken
	user.IsVerified = false

	// Create the user in the repository.
	createdUser, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		logrus.WithError(err).Error("User registration failed")
		return nil, fmt.Errorf("failed to register user: %v", err)
	}

	verificationLink := fmt.Sprintf("http://localhost:8080/users/verify?token=%s", verificationToken)

	emailBody := fmt.Sprintf("Welcome to Achievement Manager!\n\nPlease verify your email by clicking the link below:\n%s", verificationLink)

	err = email.SendEmail(user.Email, "Email Verification", emailBody)
	if err != nil {
		logrus.WithError(err).Error("Failed to send verification email")
		return nil, fmt.Errorf("failed to send verification email")
	}

	logrus.Infof("Sent verification email to %s", user.Email)

	logrus.WithFields(logrus.Fields{
		"userID": createdUser.ID.Hex(),
		"role":   createdUser.Role,
	}).Info("User registered successfully")

	return createdUser, nil
}

func (s *UserService) VerifyEmail(ctx context.Context, token string) error {
	// Look up user by the verification token
	user, err := s.repo.GetUserByVerificationToken(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid or expired verification token")
	}

	// Only update relevant fields
	update := map[string]interface{}{
		"is_verified":  true,
		"verify_token": "",
		"updated_at":   time.Now(),
	}

	_, err = s.repo.UpdateUser(ctx, user.ID, update)
	if err != nil {
		return fmt.Errorf("failed to update user verification status: %v", err)
	}

	return nil
}

func (s *UserService) RequestPasswordReset(ctx context.Context, userEmail string) error {
	user, err := s.repo.GetUserByEmail(ctx, userEmail)
	if err != nil {
		return fmt.Errorf("no account found with this email")
	}

	resetToken := uuid.NewString()
	expiration := time.Now().Add(1 * time.Hour)

	update := map[string]interface{}{
		"reset_token":     resetToken,
		"reset_token_exp": expiration,
		"updated_at":      time.Now(),
	}

	_, err = s.repo.UpdateUser(ctx, user.ID, update)
	if err != nil {
		return fmt.Errorf("failed to save reset token")
	}

	resetLink := fmt.Sprintf("http://localhost:8080/users/reset-password?token=%s", resetToken)
	body := fmt.Sprintf("Click the link below to reset your password:\n\n%s", resetLink)

	if err := email.SendEmail(user.Email, "Reset Your Password", body); err != nil {
		return fmt.Errorf("failed to send password reset email: %v", err)
	}

	logrus.Infof("Password reset email sent to %s", userEmail)
	return nil
}

func (s *UserService) ResetPassword(ctx context.Context, token, newPassword string) error {
	user, err := s.repo.GetUserByResetToken(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid or expired reset token")
	}

	if time.Now().After(user.ResetTokenExp) {
		return fmt.Errorf("reset token has expired")
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	update := map[string]interface{}{
		"hashed_password": string(hashedPwd),
		"reset_token":     "",
		"reset_token_exp": time.Time{},
		"updated_at":      time.Now(),
	}

	_, err = s.repo.UpdateUser(ctx, user.ID, update)
	if err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	return nil
}

// AuthenticateUser verifies the email and password and returns the user if credentials are valid.
func (s *UserService) AuthenticateUser(ctx context.Context, email, password string) (*models.User, error) {
	logrus.WithField("email", email).Info("Authenticating user")

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		logrus.WithField("email", email).Warn("User not found")
		return nil, fmt.Errorf("user not found")
	}

	// Email verification check
	if !user.IsVerified {
		logrus.WithField("email", email).Warn("Attempt to login with unverified email")
		return nil, fmt.Errorf("email not verified. Please check your inbox")
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
func (s *UserService) UpdateUser(ctx context.Context, id string, updatedUser map[string]interface{}) (*models.User, error) {
	logrus.WithField("userID", id).Info("Updating user")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logrus.WithError(err).Warn("Invalid user ID")
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	updatedUser["updated_at"] = time.Now()

	user, err := s.repo.UpdateUser(ctx, objID, updatedUser)
	if err != nil {
		logrus.WithError(err).Error("Failed to update user in service")
		return nil, fmt.Errorf("failed to update user: %v", err)
	}

	logrus.WithField("userID", user.ID.Hex()).Info("User updated successfully in service")
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

func (s *UserService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	return s.repo.GetAllUsers(ctx)
}
