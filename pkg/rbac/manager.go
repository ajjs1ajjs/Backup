// Package rbac provides role-based access control functionality
package rbac

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Permission represents a specific permission
type Permission string

// Predefined permissions
const (
	// Job permissions
	PermJobCreate   Permission = "job:create"
	PermJobRead     Permission = "job:read"
	PermJobUpdate   Permission = "job:update"
	PermJobDelete   Permission = "job:delete"
	PermJobExecute  Permission = "job:execute"

	// Backup permissions
	PermBackupCreate Permission = "backup:create"
	PermBackupRead   Permission = "backup:read"
	PermBackupDelete Permission = "backup:delete"
	PermBackupRestore Permission = "backup:restore"

	// VM permissions
	PermVMRead    Permission = "vm:read"
	PermVMSnapshot Permission = "vm:snapshot"
	PermVMRestore  Permission = "vm:restore"

	// Storage permissions
	PermStorageRead   Permission = "storage:read"
	PermStorageWrite  Permission = "storage:write"
	PermStorageDelete Permission = "storage:delete"

	// Replication permissions
	PermReplicationRead   Permission = "replication:read"
	PermReplicationWrite  Permission = "replication:write"
	PermReplicationDelete Permission = "replication:delete"

	// Monitoring permissions
	PermMonitoringRead  Permission = "monitoring:read"
	PermMonitoringAdmin Permission = "monitoring:admin"

	// User management permissions
	PermUserCreate Permission = "user:create"
	PermUserRead   Permission = "user:read"
	PermUserUpdate Permission = "user:update"
	PermUserDelete Permission = "user:delete"

	// System permissions
	PermSystemConfig Permission = "system:config"
	PermSystemLogs   Permission = "system:logs"
	PermSystemAdmin  Permission = "system:admin"
)

