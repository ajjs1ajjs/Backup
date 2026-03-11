package multitenancy

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"novabackup/internal/rbac"
)

// TenantMiddleware provides tenant isolation middleware
type TenantMiddleware struct {
	tenantManager TenantManager
	rbacManager   rbac.RBACManager
}

// NewTenantMiddleware creates a new tenant middleware
func NewTenantMiddleware(tenantManager TenantManager, rbacManager rbac.RBACManager) *TenantMiddleware {
	return &TenantMiddleware{
		tenantManager: tenantManager,
		rbacManager:   rbacManager,
	}
}

// TenantFromRequest extracts tenant ID from HTTP request
func (tm *TenantMiddleware) TenantFromRequest(r *http.Request) string {
	// Try to extract tenant ID from various sources:
	// 1. X-Tenant-ID header
	// 2. URL path (e.g., /api/v1/tenants/{tenantId}/...)
	// 3. User's tenant (from authenticated user)

	// Check header first
	if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
		return tenantID
	}

	// Check URL path
	if tenantID := tm.extractTenantFromPath(r.URL.Path); tenantID != "" {
		return tenantID
	}

	// Get from authenticated user
	user := rbac.GetCurrentUser(r.Context())
	if user != nil {
		return user.TenantID
	}

	return ""
}

// extractTenantFromPath extracts tenant ID from URL path
func (tm *TenantMiddleware) extractTenantFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, part := range parts {
		if part == "tenants" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// RequireTenant middleware ensures a tenant is present in the request
func (tm *TenantMiddleware) RequireTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := tm.TenantFromRequest(r)
		if tenantID == "" {
			http.Error(w, "Tenant ID required", http.StatusBadRequest)
			return
		}

		// Validate tenant exists and is active
		if err := tm.tenantManager.ValidateTenantAccess(r.Context(), tenantID); err != nil {
			http.Error(w, "Invalid tenant", http.StatusForbidden)
			return
		}

		// Add tenant to context
		ctx := tm.tenantManager.WithTenant(r.Context(), tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireTenantAdmin middleware ensures user has tenant admin permissions
func (tm *TenantMiddleware) RequireTenantAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := rbac.GetCurrentUser(r.Context())
		if user == nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Check if user has tenant admin permissions
		hasAdmin, err := tm.rbacManager.CheckPermission(r.Context(), user.ID, "tenant", "update")
		if err != nil || !hasAdmin {
			http.Error(w, "Tenant admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// TenantQuotaMiddleware checks tenant quotas before processing requests
func (tm *TenantMiddleware) TenantQuotaMiddleware(quotaType QuotaType, amount int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID := tm.tenantManager.GetTenantFromContext(r.Context())
			if tenantID == "" {
				http.Error(w, "Tenant context required", http.StatusBadRequest)
				return
			}

			// Check quota
			allowed, err := tm.tenantManager.CheckQuota(r.Context(), tenantID, quotaType, amount)
			if err != nil {
				http.Error(w, "Failed to check quota", http.StatusInternalServerError)
				return
			}

			if !allowed {
				http.Error(w, "Quota exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// TenantResourceMiddleware ensures resource belongs to tenant
func (tm *TenantMiddleware) TenantResourceMiddleware(resourceType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID := tm.tenantManager.GetTenantFromContext(r.Context())
			if tenantID == "" {
				http.Error(w, "Tenant context required", http.StatusBadRequest)
				return
			}

			// Extract resource ID from URL
			resourceID := tm.extractResourceFromPath(r.URL.Path, resourceType)
			if resourceID == "" {
				// If no specific resource ID, continue (this might be a list endpoint)
				next.ServeHTTP(w, r)
				return
			}

			// Check if resource belongs to tenant
			resources, err := tm.tenantManager.GetTenantResources(r.Context(), tenantID)
			if err != nil {
				http.Error(w, "Failed to get tenant resources", http.StatusInternalServerError)
				return
			}

			if !tm.resourceBelongsToTenant(resources, resourceType, resourceID) {
				http.Error(w, "Resource not found or access denied", http.StatusNotFound)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractResourceFromPath extracts resource ID from URL path
func (tm *TenantMiddleware) extractResourceFromPath(path, resourceType string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, part := range parts {
		if part == resourceType && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// resourceBelongsToTenant checks if a resource belongs to a tenant
func (tm *TenantMiddleware) resourceBelongsToTenant(resources *TenantResources, resourceType, resourceID string) bool {
	switch resourceType {
	case "backup":
		for _, res := range resources.Backups {
			if res.ID == resourceID || res.Name == resourceID {
				return true
			}
		}
	case "vm":
		for _, res := range resources.VMs {
			if res.ID == resourceID || res.Name == resourceID {
				return true
			}
		}
	case "storage":
		for _, res := range resources.Storage {
			if res.ID == resourceID || res.Name == resourceID {
				return true
			}
		}
	case "job":
		for _, res := range resources.Jobs {
			if res.ID == resourceID || res.Name == resourceID {
				return true
			}
		}
	case "user":
		for _, res := range resources.Users {
			if res.ID == resourceID || res.Name == resourceID {
				return true
			}
		}
	default:
		// Check custom resources
		if customRes, exists := resources.CustomResources[resourceType]; exists {
			for _, res := range customRes {
				if res.ID == resourceID || res.Name == resourceID {
					return true
				}
			}
		}
	}
	return false
}

// TenantStatsMiddleware adds tenant statistics to response headers
func (tm *TenantMiddleware) TenantStatsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := tm.tenantManager.GetTenantFromContext(r.Context())
		if tenantID != "" {
			// Get quota usage
			usage, err := tm.tenantManager.GetQuotaUsage(r.Context(), tenantID)
			if err == nil {
				// Add usage info to headers for debugging/monitoring
				w.Header().Set("X-Tenant-Backups", fmt.Sprintf("%d", usage.BackupsCount))
				w.Header().Set("X-Tenant-Storage-GB", fmt.Sprintf("%.2f", usage.StorageUsedGB))
				w.Header().Set("X-Tenant-Users", fmt.Sprintf("%d", usage.UsersCount))
				w.Header().Set("X-Tenant-Jobs", fmt.Sprintf("%d", usage.ConcurrentJobsCount))
			}
		}

		next.ServeHTTP(w, r)
	})
}

// Helper functions for tenant context
func GetCurrentTenant(ctx context.Context) string {
	// This would need to be implemented in the tenant manager
	// For now, return empty string
	return ""
}

func RequireCurrentTenant(ctx context.Context, tenantID string) error {
	currentTenant := GetCurrentTenant(ctx)
	if currentTenant == "" {
		return fmt.Errorf("no tenant in context")
	}
	if currentTenant != tenantID {
		return fmt.Errorf("tenant access denied")
	}
	return nil
}

// TenantAwareHandler wraps an HTTP handler to be tenant-aware
type TenantAwareHandler struct {
	tenantManager TenantManager
	handler       http.Handler
}

func NewTenantAwareHandler(tenantManager TenantManager, handler http.Handler) *TenantAwareHandler {
	return &TenantAwareHandler{
		tenantManager: tenantManager,
		handler:       handler,
	}
}

func (t *TenantAwareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Add tenant context if available
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID != "" {
		ctx := t.tenantManager.WithTenant(r.Context(), tenantID)
		r = r.WithContext(ctx)
	}

	t.handler.ServeHTTP(w, r)
}
