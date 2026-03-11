package rbac

import (
	"context"
	"time"
)

// Permission represents a specific permission in the system
type Permission struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"` // e.g., "backup", "restore", "storage"
	Action      string `json:"action"`   // e.g., "create", "read", "update", "delete"
	Scope       string `json:"scope"`    // e.g., "global", "tenant", "own"
}

// Role represents a role with associated permissions
type Role struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// User represents a system user with roles and tenant association
type User struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	FullName     string     `json:"full_name"`
	PasswordHash string     `json:"password_hash"`
	TenantID     string     `json:"tenant_id"`
	Roles        []Role     `json:"roles"`
	IsActive     bool       `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Tenant represents a multi-tenant organization
type Tenant struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Settings    map[string]string `json:"settings"`
	Quotas      TenantQuotas      `json:"quotas"`
	IsActive    bool              `json:"is_active"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// TenantQuotas defines resource quotas for a tenant
type TenantQuotas struct {
	MaxBackups        int64 `json:"max_backups"`
	MaxStorageGB      int64 `json:"max_storage_gb"`
	MaxUsers          int   `json:"max_users"`
	MaxConcurrentJobs int   `json:"max_concurrent_jobs"`
	RetentionDays     int   `json:"retention_days"`
}

// RBACConfig contains RBAC system configuration
type RBACConfig struct {
	DefaultRoles       []Role         `json:"default_roles"`
	DefaultPermissions []Permission   `json:"default_permissions"`
	SessionTimeout     time.Duration  `json:"session_timeout"`
	PasswordPolicy     PasswordPolicy `json:"password_policy"`
}

// PasswordPolicy defines password requirements
type PasswordPolicy struct {
	MinLength         int           `json:"min_length"`
	RequireUppercase  bool          `json:"require_uppercase"`
	RequireLowercase  bool          `json:"require_lowercase"`
	RequireNumbers    bool          `json:"require_numbers"`
	RequireSymbols    bool          `json:"require_symbols"`
	MaxFailedAttempts int           `json:"max_failed_attempts"`
	LockoutDuration   time.Duration `json:"lockout_duration"`
}

// RBACManager manages role-based access control
type RBACManager interface {
	// User management
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, userID string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, userID string) error
	ListUsers(ctx context.Context, tenantID string) ([]User, error)
	AuthenticateUser(ctx context.Context, username, password string) (*User, error)

	// Role management
	CreateRole(ctx context.Context, role *Role) error
	GetRole(ctx context.Context, roleID string) (*Role, error)
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, roleID string) error
	ListRoles(ctx context.Context) ([]Role, error)

	// Permission management
	CreatePermission(ctx context.Context, permission *Permission) error
	GetPermission(ctx context.Context, permissionID string) (*Permission, error)
	ListPermissions(ctx context.Context) ([]Permission, error)

	// Authorization
	CheckPermission(ctx context.Context, userID string, resource, action string) (bool, error)
	GetUserPermissions(ctx context.Context, userID string) ([]Permission, error)
	AssignRole(ctx context.Context, userID, roleID string) error
	RemoveRole(ctx context.Context, userID, roleID string) error

	// Tenant management
	CreateTenant(ctx context.Context, tenant *Tenant) error
	GetTenant(ctx context.Context, tenantID string) (*Tenant, error)
	UpdateTenant(ctx context.Context, tenant *Tenant) error
	DeleteTenant(ctx context.Context, tenantID string) error
	ListTenants(ctx context.Context) ([]Tenant, error)

	// Session management
	CreateSession(ctx context.Context, userID string) (string, error)
	ValidateSession(ctx context.Context, sessionToken string) (*User, error)
	InvalidateSession(ctx context.Context, sessionToken string) error
}

