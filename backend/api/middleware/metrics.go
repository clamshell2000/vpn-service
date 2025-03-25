package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/vpn-service/backend/src/monitoring"
	"github.com/vpn-service/backend/src/utils"
)

// MetricsMiddleware is middleware that collects metrics for API requests
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Start timer
		start := time.Now()

		// Create response writer wrapper to capture status code
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call next handler
		next.ServeHTTP(rw, r)

		// Calculate duration
		duration := time.Since(start)

		// Record metrics
		if monitoring.MetricsCollector != nil {
			monitoring.MetricsCollector.IncrementAPIRequestCount(r.Method, r.URL.Path, strconv.Itoa(rw.statusCode))
			monitoring.MetricsCollector.ObserveAPIRequestDuration(r.Method, r.URL.Path, strconv.Itoa(rw.statusCode), duration)
		}

		// Log request
		utils.LogInfo("API Request: %s %s %d %s", r.Method, r.URL.Path, rw.statusCode, duration)
	})
}

// responseWriter is a wrapper for http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the status code if not already set
func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}
