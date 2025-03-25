package core

import (
	"fmt"
	"time"

	"github.com/vpn-service/backend/db"
	"github.com/vpn-service/backend/db/models"
	"github.com/vpn-service/backend/src/config"
	"github.com/vpn-service/backend/src/utils"
	"github.com/vpn-service/backend/vpn/wireguard"
	"golang.org/x/crypto/bcrypt"
)

// UserManager manages user operations
type UserManager struct {
	config *config.Config
}

// NewUserManager creates a new user manager
func NewUserManager(cfg *config.Config) *UserManager {
	return &UserManager{
		config: cfg,
	}
}

// RegisterUser registers a new user
func (um *UserManager) RegisterUser(username, email, password string) (*models.User, error) {
	// Check if user already exists
	exists, err := um.userExists(username, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check if user exists: %v", err)
	}
	if exists {
		return nil, fmt.Errorf("user already exists")
	}

	// Hash password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Create user
	user := models.NewUser(username, email, hashedPassword)

	// Save user to database
	if err := um.saveUser(user); err != nil {
		return nil, fmt.Errorf("failed to save user: %v", err)
	}

	// Log analytics
	utils.LogAnalytics(user.ID, "user_register", fmt.Sprintf("username=%s email=%s", username, email))

	return user, nil
}

// AuthenticateUser authenticates a user
func (um *UserManager) AuthenticateUser(username, password string) (*models.User, error) {
	// Get user from database
	user, err := um.getUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	// Verify password
	if err := verifyPassword(password, user.Password); err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	// Log analytics
	utils.LogAnalytics(user.ID, "user_login", fmt.Sprintf("username=%s", username))

	return user, nil
}

// GetUser gets a user by ID
func (um *UserManager) GetUser(id string) (*models.User, error) {
	// Get user from database
	user, err := um.getUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return user, nil
}

// UpdateUser updates a user
func (um *UserManager) UpdateUser(id, email string) (*models.User, error) {
	// Get user from database
	user, err := um.getUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	// Update user
	user.Email = email
	user.UpdatedAt = time.Now()

	// Save user to database
	if err := um.saveUser(user); err != nil {
		return nil, fmt.Errorf("failed to save user: %v", err)
	}

	// Log analytics
	utils.LogAnalytics(user.ID, "user_update", fmt.Sprintf("email=%s", email))

	return user, nil
}

// ChangePassword changes a user's password
func (um *UserManager) ChangePassword(id, oldPassword, newPassword string) error {
	// Get user from database
	user, err := um.getUserByID(id)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}

	// Verify old password
	if err := verifyPassword(oldPassword, user.Password); err != nil {
		return fmt.Errorf("invalid password")
	}

	// Hash new password
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Update user
	user.Password = hashedPassword
	user.UpdatedAt = time.Now()

	// Save user to database
	if err := um.saveUser(user); err != nil {
		return fmt.Errorf("failed to save user: %v", err)
	}

	// Log analytics
	utils.LogAnalytics(user.ID, "user_change_password", "")

	return nil
}

// GetAllUsers gets all users
func (um *UserManager) GetAllUsers() ([]*models.User, error) {
	// In a real implementation, this would query the database
	// For now, we'll just return some mock users
	users := []*models.User{
		{
			ID:        "user-123",
			Username:  "user1",
			Email:     "user1@example.com",
			Password:  "$2a$10$1234567890123456789012345678901234567890123456789012345678901234",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-12 * time.Hour),
		},
		{
			ID:        "user-456",
			Username:  "user2",
			Email:     "user2@example.com",
			Password:  "$2a$10$1234567890123456789012345678901234567890123456789012345678901234",
			CreatedAt: time.Now().Add(-48 * time.Hour),
			UpdatedAt: time.Now().Add(-24 * time.Hour),
		},
	}

	return users, nil
}

// DeleteUser deletes a user
func (um *UserManager) DeleteUser(id string) error {
	// In a real implementation, this would delete the user from the database
	// For now, we'll just log it
	utils.LogInfo("Deleting user: %s", id)
	return nil
}

// SetUserPassword sets a user's password
func (um *UserManager) SetUserPassword(id, password string) error {
	// Get user from database
	user, err := um.getUserByID(id)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}

	// Hash password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Update user
	user.Password = hashedPassword
	user.UpdatedAt = time.Now()

	// Save user to database
	if err := um.saveUser(user); err != nil {
		return fmt.Errorf("failed to save user: %v", err)
	}

	// Log analytics
	utils.LogAnalytics(user.ID, "user_password_reset", "")

	return nil
}

// GetUserPeers gets a user's VPN peers
func (um *UserManager) GetUserPeers(id string) ([]*wireguard.PeerConfig, error) {
	// In a real implementation, this would query the database
	// For now, we'll just return some mock peers
	peers := []*wireguard.PeerConfig{
		{
			ID:        "peer-123",
			UserID:    id,
			ServerID:  "server-1",
			DeviceType: "android",
			PublicKey:  "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG=",
			IP:         "10.0.0.2/32",
			CreatedAt:  time.Now().Add(-24 * time.Hour),
		},
		{
			ID:        "peer-456",
			UserID:    id,
			ServerID:  "server-2",
			DeviceType: "ios",
			PublicKey:  "HIJKLMNOPQRSTUVWXYZ0123456789ABCDEFGabcdefg=",
			IP:         "10.0.0.3/32",
			CreatedAt:  time.Now().Add(-12 * time.Hour),
		},
	}

	return peers, nil
}

// DeleteUserPeer deletes a user's VPN peer
func (um *UserManager) DeleteUserPeer(userID, peerID string) error {
	// In a real implementation, this would delete the peer from the database
	// For now, we'll just log it
	utils.LogInfo("Deleting peer %s for user %s", peerID, userID)
	return nil
}

// userExists checks if a user already exists
func (um *UserManager) userExists(username, email string) (bool, error) {
	// In a real implementation, this would check the database
	// For now, we'll just return false
	return false, nil
}

// getUserByUsername gets a user by username
func (um *UserManager) getUserByUsername(username string) (*models.User, error) {
	// In a real implementation, this would query the database
	// For now, we'll just return a mock user
	return &models.User{
		ID:        "user-123",
		Username:  username,
		Email:     "user@example.com",
		Password:  "$2a$10$1234567890123456789012345678901234567890123456789012345678901234",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// getUserByID gets a user by ID
func (um *UserManager) getUserByID(id string) (*models.User, error) {
	// In a real implementation, this would query the database
	// For now, we'll just return a mock user
	return &models.User{
		ID:        id,
		Username:  "user",
		Email:     "user@example.com",
		Password:  "$2a$10$1234567890123456789012345678901234567890123456789012345678901234",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// saveUser saves a user to the database
func (um *UserManager) saveUser(user *models.User) error {
	// In a real implementation, this would save to the database
	// For now, we'll just log it
	utils.LogInfo("Saving user: %s", user.ID)
	return nil
}

// hashPassword hashes a password
func hashPassword(password string) (string, error) {
	// Generate bcrypt hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// verifyPassword verifies a password
func verifyPassword(password, hash string) error {
	// Compare password with hash
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
