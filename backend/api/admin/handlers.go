package admin

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/vpn-service/backend/db/models"
	"github.com/vpn-service/backend/src/core"
	"github.com/vpn-service/backend/src/utils"
)

// UserManager is the user manager instance
var UserManager *core.UserManager

// UserResponse represents a user response
type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// UserUpdateRequest represents a user update request
type UserUpdateRequest struct {
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	Active   bool   `json:"active"`
}

// ListUsersHandler handles user listing requests
func ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Get users
	users, err := UserManager.GetAllUsers()
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get users")
		return
	}

	// Convert to response
	response := make([]UserResponse, len(users))
	for i, user := range users {
		response[i] = convertUserToResponse(user)
	}

	// Return users
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GetUserHandler handles user retrieval requests
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]

	// Get user
	user, err := UserManager.GetUser(userID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	// Return user
	utils.WriteJSONResponse(w, http.StatusOK, convertUserToResponse(user))
}

// UpdateUserHandler handles user update requests
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]

	// Get user
	user, err := UserManager.GetUser(userID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	// Parse request
	var req UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := validateUserUpdateRequest(req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update user
	if req.Email != "" {
		user, err = UserManager.UpdateUser(userID, req.Email)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to update user")
			return
		}
	}

	// Update password if provided
	if req.Password != "" {
		if err := UserManager.SetUserPassword(userID, req.Password); err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to update password")
			return
		}
	}

	// Return user
	utils.WriteJSONResponse(w, http.StatusOK, convertUserToResponse(user))
}

// DeleteUserHandler handles user deletion requests
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]

	// Delete user
	if err := UserManager.DeleteUser(userID); err != nil {
		utils.WriteErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	// Return success
	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// GetUserPeersHandler handles user peers retrieval requests
func GetUserPeersHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]

	// Get user peers
	peers, err := UserManager.GetUserPeers(userID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get user peers")
		return
	}

	// Return peers
	utils.WriteJSONResponse(w, http.StatusOK, peers)
}

// DeleteUserPeerHandler handles user peer deletion requests
func DeleteUserPeerHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID and peer ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]
	peerID := vars["peerID"]

	// Delete peer
	if err := UserManager.DeleteUserPeer(userID, peerID); err != nil {
		utils.WriteErrorResponse(w, http.StatusNotFound, "Peer not found")
		return
	}

	// Return success
	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// convertUserToResponse converts a user model to a response
func convertUserToResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// validateUserUpdateRequest validates a user update request
func validateUserUpdateRequest(req UserUpdateRequest) error {
	// Validate email if provided
	if req.Email != "" && !utils.IsValidEmail(req.Email) {
		return utils.NewError("invalid email")
	}

	// Validate password if provided
	if req.Password != "" && len(req.Password) < 8 {
		return utils.NewError("password must be at least 8 characters")
	}

	return nil
}
