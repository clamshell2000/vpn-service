package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

// WriteJSONToFile writes a JSON object to a file
func WriteJSONToFile(path string, data interface{}) error {
	// Create file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Encode data
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode data: %v", err)
	}

	return nil
}

// ReadJSONFromFile reads a JSON object from a file
func ReadJSONFromFile(path string, data interface{}) error {
	// Open file
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Decode data
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(data); err != nil {
		return fmt.Errorf("failed to decode data: %v", err)
	}

	return nil
}

// Base64Encode encodes data to base64
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
