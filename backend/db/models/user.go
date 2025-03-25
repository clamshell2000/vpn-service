package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password_hash"` // Password hash is not included in JSON
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// NewUser creates a new user
func NewUser(username, email, passwordHash string) *User {
	now := time.Now()
	return &User{
		ID:        generateUUID(),
		Username:  username,
		Email:     email,
		Password:  passwordHash,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// generateUUID generates a UUID
func generateUUID() string {
	// This is a placeholder. In a real implementation, use a proper UUID library
	return "user-" + time.Now().Format("20060102150405")
}
