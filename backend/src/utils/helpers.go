package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"time"

	"github.com/google/uuid"
)

// GenerateUUID generates a random UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// WriteJSONToFile writes JSON data to a file
func WriteJSONToFile(path string, data interface{}) error {
	// Marshal data to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// Write to file
	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// ReadJSONFromFile reads JSON data from a file
func ReadJSONFromFile(path string, data interface{}) error {
	// Read file
	jsonData, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// Unmarshal JSON
	if err := json.Unmarshal(jsonData, data); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return nil
}

// RandomBool returns a random boolean with the given probability of being true
func RandomBool(probability float64) bool {
	if probability <= 0 {
		return false
	}
	if probability >= 1 {
		return true
	}

	// Generate random number between 0 and 1
	n, err := rand.Int(rand.Reader, big.NewInt(100))
	if err != nil {
		// Default to true in case of error
		return true
	}

	return float64(n.Int64())/100.0 < probability
}

// FormatDuration formats a duration as a human-readable string
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	} else if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%d minutes, %d seconds", minutes, seconds)
	} else if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%d hours, %d minutes", hours, minutes)
	} else {
		days := int(d.Hours()) / 24
		hours := int(d.Hours()) % 24
		return fmt.Sprintf("%d days, %d hours", days, hours)
	}
}

// FormatBytes formats a byte count as a human-readable string
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	// Open source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer sourceFile.Close()

	// Create destination file
	destinationFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer destinationFile.Close()

	// Copy contents
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %v", err)
	}

	// Sync to ensure write
	if err := destinationFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %v", err)
	}

	return nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// CreateDirIfNotExists creates a directory if it doesn't exist
func CreateDirIfNotExists(path string) error {
	if !DirExists(path) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}
