package utils

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	// InfoLogger logs information messages
	InfoLogger *log.Logger
	// ErrorLogger logs error messages
	ErrorLogger *log.Logger
	// AnalyticsLogger logs usage analytics
	AnalyticsLogger *log.Logger
)

// InitLoggers initializes the loggers
func InitLoggers(logDir, analyticsLogFile string) error {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// Create log files
	apiLogFile, err := os.OpenFile(filepath.Join(logDir, "api.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	errorLogFile, err := os.OpenFile(filepath.Join(logDir, "error.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	analyticsLog, err := os.OpenFile(analyticsLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	// Initialize loggers
	InfoLogger = log.New(apiLogFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(errorLogFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	AnalyticsLogger = log.New(analyticsLog, "ANALYTICS: ", log.Ldate|log.Ltime)

	return nil
}

// LogRequest logs an HTTP request
func LogRequest(r *http.Request) {
	if InfoLogger != nil {
		InfoLogger.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
	} else {
		log.Printf("INFO: %s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
	}
}

// LogInfo logs an information message
func LogInfo(format string, v ...interface{}) {
	if InfoLogger != nil {
		InfoLogger.Printf(format, v...)
	} else {
		log.Printf("INFO: "+format, v...)
	}
}

// LogError logs an error message
func LogError(format string, v ...interface{}) {
	if ErrorLogger != nil {
		ErrorLogger.Printf(format, v...)
	} else {
		log.Printf("ERROR: "+format, v...)
	}
}

// LogAnalytics logs usage analytics
func LogAnalytics(userID, action, details string) {
	timestamp := time.Now().Format(time.RFC3339)
	message := fmt.Sprintf("user=%s action=%s timestamp=%s details=%s", userID, action, timestamp, details)
	
	if AnalyticsLogger != nil {
		AnalyticsLogger.Println(message)
	} else {
		log.Printf("ANALYTICS: %s", message)
	}
}
