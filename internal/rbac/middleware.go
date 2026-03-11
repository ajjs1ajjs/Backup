package rbac

import (
	"context"
	"net/http"
	"strings"
)

// ContextKey is used for storing user context in HTTP requests
type ContextKey string

const (
	UserContextKey ContextKey = "user"
	TenantContextKey ContextKey = "tenant"
)

// AuthMiddleware provides authentication and authorization middleware
type AuthMiddleware struct {
	rbacManager RBACManager
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(rbacManager RBACManager) *AuthMiddleware {
	return &AuthMiddleware{rbacManager: rbacManager}
}

// AuthenticateMiddleware validates session tokens and sets user context
func (a *AuthMiddleware) AuthenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session token from Authorization header or cookie
		token := a.extractToken(r)
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate session
		user, err := a.rbacManager.ValidateSession(r.Context(), token)
		if err != nil {
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		// Set user and tenant in context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		ctx = context.WithValue(ctx, TenantContextKey, user.TenantID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission middleware checks if user has required permission
func (a *AuthMiddleware) RequirePermission(resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := a.getUserFromContext(r.Context())
			if user == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check permission
			hasPermission, err := a.rbacManager.CheckPermission(r.Context(), user.ID, resource, action)
			if err != nil || !hasPermission {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole middleware checks if user has required role
func (a *AuthMiddleware) RequireRole(roleName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := a.getUserFromContext(r.Context())
			if user == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if user has the required role
			hasRole := false
			for _, role := range user.Roles {
				if role.Name == roleName {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireTenantAccess middleware ensures user can only access their own tenant
func (a *AuthMiddleware) RequireTenantAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := a.getUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get tenant ID from URL or context
		requestedTenantID := a.getTenantFromRequest(r)
		if requestedTenantID == "" {
			requestedTenantID = user.TenantID
		}

		// Check if user can access the requested tenant
		if user.TenantID != requestedTenantID && !HasGlobalPermission(user, "tenant", "read") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), TenantContextKey, requestedTenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper functions
func (a *AuthMiddleware) extractToken(r *http.Request) string {
	// Try Authorization header first
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Try cookie
	cookie, err := r.Cookie("session_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
}

func (a *AuthMiddleware) getUserFromContext(ctx context.Context) *User {
	if user, ok := ctx.Value(UserContextKey).(*User); ok {
		return user
	}
	return nil
}

func (a *AuthMiddleware) getTenantFromRequest(r *http.Request) string {
	// Try to extract tenant ID from URL path
	// This is a simplified implementation - in practice you'd use proper routing
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	
	// Look for tenant ID in common patterns
	for i, part := range parts {
		if part == "tenants" && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	return ""
}

// GetCurrentUser returns the current user from context
func GetCurrentUser(ctx context.Context) *User {
	if user, ok := ctx.Value(UserContextKey).(*User); ok {
		return user
	}
	return nil
}

// GetCurrentTenant returns the current tenant ID from context
func GetCurrentTenant(ctx context.Context) string {
	if tenantID, ok := ctx.Value(TenantContextKey).(string); ok {
		return tenantID
	}
	return ""
}

// HasCurrentUserPermission checks if current user has permission
func HasCurrentUserPermission(ctx context.Context, resource, action string) bool {
	user := GetCurrentUser(ctx)
	if user == nil {
		return false
	}
	return HasPermission(user, resource, action)
}

// RequireCurrentUserPermission panics if user doesn't have permission
func RequireCurrentUserPermission(ctx context.Context, resource, action string) {
	if !HasCurrentUserPermission(ctx, resource, action) {
		panic("insufficient permissions")
	}
}
