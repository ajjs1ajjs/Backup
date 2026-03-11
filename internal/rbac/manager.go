package rbac

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// SQLRBACManager implements RBACManager using SQL database
type SQLRBACManager struct {
	db *sql.DB
}

// NewSQLRBACManager creates a new SQL-based RBAC manager
func NewSQLRBACManager(db *sql.DB) *SQLRBACManager {
	return &SQLRBACManager{db: db}
}

// Initialize creates the necessary database tables
func (r *SQLRBACManager) Initialize(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS tenants (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			settings JSON,
			quotas JSON,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS permissions (
			id VARCHAR(100) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			resource VARCHAR(100) NOT NULL,
			action VARCHAR(50) NOT NULL,
			scope VARCHAR(20) NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS roles (
			id VARCHAR(100) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS role_permissions (
			role_id VARCHAR(100),
			permission_id VARCHAR(100),
			PRIMARY KEY (role_id, permission_id),
			FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
			FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY,
			username VARCHAR(100) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			full_name VARCHAR(255),
			password_hash VARCHAR(255) NOT NULL,
			tenant_id VARCHAR(36),
			is_active BOOLEAN DEFAULT true,
			last_login_at TIMESTAMP,
			failed_attempts INTEGER DEFAULT 0,
			locked_until TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS user_roles (
			user_id VARCHAR(36),
			role_id VARCHAR(100),
			PRIMARY KEY (user_id, role_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			token VARCHAR(64) PRIMARY KEY,
			user_id VARCHAR(36),
			expires_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
	}

	for _, query := range queries {
		if _, err := r.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	// Insert default permissions and roles
	if err := r.seedDefaults(ctx); err != nil {
		return fmt.Errorf("failed to seed defaults: %w", err)
	}

	return nil
}

func (r *SQLRBACManager) seedDefaults(ctx context.Context) error {
	// Insert default permissions
	for _, perm := range DefaultPermissions {
		query := `INSERT OR IGNORE INTO permissions (id, name, description, resource, action, scope) VALUES (?, ?, ?, ?, ?, ?)`
		_, err := r.db.ExecContext(ctx, query, perm.ID, perm.Name, perm.Description, perm.Resource, perm.Action, perm.Scope)
		if err != nil {
			return fmt.Errorf("failed to insert permission %s: %w", perm.ID, err)
		}
	}

	// Insert default roles
	for _, role := range DefaultRoles {
		query := `INSERT OR IGNORE INTO roles (id, name, description) VALUES (?, ?, ?)`
		_, err := r.db.ExecContext(ctx, query, role.ID, role.Name, role.Description)
		if err != nil {
			return fmt.Errorf("failed to insert role %s: %w", role.ID, err)
		}

		// Insert role permissions
		for _, perm := range role.Permissions {
			query := `INSERT OR IGNORE INTO role_permissions (role_id, permission_id) VALUES (?, ?)`
			_, err := r.db.ExecContext(ctx, query, role.ID, perm.ID)
			if err != nil {
				return fmt.Errorf("failed to insert role permission %s-%s: %w", role.ID, perm.ID, err)
			}
		}
	}

	return nil
}

// User management
func (r *SQLRBACManager) CreateUser(ctx context.Context, user *User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `INSERT INTO users (id, username, email, full_name, password_hash, tenant_id, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err = r.db.ExecContext(ctx, query, user.ID, user.Username, user.Email, user.FullName, string(hashedPassword), user.TenantID, user.IsActive)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Assign roles
	for _, role := range user.Roles {
		if err := r.AssignRole(ctx, user.ID, role.ID); err != nil {
			return fmt.Errorf("failed to assign role %s: %w", role.ID, err)
		}
	}

	return nil
}

func (r *SQLRBACManager) GetUser(ctx context.Context, userID string) (*User, error) {
	query := `SELECT id, username, email, full_name, tenant_id, is_active, last_login_at, created_at, updated_at FROM users WHERE id = ?`
	var user User
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName,
		&user.TenantID, &user.IsActive, &lastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	// Get user roles
	roles, err := r.getUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	user.Roles = roles

	return &user, nil
}

func (r *SQLRBACManager) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `SELECT id, username, email, full_name, tenant_id, is_active, last_login_at, created_at, updated_at FROM users WHERE username = ?`
	var user User
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName,
		&user.TenantID, &user.IsActive, &lastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	// Get user roles
	roles, err := r.getUserRoles(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	user.Roles = roles

	return &user, nil
}

func (r *SQLRBACManager) getUserRoles(ctx context.Context, userID string) ([]Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.created_at, r.updated_at
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		// Get role permissions
		permissions, err := r.getRolePermissions(ctx, role.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get role permissions: %w", err)
		}
		role.Permissions = permissions

		roles = append(roles, role)
	}

	return roles, nil
}

func (r *SQLRBACManager) getRolePermissions(ctx context.Context, roleID string) ([]Permission, error) {
	query := `
		SELECT p.id, p.name, p.description, p.resource, p.action, p.scope
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = ?
	`

	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query role permissions: %w", err)
	}
	defer rows.Close()

	var permissions []Permission
	for rows.Next() {
		var perm Permission
		if err := rows.Scan(&perm.ID, &perm.Name, &perm.Description, &perm.Resource, &perm.Action, &perm.Scope); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

func (r *SQLRBACManager) UpdateUser(ctx context.Context, user *User) error {
	query := `UPDATE users SET username = ?, email = ?, full_name = ?, tenant_id = ?, is_active = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, user.Username, user.Email, user.FullName, user.TenantID, user.IsActive, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Update roles - remove all and re-add
	_, err = r.db.ExecContext(ctx, "DELETE FROM user_roles WHERE user_id = ?", user.ID)
	if err != nil {
		return fmt.Errorf("failed to remove user roles: %w", err)
	}

	for _, role := range user.Roles {
		if err := r.AssignRole(ctx, user.ID, role.ID); err != nil {
			return fmt.Errorf("failed to assign role %s: %w", role.ID, err)
		}
	}

	return nil
}

func (r *SQLRBACManager) DeleteUser(ctx context.Context, userID string) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (r *SQLRBACManager) ListUsers(ctx context.Context, tenantID string) ([]User, error) {
	query := `SELECT id, username, email, full_name, tenant_id, is_active, last_login_at, created_at, updated_at FROM users WHERE tenant_id = ?`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var lastLoginAt sql.NullTime

		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.FullName,
			&user.TenantID, &user.IsActive, &lastLoginAt, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		if lastLoginAt.Valid {
			user.LastLoginAt = &lastLoginAt.Time
		}

		// Get user roles
		roles, err := r.getUserRoles(ctx, user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user roles: %w", err)
		}
		user.Roles = roles

		users = append(users, user)
	}

	return users, nil
}

func (r *SQLRBACManager) AuthenticateUser(ctx context.Context, username, password string) (*User, error) {
	user, err := r.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.IsActive {
		return nil, fmt.Errorf("user is inactive")
	}

	// Check if account is locked
	// TODO: Implement account lockout logic

	// Verify password
	query := `SELECT password_hash FROM users WHERE id = ?`
	var storedHash string
	err = r.db.QueryRowContext(ctx, query, user.ID).Scan(&storedHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get password hash: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
		// Increment failed attempts
		_, err = r.db.ExecContext(ctx, "UPDATE users SET failed_attempts = failed_attempts + 1 WHERE id = ?", user.ID)
		return nil, fmt.Errorf("invalid password")
	}

	// Reset failed attempts and update last login
	_, err = r.db.ExecContext(ctx, "UPDATE users SET failed_attempts = 0, last_login_at = CURRENT_TIMESTAMP WHERE id = ?", user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update login info: %w", err)
	}

	user.LastLoginAt = &[]time.Time{time.Now()}[0]
	return user, nil
}

// Session management
func (r *SQLRBACManager) CreateSession(ctx context.Context, userID string) (string, error) {
	token := generateSessionToken()
	expiresAt := time.Now().Add(24 * time.Hour) // 24 hour session

	query := `INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, token, userID, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return token, nil
}

func (r *SQLRBACManager) ValidateSession(ctx context.Context, sessionToken string) (*User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.full_name, u.tenant_id, u.is_active, u.last_login_at, u.created_at, u.updated_at
		FROM users u
		INNER JOIN sessions s ON u.id = s.user_id
		WHERE s.token = ? AND s.expires_at > CURRENT_TIMESTAMP
	`

	var user User
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, sessionToken).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName,
		&user.TenantID, &user.IsActive, &lastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	// Get user roles
	roles, err := r.getUserRoles(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	user.Roles = roles

	return &user, nil
}

func (r *SQLRBACManager) InvalidateSession(ctx context.Context, sessionToken string) error {
	query := `DELETE FROM sessions WHERE token = ?`
	_, err := r.db.ExecContext(ctx, query, sessionToken)
	if err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}
	return nil
}

// Authorization
func (r *SQLRBACManager) CheckPermission(ctx context.Context, userID string, resource, action string) (bool, error) {
	query := `
		SELECT COUNT(*) > 0
		FROM users u
		INNER JOIN user_roles ur ON u.id = ur.user_id
		INNER JOIN role_permissions rp ON ur.role_id = rp.role_id
		INNER JOIN permissions p ON rp.permission_id = p.id
		WHERE u.id = ? AND p.resource = ? AND p.action = ? AND p.scope IN ('global', 'tenant')
	`

	var hasPermission bool
	err := r.db.QueryRowContext(ctx, query, userID, resource, action).Scan(&hasPermission)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return hasPermission, nil
}

func (r *SQLRBACManager) GetUserPermissions(ctx context.Context, userID string) ([]Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.description, p.resource, p.action, p.scope
		FROM users u
		INNER JOIN user_roles ur ON u.id = ur.user_id
		INNER JOIN role_permissions rp ON ur.role_id = rp.role_id
		INNER JOIN permissions p ON rp.permission_id = p.id
		WHERE u.id = ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	defer rows.Close()

	var permissions []Permission
	for rows.Next() {
		var perm Permission
		if err := rows.Scan(&perm.ID, &perm.Name, &perm.Description, &perm.Resource, &perm.Action, &perm.Scope); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

func (r *SQLRBACManager) AssignRole(ctx context.Context, userID, roleID string) error {
	query := `INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)`
	_, err := r.db.ExecContext(ctx, query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}
	return nil
}

func (r *SQLRBACManager) RemoveRole(ctx context.Context, userID, roleID string) error {
	query := `DELETE FROM user_roles WHERE user_id = ? AND role_id = ?`
	_, err := r.db.ExecContext(ctx, query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}
	return nil
}

// Role management
func (r *SQLRBACManager) CreateRole(ctx context.Context, role *Role) error {
	query := `INSERT INTO roles (id, name, description) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, role.ID, role.Name, role.Description)
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	// Add permissions
	for _, perm := range role.Permissions {
		query := `INSERT OR IGNORE INTO role_permissions (role_id, permission_id) VALUES (?, ?)`
		_, err := r.db.ExecContext(ctx, query, role.ID, perm.ID)
		if err != nil {
			return fmt.Errorf("failed to add role permission: %w", err)
		}
	}

	return nil
}

func (r *SQLRBACManager) GetRole(ctx context.Context, roleID string) (*Role, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM roles WHERE id = ?`
	var role Role

	err := r.db.QueryRowContext(ctx, query, roleID).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	permissions, err := r.getRolePermissions(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	role.Permissions = permissions

	return &role, nil
}

func (r *SQLRBACManager) UpdateRole(ctx context.Context, role *Role) error {
	query := `UPDATE roles SET name = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, role.Name, role.Description, role.ID)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	// Update permissions - remove all and re-add
	_, err = r.db.ExecContext(ctx, "DELETE FROM role_permissions WHERE role_id = ?", role.ID)
	if err != nil {
		return fmt.Errorf("failed to remove role permissions: %w", err)
	}

	for _, perm := range role.Permissions {
		query := `INSERT OR IGNORE INTO role_permissions (role_id, permission_id) VALUES (?, ?)`
		_, err := r.db.ExecContext(ctx, query, role.ID, perm.ID)
		if err != nil {
			return fmt.Errorf("failed to add role permission: %w", err)
		}
	}

	return nil
}

func (r *SQLRBACManager) DeleteRole(ctx context.Context, roleID string) error {
	query := `DELETE FROM roles WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}

func (r *SQLRBACManager) ListRoles(ctx context.Context) ([]Role, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM roles ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		permissions, err := r.getRolePermissions(ctx, role.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get role permissions: %w", err)
		}
		role.Permissions = permissions

		roles = append(roles, role)
	}

	return roles, nil
}

// Permission management
func (r *SQLRBACManager) CreatePermission(ctx context.Context, permission *Permission) error {
	query := `INSERT INTO permissions (id, name, description, resource, action, scope) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, permission.ID, permission.Name, permission.Description, permission.Resource, permission.Action, permission.Scope)
	if err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}
	return nil
}

func (r *SQLRBACManager) GetPermission(ctx context.Context, permissionID string) (*Permission, error) {
	query := `SELECT id, name, description, resource, action, scope FROM permissions WHERE id = ?`
	var perm Permission

	err := r.db.QueryRowContext(ctx, query, permissionID).Scan(&perm.ID, &perm.Name, &perm.Description, &perm.Resource, &perm.Action, &perm.Scope)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return &perm, nil
}

func (r *SQLRBACManager) ListPermissions(ctx context.Context) ([]Permission, error) {
	query := `SELECT id, name, description, resource, action, scope FROM permissions ORDER BY resource, action`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	defer rows.Close()

	var permissions []Permission
	for rows.Next() {
		var perm Permission
		if err := rows.Scan(&perm.ID, &perm.Name, &perm.Description, &perm.Resource, &perm.Action, &perm.Scope); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// Tenant management
func (r *SQLRBACManager) CreateTenant(ctx context.Context, tenant *Tenant) error {
	query := `INSERT INTO tenants (id, name, description, settings, quotas, is_active) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, tenant.ID, tenant.Name, tenant.Description, tenant.Settings, tenant.Quotas, tenant.IsActive)
	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}
	return nil
}

func (r *SQLRBACManager) GetTenant(ctx context.Context, tenantID string) (*Tenant, error) {
	query := `SELECT id, name, description, settings, quotas, is_active, created_at, updated_at FROM tenants WHERE id = ?`
	var tenant Tenant

	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(&tenant.ID, &tenant.Name, &tenant.Description, &tenant.Settings, &tenant.Quotas, &tenant.IsActive, &tenant.CreatedAt, &tenant.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return &tenant, nil
}

func (r *SQLRBACManager) UpdateTenant(ctx context.Context, tenant *Tenant) error {
	query := `UPDATE tenants SET name = ?, description = ?, settings = ?, quotas = ?, is_active = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, tenant.Name, tenant.Description, tenant.Settings, tenant.Quotas, tenant.IsActive, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}
	return nil
}

func (r *SQLRBACManager) DeleteTenant(ctx context.Context, tenantID string) error {
	query := `DELETE FROM tenants WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}
	return nil
}

func (r *SQLRBACManager) ListTenants(ctx context.Context) ([]Tenant, error) {
	query := `SELECT id, name, description, settings, quotas, is_active, created_at, updated_at FROM tenants ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []Tenant
	for rows.Next() {
		var tenant Tenant
		if err := rows.Scan(&tenant.ID, &tenant.Name, &tenant.Description, &tenant.Settings, &tenant.Quotas, &tenant.IsActive, &tenant.CreatedAt, &tenant.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenants = append(tenants, tenant)
	}

	return tenants, nil
}

// Helper functions
func generateSessionToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Utility functions for creating new entities
func NewUser(username, email, fullName, password, tenantID string) *User {
	return &User{
		ID:           generateID(),
		Username:     username,
		Email:        email,
		FullName:     fullName,
		PasswordHash: password,
		TenantID:     tenantID,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func NewTenant(name, description string) *Tenant {
	return &Tenant{
		ID:          generateID(),
		Name:        name,
		Description: description,
		Settings:    make(map[string]string),
		Quotas:      TenantQuotas{},
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
