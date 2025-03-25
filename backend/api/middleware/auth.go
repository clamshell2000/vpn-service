package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/vpn-service/backend/src/config"
	"github.com/vpn-service/backend/src/utils"
)

// JWTAuthMiddleware authenticates requests using JWT
func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.RespondWithError(w, http.StatusUnauthorized, "Authorization header is required")
			return
		}

		// Check if the Authorization header has the correct format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.RespondWithError(w, http.StatusUnauthorized, "Authorization header must be in the format: Bearer {token}")
			return
		}

		// Parse and validate token
		tokenString := parts[1]
		userID, err := validateToken(tokenString)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Add user ID to request context
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware logs all requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log request
		utils.LogRequest(r)
		
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// validateToken validates a JWT token and returns the user ID
func validateToken(tokenString string) (string, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return "", err
	}

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.NewValidationError("invalid signing method", jwt.ValidationErrorSignatureInvalid)
		}
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		return "", err
	}

	// Validate token
	if !token.Valid {
		return "", jwt.NewValidationError("invalid token", jwt.ValidationErrorSignatureInvalid)
	}

	// Get claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", jwt.NewValidationError("invalid claims", jwt.ValidationErrorClaimsInvalid)
	}

	// Get user ID
	userID, ok := claims["id"].(string)
	if !ok {
		return "", jwt.NewValidationError("invalid user ID", jwt.ValidationErrorClaimsInvalid)
	}

	return userID, nil
}