// Role represents a user role with permissions
type Role struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// User represents a system user
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never expose in JSON
	RoleIDs      []string  `json:"role_ids"`
	Active       bool      `json:"active"`
	LastLogin    time.Time `json:"last_login"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Session represents an active user session
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// RBACManager manages roles, users, and permissions
type RBACManager struct {
	logger   *zap.Logger
	roles    map[string]*Role
	users    map[string]*User
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewRBACManager creates a new RBAC manager
func NewRBACManager(logger *zap.Logger) *RBACManager {
	manager := &RBACManager{
		logger:   logger.With(zap.String("component", "rbac")),
		roles:    make(map[string]*Role),
		users:    make(map[string]*User),
		sessions: make(map[string]*Session),
	}

	// Create default roles
	manager.createDefaultRoles()

	return manager
}

// createDefaultRoles creates the default system roles
func (r *RBACManager) createDefaultRoles() {
	now := time.Now()

	// Admin role - full access
	adminRole := &Role{
		ID:          "role_admin",
		Name:        "Administrator",
		Description: "Full system access",
		Permissions: []Permission{
			PermJobCreate, PermJobRead, PermJobUpdate, PermJobDelete, PermJobExecute,
			PermBackupCreate, PermBackupRead, PermBackupDelete, PermBackupRestore,
			PermVMRead, PermVMSnapshot, PermVMRestore,
			PermStorageRead, PermStorageWrite, PermStorageDelete,
			PermReplicationRead, PermReplicationWrite, PermReplicationDelete,
			PermMonitoringRead, PermMonitoringAdmin,
			PermUserCreate, PermUserRead, PermUserUpdate, PermUserDelete,
			PermSystemConfig, PermSystemLogs, PermSystemAdmin,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.roles[adminRole.ID] = adminRole

	// Backup Operator role
	operatorRole := &Role{
		ID:          "role_operator",
		Name:        "Backup Operator",
		Description: "Can manage backups and jobs",
		Permissions: []Permission{
			PermJobCreate, PermJobRead, PermJobUpdate, PermJobDelete, PermJobExecute,
			PermBackupCreate, PermBackupRead, PermBackupDelete, PermBackupRestore,
			PermVMRead, PermVMSnapshot, PermVMRestore,
			PermStorageRead, PermStorageWrite,
			PermReplicationRead, PermReplicationWrite,
			PermMonitoringRead,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.roles[operatorRole.ID] = operatorRole

	// Viewer role - read-only
	viewerRole := &Role{
		ID:          "role_viewer",
		Name:        "Viewer",
		Description: "Read-only access",
		Permissions: []Permission{
			PermJobRead,
			PermBackupRead,
			PermVMRead,
			PermStorageRead,
			PermReplicationRead,
			PermMonitoringRead,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.roles[viewerRole.ID] = viewerRole

	r.logger.Info("Default roles created",
		zap.Int("count", len(r.roles)))
}

// CreateRole creates a new role
func (r *RBACManager) CreateRole(role *Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.roles[role.ID]; exists {
		return fmt.Errorf("role already exists: %s", role.ID)
	}

	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()
	r.roles[role.ID] = role

	r.logger.Info("Role created", zap.String("id", role.ID), zap.String("name", role.Name))
	return nil
}

// GetRole gets a role by ID
func (r *RBACManager) GetRole(roleID string) (*Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	role, exists := r.roles[roleID]
	if !exists {
		return nil, fmt.Errorf("role not found: %s", roleID)
	}

	return role, nil
}

// UpdateRole updates a role
func (r *RBACManager) UpdateRole(role *Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.roles[role.ID]; !exists {
		return fmt.Errorf("role not found: %s", role.ID)
	}

	role.UpdatedAt = time.Now()
	r.roles[role.ID] = role

	r.logger.Info("Role updated", zap.String("id", role.ID))
	return nil
}

// DeleteRole deletes a role
func (r *RBACManager) DeleteRole(roleID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.roles[roleID]; !exists {
		return fmt.Errorf("role not found: %s", roleID)
	}

	delete(r.roles, roleID)
	r.logger.Info("Role deleted", zap.String("id", roleID))
	return nil
}

// ListRoles lists all roles
func (r *RBACManager) ListRoles() []*Role {
	r.mu.RLock()
	defer r.mu.RUnlock()

	roles := make([]*Role, 0, len(r.roles))
	for _, role := range r.roles {
		roles = append(roles, role)
	}

	return roles
}

// CreateUser creates a new user
func (r *RBACManager) CreateUser(user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; exists {
		return fmt.Errorf("user already exists: %s", user.ID)
	}

	// Check if username is taken
	for _, u := range r.users {
		if u.Username == user.Username {
			return fmt.Errorf("username already taken: %s", user.Username)
		}
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Active = true

	r.users[user.ID] = user

	r.logger.Info("User created", zap.String("id", user.ID), zap.String("username", user.Username))
	return nil
}

// GetUser gets a user by ID
func (r *RBACManager) GetUser(userID string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", userID)
	}

	return user, nil
}

// GetUserByUsername gets a user by username
func (r *RBACManager) GetUserByUsername(username string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user not found: %s", username)
}

// UpdateUser updates a user
func (r *RBACManager) UpdateUser(user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; !exists {
		return fmt.Errorf("user not found: %s", user.ID)
	}

	user.UpdatedAt = time.Now()
	r.users[user.ID] = user

	r.logger.Info("User updated", zap.String("id", user.ID))
	return nil
}

// DeleteUser deletes a user
func (r *RBACManager) DeleteUser(userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[userID]; !exists {
		return fmt.Errorf("user not found: %s", userID)
	}

	delete(r.users, userID)
	r.logger.Info("User deleted", zap.String("id", userID))
	return nil
}

// ListUsers lists all users
func (r *RBACManager) ListUsers() []*User {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}

	return users
}

// AssignRoleToUser assigns a role to a user
func (r *RBACManager) AssignRoleToUser(userID, roleID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[userID]
	if !exists {
		return fmt.Errorf("user not found: %s", userID)
	}

	if _, exists := r.roles[roleID]; !exists {
		return fmt.Errorf("role not found: %s", roleID)
	}

	// Check if role is already assigned
	for _, id := range user.RoleIDs {
		if id == roleID {
			return fmt.Errorf("role already assigned to user")
		}
	}

	user.RoleIDs = append(user.RoleIDs, roleID)
	user.UpdatedAt = time.Now()

	r.logger.Info("Role assigned to user",
		zap.String("user", userID),
		zap.String("role", roleID))
	return nil
}

// RemoveRoleFromUser removes a role from a user
func (r *RBACManager) RemoveRoleFromUser(userID, roleID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[userID]
	if !exists {
		return fmt.Errorf("user not found: %s", userID)
	}

	newRoleIDs := make([]string, 0)
	for _, id := range user.RoleIDs {
		if id != roleID {
			newRoleIDs = append(newRoleIDs, id)
		}
	}

	user.RoleIDs = newRoleIDs
	user.UpdatedAt = time.Now()

	r.logger.Info("Role removed from user",
		zap.String("user", userID),
		zap.String("role", roleID))
	return nil
}

// GetUserPermissions gets all permissions for a user
func (r *RBACManager) GetUserPermissions(userID string) ([]Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", userID)
	}

	permissionMap := make(map[Permission]bool)
	for _, roleID := range user.RoleIDs {
		role, exists := r.roles[roleID]
		if exists {
			for _, perm := range role.Permissions {
				permissionMap[perm] = true
			}
		}
	}

	permissions := make([]Permission, 0, len(permissionMap))
	for perm := range permissionMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// HasPermission checks if a user has a specific permission
func (r *RBACManager) HasPermission(userID string, permission Permission) bool {
	permissions, err := r.GetUserPermissions(userID)
	if err != nil {
		return false
	}

	for _, perm := range permissions {
		if perm == permission || perm == PermSystemAdmin {
			// System admin has all permissions
			return true
		}
	}

	return false
}

// CreateSession creates a new user session
func (r *RBACManager) CreateSession(userID, ip, userAgent string, duration time.Duration) (*Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[userID]; !exists {
		return nil, fmt.Errorf("user not found: %s", userID)
	}

	session := &Session{
		ID:        fmt.Sprintf("sess_%d", time.Now().UnixNano()),
		UserID:    userID,
		Token:     generateToken(),
		IP:        ip,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(duration),
	}

	r.sessions[session.ID] = session

	r.logger.Info("Session created",
		zap.String("session", session.ID),
		zap.String("user", userID))
	return session, nil
}

// GetSession gets a session by ID
func (r *RBACManager) GetSession(sessionID string) (*Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

// GetSessionByToken gets a session by token
func (r *RBACManager) GetSessionByToken(token string) (*Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, session := range r.sessions {
		if session.Token == token {
			if time.Now().After(session.ExpiresAt) {
				return nil, fmt.Errorf("session expired")
			}
			return session, nil
		}
	}

	return nil, fmt.Errorf("session not found")
}

// DeleteSession deletes a session
func (r *RBACManager) DeleteSession(sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[sessionID]; !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	delete(r.sessions, sessionID)
	r.logger.Info("Session deleted", zap.String("session", sessionID))
	return nil
}

// CleanupExpiredSessions removes expired sessions
func (r *RBACManager) CleanupExpiredSessions() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	expiredCount := 0

	for id, session := range r.sessions {
		if now.After(session.ExpiresAt) {
			delete(r.sessions, id)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		r.logger.Info("Cleaned up expired sessions", zap.Int("count", expiredCount))
	}
}

// generateToken generates a random session token
func generateToken() string {
	// In production, use crypto/rand for secure token generation
	return fmt.Sprintf("tkn_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

// AuthorizeMiddleware creates a middleware function for authorization
func (r *RBACManager) AuthorizeMiddleware(requiredPermission Permission) func(ctx context.Context, userID string) error {
	return func(ctx context.Context, userID string) error {
		if !r.HasPermission(userID, requiredPermission) {
			return fmt.Errorf("access denied: missing permission %s", requiredPermission)
		}
		return nil
	}
}