// Default permissions for the system
var DefaultPermissions = []Permission{
	{ID: "backup.create", Name: "Create Backup", Description: "Create backup jobs", Resource: "backup", Action: "create", Scope: "tenant"},
	{ID: "backup.read", Name: "View Backups", Description: "View backup information", Resource: "backup", Action: "read", Scope: "tenant"},
	{ID: "backup.update", Name: "Update Backup", Description: "Modify backup jobs", Resource: "backup", Action: "update", Scope: "tenant"},
	{ID: "backup.delete", Name: "Delete Backup", Description: "Delete backup jobs", Resource: "backup", Action: "delete", Scope: "tenant"},
	{ID: "restore.create", Name: "Create Restore", Description: "Create restore jobs", Resource: "restore", Action: "create", Scope: "tenant"},
	{ID: "restore.read", Name: "View Restores", Description: "View restore information", Resource: "restore", Action: "read", Scope: "tenant"},
	{ID: "storage.read", Name: "View Storage", Description: "View storage information", Resource: "storage", Action: "read", Scope: "tenant"},
	{ID: "storage.update", Name: "Update Storage", Description: "Modify storage configuration", Resource: "storage", Action: "update", Scope: "tenant"},
	{ID: "user.create", Name: "Create User", Description: "Create user accounts", Resource: "user", Action: "create", Scope: "tenant"},
	{ID: "user.read", Name: "View Users", Description: "View user information", Resource: "user", Action: "read", Scope: "tenant"},
	{ID: "user.update", Name: "Update User", Description: "Modify user accounts", Resource: "user", Action: "update", Scope: "tenant"},
	{ID: "user.delete", Name: "Delete User", Description: "Delete user accounts", Resource: "user", Action: "delete", Scope: "tenant"},
	{ID: "tenant.read", Name: "View Tenant", Description: "View tenant information", Resource: "tenant", Action: "read", Scope: "global"},
	{ID: "tenant.update", Name: "Update Tenant", Description: "Modify tenant configuration", Resource: "tenant", Action: "update", Scope: "global"},
	{ID: "system.admin", Name: "System Admin", Description: "Full system administration", Resource: "system", Action: "admin", Scope: "global"},
}

// Default roles for the system
var DefaultRoles = []Role{
	{
		ID:          "admin",
		Name:        "Administrator",
		Description: "Full system administrator with all permissions",
		Permissions: []Permission{DefaultPermissions[14]}, // system.admin
	},
	{
		ID:          "tenant_admin",
		Name:        "Tenant Administrator",
		Description: "Tenant administrator with full tenant permissions",
		Permissions: []Permission{
			DefaultPermissions[0],  // backup.create
			DefaultPermissions[1],  // backup.read
			DefaultPermissions[2],  // backup.update
			DefaultPermissions[3],  // backup.delete
			DefaultPermissions[4],  // restore.create
			DefaultPermissions[5],  // restore.read
			DefaultPermissions[6],  // storage.read
			DefaultPermissions[7],  // storage.update
			DefaultPermissions[8],  // user.create
			DefaultPermissions[9],  // user.read
			DefaultPermissions[10], // user.update
			DefaultPermissions[11], // user.delete
		},
	},
	{
		ID:          "operator",
		Name:        "Operator",
		Description: "Backup operator with backup and restore permissions",
		Permissions: []Permission{
			DefaultPermissions[0], // backup.create
			DefaultPermissions[1], // backup.read
			DefaultPermissions[2], // backup.update
			DefaultPermissions[4], // restore.create
			DefaultPermissions[5], // restore.read
			DefaultPermissions[6], // storage.read
		},
	},
	{
		ID:          "viewer",
		Name:        "Viewer",
		Description: "Read-only access to backups and restores",
		Permissions: []Permission{
			DefaultPermissions[1], // backup.read
			DefaultPermissions[5], // restore.read
			DefaultPermissions[6], // storage.read
		},
	},
}

// Helper functions for permission checking
func HasPermission(user *User, resource, action string) bool {
	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			if perm.Resource == resource && perm.Action == action {
				return true
			}
		}
	}
	return false
}

func HasGlobalPermission(user *User, resource, action string) bool {
	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			if perm.Resource == resource && perm.Action == action && perm.Scope == "global" {
				return true
			}
		}
	}
	return false
}

func HasTenantPermission(user *User, resource, action string) bool {
	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			if perm.Resource == resource && perm.Action == action && (perm.Scope == "tenant" || perm.Scope == "global") {
				return true
			}
		}
	}
	return false
}

func HasOwnPermission(user *User, resource, action string) bool {
	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			if perm.Resource == resource && perm.Action == action && (perm.Scope == "own" || perm.Scope == "tenant" || perm.Scope == "global") {
				return true
			}
		}
	}
	return false
}
