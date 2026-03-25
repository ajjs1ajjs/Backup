// Error Handling Middleware - Hide internal error details from clients
package api

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Request string `json:"request_id,omitempty"`
}

// ErrorMiddleware returns a Gin middleware that handles panics and errors
// It logs detailed errors server-side but returns generic messages to clients
func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log detailed error with stack trace
				stack := debug.Stack()
				log.Printf("🔴 PANIC: %v\n%s", err, string(stack))

				// Return generic error to client
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error: "Внутрішня помилка сервера. Будь ласка, спробуйте пізніше.",
					Code:  "INTERNAL_ERROR",
				})
				c.Abort()
				return
			}
		}()

		c.Next()

		// Handle errors that occurred during processing
		if len(c.Errors) > 0 {
			// Log all errors
			for _, e := range c.Errors {
				log.Printf("🔴 Error: %v", e)
			}

			// Return generic error
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "Помилка обробки запиту.",
				Code:  "REQUEST_ERROR",
			})
		}
	}
}

// SafeError wraps an error for safe client response
// Internal errors are logged but not exposed to clients
type SafeError struct {
	PublicMessage string
	InternalError error
	StatusCode    int
	Code          string
	LogContext    map[string]interface{}
}

// HandleSafeError processes a SafeError and sends appropriate response
func HandleSafeError(c *gin.Context, err SafeError) {
	// Log internal error with context
	if err.InternalError != nil {
		logMsg := fmt.Sprintf("[%s] %v", err.Code, err.InternalError)
		if err.LogContext != nil {
			log.Printf("🔴 %s - Context: %+v", logMsg, err.LogContext)
		} else {
			log.Printf("🔴 %s", logMsg)
		}
	}

	// Determine status code
	statusCode := err.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}

	// Return public message to client
	c.JSON(statusCode, ErrorResponse{
		Error: err.PublicMessage,
		Code:  err.Code,
	})
}

// NewSafeError creates a new SafeError with default values
func NewSafeError(internalErr error, publicMsg string, code string) SafeError {
	return SafeError{
		PublicMessage: publicMsg,
		InternalError: internalErr,
		StatusCode:    http.StatusInternalServerError,
		Code:          code,
		LogContext:    make(map[string]interface{}),
	}
}

// WithStatus sets the HTTP status code
func (e SafeError) WithStatus(status int) SafeError {
	e.StatusCode = status
	return e
}

// WithContext adds context for logging
func (e SafeError) WithContext(key string, value interface{}) SafeError {
	e.LogContext[key] = value
	return e
}

// isSensitiveError checks if an error message contains sensitive information
func isSensitiveError(errMsg string) bool {
	sensitivePatterns := []string{
		"password",
		"secret",
		"key",
		"token",
		"credential",
		"connection string",
		"dsn",
		"private",
		"certificate",
		"stack trace",
		"goroutine",
		"panic",
		"filepath",
		"C:\\",
		"/home/",
		"/var/",
		"mysql",
		"postgres",
		"mongodb",
	}

	errLower := strings.ToLower(errMsg)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(errLower, pattern) {
			return true
		}
	}

	return false
}

// SanitizeError returns a safe error message for clients
// If the error contains sensitive information, returns a generic message
func SanitizeError(err error, fallback string) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()

	if isSensitiveError(errMsg) {
		log.Printf("⚠️ Sensitive error detected, sanitizing: %s", errMsg)
		return fallback
	}

	// For non-sensitive errors, return a sanitized version
	// Remove potential path information
	sanitized := sanitizePaths(errMsg)

	return sanitized
}

// sanitizePaths removes or masks file paths from error messages
func sanitizePaths(msg string) string {
	// Common path patterns to sanitize
	pathPatterns := []struct {
		pattern     string
		replacement string
	}{
		{"C:\\\\", "[PATH]"},
		{"/home/", "[PATH]"},
		{"/var/", "[PATH]"},
		{"/etc/", "[PATH]"},
		{"/tmp/", "[PATH]"},
		{"/opt/", "[PATH]"},
	}

	sanitized := msg
	for _, p := range pathPatterns {
		sanitized = strings.ReplaceAll(sanitized, p.pattern, p.replacement)
	}

	return sanitized
}

// APIError returns a standardized API error response
func APIError(c *gin.Context, statusCode int, code string, message string) {
	c.JSON(statusCode, ErrorResponse{
		Error: message,
		Code:  code,
	})
}

// Common error response helpers
func BadRequest(c *gin.Context, message string) {
	APIError(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

func Unauthorized(c *gin.Context, message string) {
	APIError(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

func Forbidden(c *gin.Context, message string) {
	APIError(c, http.StatusForbidden, "FORBIDDEN", message)
}

func NotFound(c *gin.Context, message string) {
	APIError(c, http.StatusNotFound, "NOT_FOUND", message)
}

func Conflict(c *gin.Context, message string) {
	APIError(c, http.StatusConflict, "CONFLICT", message)
}

func InternalError(c *gin.Context) {
	APIError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутрішня помилка сервера")
}

func ServiceUnavailable(c *gin.Context, message string) {
	APIError(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message)
}
