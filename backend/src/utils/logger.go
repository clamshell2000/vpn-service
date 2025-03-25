package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Logger is the global logger instance
	Logger *zap.Logger

	// SugaredLogger is the global sugared logger instance
	SugaredLogger *zap.SugaredLogger

	// analyticsLogger is used for logging analytics events
	analyticsLogger *zap.Logger
)

// InitLogger initializes the logger
func InitLogger(logDir string) error {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Configure encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create core for main logs
	mainLogPath := filepath.Join(logDir, "api.log")
	mainLogFile, err := os.OpenFile(mainLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	mainCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(mainLogFile),
		zap.InfoLevel,
	)

	// Create core for analytics logs
	analyticsLogPath := filepath.Join(logDir, "usage_analytics.log")
	analyticsLogFile, err := os.OpenFile(analyticsLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open analytics log file: %v", err)
	}

	analyticsCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(analyticsLogFile),
		zap.InfoLevel,
	)

	// Create loggers
	Logger = zap.New(mainCore, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	SugaredLogger = Logger.Sugar()
	analyticsLogger = zap.New(analyticsCore)

	return nil
}

// LogInfo logs an info message
func LogInfo(format string, args ...interface{}) {
	if SugaredLogger != nil {
		SugaredLogger.Infof(format, args...)
	} else {
		fmt.Printf("[INFO] "+format+"\n", args...)
	}
}

// LogWarning logs a warning message
func LogWarning(format string, args ...interface{}) {
	if SugaredLogger != nil {
		SugaredLogger.Warnf(format, args...)
	} else {
		fmt.Printf("[WARN] "+format+"\n", args...)
	}
}

// LogError logs an error message
func LogError(format string, args ...interface{}) {
	if SugaredLogger != nil {
		SugaredLogger.Errorf(format, args...)
	} else {
		fmt.Printf("[ERROR] "+format+"\n", args...)
	}
}

// LogDebug logs a debug message
func LogDebug(format string, args ...interface{}) {
	if SugaredLogger != nil {
		SugaredLogger.Debugf(format, args...)
	} else {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// LogFatal logs a fatal message and exits
func LogFatal(format string, args ...interface{}) {
	if SugaredLogger != nil {
		SugaredLogger.Fatalf(format, args...)
	} else {
		fmt.Printf("[FATAL] "+format+"\n", args...)
		os.Exit(1)
	}
}

// LogAnalytics logs an analytics event
func LogAnalytics(userID, eventType, details string) {
	if analyticsLogger != nil {
		analyticsLogger.Info("analytics_event",
			zap.String("user_id", userID),
			zap.String("event_type", eventType),
			zap.String("details", details),
			zap.String("timestamp", time.Now().Format(time.RFC3339)),
		)
	} else {
		fmt.Printf("[ANALYTICS] user_id=%s event_type=%s details=%s\n", userID, eventType, details)
	}
}

// CloseLogger closes the logger
func CloseLogger() {
	if Logger != nil {
		Logger.Sync()
	}
	if analyticsLogger != nil {
		analyticsLogger.Sync()
	}
}
