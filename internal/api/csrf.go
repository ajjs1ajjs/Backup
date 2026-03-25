// CSRF Middleware - Cross-Site Request Forgery Protection
package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	csrfTokenHeader   = "X-CSRF-Token"
	csrfCookieName    = "_csrf"
	csrfTokenLifetime = 24 * time.Hour
)

// CSRFConfig holds CSRF protection configuration
type CSRFConfig struct {
	SecretKey  string
	Secure     bool // Use secure cookies
	SameSite   string
	HeaderName string
	CookieName string
}

// DefaultCSRFConfig returns a default CSRF configuration
func DefaultCSRFConfig() *CSRFConfig {
	return &CSRFConfig{
		SecretKey:  "nova-backup-csrf-secret-key-change-in-production",
		Secure:     false,
		SameSite:   "Strict",
		HeaderName: csrfTokenHeader,
		CookieName: csrfCookieName,
	}
}

// CSRFMiddleware returns a Gin middleware that validates CSRF tokens
// for state-changing requests (POST, PUT, DELETE, PATCH)
func CSRFMiddleware(config *CSRFConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultCSRFConfig()
	}

	return func(c *gin.Context) {
		// Skip CSRF for safe methods
		if isSafeMethod(c.Request.Method) {
			c.Next()
			return
		}

		// Skip CSRF for login and logout (they have their own protection)
		path := c.Request.URL.Path
		if path == "/api/auth/login" || path == "/api/auth/logout" {
			c.Next()
			return
		}

		// Get token from header
		token := c.GetHeader(config.HeaderName)
		if token == "" {
			// Try to get from form
			token = c.PostForm(config.HeaderName)
		}

		if token == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "CSRF token missing. Please include the " + config.HeaderName + " header.",
			})
			c.Abort()
			return
		}

		// Get session token for validation
		sessionToken := c.GetHeader("Authorization")
		if sessionToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization token required",
			})
			c.Abort()
			return
		}

		// Validate CSRF token
		if !validateCSRFToken(token, sessionToken, config.SecretKey) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid CSRF token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GenerateCSRFToken generates a CSRF token for the given session
func GenerateCSRFToken(sessionToken, secretKey string) string {
	// Create HMAC of session token with secret key
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(sessionToken))
	h.Write([]byte(time.Now().Round(csrfTokenLifetime).Format(time.RFC3339)))
	return hex.EncodeToString(h.Sum(nil))
}

// validateCSRFToken validates a CSRF token
func validateCSRFToken(token, sessionToken, secretKey string) bool {
	// Generate expected token for current period
	expected := GenerateCSRFToken(sessionToken, secretKey)

	// Use constant-time comparison to prevent timing attacks
	return secureCompare(token, expected)
}

// secureCompare compares two strings in constant time
func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	result := byte(0)
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}

// isSafeMethod checks if HTTP method is safe (no side effects)
func isSafeMethod(method string) bool {
	safeMethods := map[string]bool{
		"GET":     true,
		"HEAD":    true,
		"OPTIONS": true,
		"TRACE":   true,
	}
	return safeMethods[strings.ToUpper(method)]
}

// GetCSRFToken extracts and returns CSRF token from request
func GetCSRFToken(c *gin.Context) string {
	// Try header first
	token := c.GetHeader(csrfTokenHeader)
	if token != "" {
		return token
	}

	// Try cookie
	cookie, err := c.Cookie(csrfCookieName)
	if err == nil {
		return cookie
	}

	return ""
}
