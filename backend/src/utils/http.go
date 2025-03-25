package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

// RespondWithError sends an error response
func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, map[string]string{"error": message})
}

// RespondWithJSON sends a JSON response
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	
	// Set content type
	w.Header().Set("Content-Type", "application/json")
	
	// Set status code
	w.WriteHeader(code)
	
	// Encode payload
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}
