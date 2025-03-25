package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/vpn-service/backend/src/config"
	"github.com/vpn-service/backend/src/utils"
)

// RegisterRoutes registers the auth routes
func RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/register", RegisterHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/login", LoginHandler).Methods("POST", "OPTIONS")
}

// User represents a user in the system
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email"`
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// RegisterHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate request
	if req.Username == "" || req.Password == "" || req.Email == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Username, password, and email are required")
		return
	}

	// TODO: Check if user already exists
	// TODO: Hash password
	// TODO: Save user to database

	// Create user (mock implementation)
	user := User{
		ID:       utils.GenerateUUID(),
		Username: req.Username,
		Email:    req.Email,
	}

	// Generate token
	token, err := generateToken(user.ID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	// Respond with token and user
	utils.RespondWithJSON(w, http.StatusCreated, AuthResponse{
		Token: token,
		User:  user,
	})
}

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate request
	if req.Username == "" || req.Password == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	// TODO: Authenticate user against database
	// TODO: Verify password hash

	// Mock user authentication (replace with actual database lookup)
	user := User{
		ID:       "user-123",
		Username: req.Username,
		Email:    "user@example.com",
	}

	// Generate token
	token, err := generateToken(user.ID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	// Respond with token and user
	utils.RespondWithJSON(w, http.StatusOK, AuthResponse{
		Token: token,
		User:  user,
	})
}

// generateToken generates a JWT token for the given user ID
func generateToken(userID string) (string, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return "", err
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  userID,
		"exp": time.Now().Add(time.Hour * time.Duration(cfg.JWT.Expiration)).Unix(),
	})

	// Sign token
	return token.SignedString([]byte(cfg.JWT.Secret))
}
