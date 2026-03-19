package api

import (
	"net/http"
	"novabackup/internal/rbac"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware returns a Gin middleware that validates authentication tokens
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for health check and login endpoints
		path := c.Request.URL.Path
		if path == "/api/health" || path == "/api/auth/login" {
			c.Next()
			return
		}

		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Відсутній токен авторизації"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		token := authHeader
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			token = parts[1]
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Відсутній токен авторизації"})
			c.Abort()
			return
		}

		// Validate token with RBAC engine
		user, err := RBACEngine.ValidateSession(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Недійсна сесія: " + err.Error()})
			c.Abort()
			return
		}

		// Store user in context for handlers to use
		c.Set("user", user)
		c.Next()
	}
}

// RequirePermission returns a middleware that checks if user has required permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userValue, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Користувач не автентифікований"})
			c.Abort()
			return
		}

		user, ok := userValue.(*rbac.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Помилка отримання даних користувача"})
			c.Abort()
			return
		}

		if !RBACEngine.CheckPermission(user, permission) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":      "Недостатньо прав доступу",
				"permission": permission,
				"role":       user.Role,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuditMiddleware logs all API requests
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Log after request is processed
		userValue, exists := c.Get("user")
		var userID, username string
		if exists {
			if user, ok := userValue.(*rbac.User); ok {
				userID = user.ID
				username = user.Username
			}
		}

		// Log the action
		AuditEngine.Log(
			userID,
			username,
			c.Request.Method+" "+c.Request.URL.Path,
			c.Request.URL.String(),
			c.ClientIP(),
			c.Writer.Status() < 400,
			map[string]interface{}{
				"status": c.Writer.Status(),
			},
		)
	}
}
