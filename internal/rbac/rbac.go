// Package rbac provides Role-Based Access Control (RBAC) functionality for NovaBackup.
// It manages roles, permissions, and user assignments with thread-safe operations.
package rbac

import (
	"errors"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Resource constants define the available resources that can be protected by RBAC.
const (
	ResourceJobs        = "jobs"
	ResourceBackups     = "backups"
	ResourceStorage     = "storage"
	ResourceSettings    = "settings"
	ResourceUsers       = "users"
	ResourceRoles       = "roles"
	ResourceReplication = "replication"
	ResourceRestore     = "restore"
)

// Action constants define the available actions that can be performed on resources.
const (
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

// Scope constants define the scope of a permission.
const (
	ScopeGlobal = "global" // Access to all resources of the type
	ScopeOwn    = "own"    // Access only to resources owned by the user
	ScopeTeam   = "team"   // Access to resources within the user's team
)

// Predefined role IDs
const (
	RoleAdminID       = "admin"
	RoleOperatorID    = "operator"
	RoleViewerID      = "viewer"
	RoleBackupAdminID = "backup_admin"
)

// Common errors
var (
	ErrRoleNotFound      = errors.New("role not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrRoleAlreadyExists = errors.New("role already exists")
	ErrInvalidPermission = errors.New("invalid permission")
)

// Permission represents a single permission grant for a resource and action.
type Permission struct {
	// Resource is the target resource (e.g., "jobs", "backups", "storage")
	Resource string `json:"resource"`
	// Action is the allowed action (e.g., "create", "read", "update", "delete")
	Action string `json:"action"`
	// Scope defines the scope of the permission (e.g., "global", "own", "team")
	Scope string `json:"scope"`
}

// IsValid checks if the permission has valid values.
func (p *Permission) IsValid() bool {
	if p.Resource == "" || p.Action == "" {
		return false
	}
	switch p.Action {
	case ActionCreate, ActionRead, ActionUpdate, ActionDelete:
	default:
		return false
	}
	if p.Scope != "" {
		switch p.Scope {
		case ScopeGlobal, ScopeOwn, ScopeTeam:
		default:
			return false
		}
	}
	return true
}

// Role represents a collection of permissions that can be assigned to users.
type Role struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// HasPermission checks if the role has a specific permission.
func (r *Role) HasPermission(resource, action string) bool {
	for _, perm := range r.Permissions {
		if perm.Resource == resource && perm.Action == action {
			return true
		}
	}
	return false
}

// AddPermission adds a permission to the role if it doesn't already exist.
func (r *Role) AddPermission(p Permission) {
	if !r.HasPermission(p.Resource, p.Action) {
		r.Permissions = append(r.Permissions, p)
		r.UpdatedAt = time.Now()
	}
}

// RemovePermission removes a permission from the role.
func (r *Role) RemovePermission(resource, action string) {
	for i, perm := range r.Permissions {
		if perm.Resource == resource && perm.Action == action {
			r.Permissions = append(r.Permissions[:i], r.Permissions[i+1:]...)
			r.UpdatedAt = time.Now()
			return
		}
	}
}

// User represents a user in the system with assigned roles.
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	RoleIDs   []string  `json:"role_ids"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// HasRole checks if the user has a specific role assigned.
func (u *User) HasRole(roleID string) bool {
	for _, rid := range u.RoleIDs {
		if rid == roleID {
			return true
		}
	}
	return false
}

// AddRole adds a role to the user if not already assigned.
func (u *User) AddRole(roleID string) {
	if !u.HasRole(roleID) {
		u.RoleIDs = append(u.RoleIDs, roleID)
		u.UpdatedAt = time.Now()
	}
}

// RemoveRole removes a role from the user.
func (u *User) RemoveRole(roleID string) {
	for i, rid := range u.RoleIDs {
		if rid == roleID {
			u.RoleIDs = append(u.RoleIDs[:i], u.RoleIDs[i+1:]...)
			u.UpdatedAt = time.Now()
			return
		}
	}
}

// RBACManager manages roles, users, and permissions with thread-safe access.
type RBACManager struct {
	roles  map[string]*Role
	users  map[string]*User
	mu     sync.RWMutex
	logger *zap.Logger
}

// NewRBACManager creates a new RBAC manager with predefined roles.
func NewRBACManager(logger *zap.Logger) *RBACManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	manager := &RBACManager{
		roles:  make(map[string]*Role),
		users:  make(map[string]*User),
		logger: logger,
	}

	manager.initializePredefinedRoles()

	return manager
}

// initializePredefinedRoles creates the standard set of roles.
func (m *RBACManager) initializePredefinedRoles() {
	// Admin role - full access to all resources
	adminRole := &Role{
		ID:          RoleAdminID,
		Name:        "Administrator",
		Description: "Full system access with all permissions",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	resources := []string{
		ResourceJobs, ResourceBackups, ResourceStorage,
		ResourceSettings, ResourceUsers, ResourceRoles,
		ResourceReplication, ResourceRestore,
	}
	actions := []string{ActionCreate, ActionRead, ActionUpdate, ActionDelete}
	for _, resource := range resources {
		for _, action := range actions {
			adminRole.AddPermission(Permission{
				Resource: resource,
				Action:   action,
				Scope:    ScopeGlobal,
			})
		}
	}
	m.roles[RoleAdminID] = adminRole

	// Operator role - manage backups, no delete
	operatorRole := &Role{
		ID:          RoleOperatorID,
		Name:        "Operator",
		Description: "Can manage backups and jobs but cannot delete resources",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	for _, resource := range []string{ResourceBackups, ResourceJobs} {
		for _, action := range []string{ActionCreate, ActionRead, ActionUpdate} {
			operatorRole.AddPermission(Permission{
				Resource: resource,
				Action:   action,
				Scope:    ScopeGlobal,
			})
		}
	}
	operatorRole.AddPermission(Permission{Resource: ResourceStorage, Action: ActionRead, Scope: ScopeGlobal})
	operatorRole.AddPermission(Permission{Resource: ResourceSettings, Action: ActionRead, Scope: ScopeGlobal})
	operatorRole.AddPermission(Permission{Resource: ResourceRestore, Action: ActionRead, Scope: ScopeGlobal})
	m.roles[RoleOperatorID] = operatorRole

	// Viewer role - read-only access
	viewerRole := &Role{
		ID:          RoleViewerID,
		Name:        "Viewer",
		Description: "Read-only access to all resources",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	for _, resource := range resources {
		viewerRole.AddPermission(Permission{
			Resource: resource,
			Action:   ActionRead,
			Scope:    ScopeGlobal,
		})
	}
	m.roles[RoleViewerID] = viewerRole

	// BackupAdmin role - backup operations only
	backupAdminRole := &Role{
		ID:          RoleBackupAdminID,
		Name:        "Backup Administrator",
		Description: "Full access to backup operations only",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	for _, action := range actions {
		backupAdminRole.AddPermission(Permission{
			Resource: ResourceBackups,
			Action:   action,
			Scope:    ScopeGlobal,
		})
	}
	backupAdminRole.AddPermission(Permission{Resource: ResourceStorage, Action: ActionRead, Scope: ScopeGlobal})
	backupAdminRole.AddPermission(Permission{Resource: ResourceJobs, Action: ActionRead, Scope: ScopeGlobal})
	backupAdminRole.AddPermission(Permission{Resource: ResourceRestore, Action: ActionRead, Scope: ScopeGlobal})
	m.roles[RoleBackupAdminID] = backupAdminRole

	m.logger.Info("Predefined roles initialized", zap.Int("count", len(m.roles)))
}

// CreateRole creates a new role with the given ID and name.
func (m *RBACManager) CreateRole(id, name, description string, permissions []Permission) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.roles[id]; exists {
		m.logger.Warn("Failed to create role - already exists", zap.String("role_id", id))
		return ErrRoleAlreadyExists
	}

	for _, p := range permissions {
		if !p.IsValid() {
			m.logger.Warn("Failed to create role - invalid permission",
				zap.String("role_id", id),
				zap.Any("permission", p))
			return ErrInvalidPermission
		}
	}

	role := &Role{
		ID:          id,
		Name:        name,
		Description: description,
		Permissions: permissions,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	m.roles[id] = role
	m.logger.Info("Role created",
		zap.String("role_id", id),
		zap.String("name", name),
		zap.Int("permissions_count", len(permissions)))

	return nil
}

// GetRole retrieves a role by its ID.
func (m *RBACManager) GetRole(id string) (*Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	role, exists := m.roles[id]
	if !exists {
		m.logger.Debug("Role not found", zap.String("role_id", id))
		return nil, ErrRoleNotFound
	}

	roleCopy := *role
	roleCopy.Permissions = make([]Permission, len(role.Permissions))
	copy(roleCopy.Permissions, role.Permissions)

	return &roleCopy, nil
}

// UpdateRole updates an existing role's details.
func (m *RBACManager) UpdateRole(id, name, description string, permissions []Permission) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	role, exists := m.roles[id]
	if !exists {
		m.logger.Warn("Failed to update role - not found", zap.String("role_id", id))
		return ErrRoleNotFound
	}

	for _, p := range permissions {
		if !p.IsValid() {
			m.logger.Warn("Failed to update role - invalid permission",
				zap.String("role_id", id),
				zap.Any("permission", p))
			return ErrInvalidPermission
		}
	}

	role.Name = name
	role.Description = description
	role.Permissions = permissions
	role.UpdatedAt = time.Now()

	m.logger.Info("Role updated",
		zap.String("role_id", id),
		zap.String("name", name),
		zap.Int("permissions_count", len(permissions)))

	return nil
}

// DeleteRole removes a role from the system.
func (m *RBACManager) DeleteRole(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.roles[id]; !exists {
		m.logger.Warn("Failed to delete role - not found", zap.String("role_id", id))
		return ErrRoleNotFound
	}

	switch id {
	case RoleAdminID, RoleOperatorID, RoleViewerID, RoleBackupAdminID:
		m.logger.Warn("Cannot delete predefined role", zap.String("role_id", id))
		return errors.New("cannot delete predefined role")
	}

	for _, user := range m.users {
		user.RemoveRole(id)
	}

	delete(m.roles, id)
	m.logger.Info("Role deleted", zap.String("role_id", id))

	return nil
}

// AssignRoleToUser assigns a role to a user.
func (m *RBACManager) AssignRoleToUser(userID, roleID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[userID]
	if !exists {
		m.logger.Warn("Failed to assign role - user not found",
			zap.String("user_id", userID),
			zap.String("role_id", roleID))
		return ErrUserNotFound
	}

	role, exists := m.roles[roleID]
	if !exists {
		m.logger.Warn("Failed to assign role - role not found",
			zap.String("user_id", userID),
			zap.String("role_id", roleID))
		return ErrRoleNotFound
	}

	user.AddRole(roleID)
	m.logger.Info("Role assigned to user",
		zap.String("user_id", userID),
		zap.String("role_id", roleID),
		zap.String("role_name", role.Name))

	return nil
}

// RemoveRoleFromUser removes a role assignment from a user.
func (m *RBACManager) RemoveRoleFromUser(userID, roleID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[userID]
	if !exists {
		m.logger.Warn("Failed to remove role - user not found",
			zap.String("user_id", userID),
			zap.String("role_id", roleID))
		return ErrUserNotFound
	}

	user.RemoveRole(roleID)
	m.logger.Info("Role removed from user",
		zap.String("user_id", userID),
		zap.String("role_id", roleID))

	return nil
}

// CheckPermission checks if a user has a specific permission through any of their roles.
func (m *RBACManager) CheckPermission(userID, resource, action string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[userID]
	if !exists {
		m.logger.Debug("Permission check failed - user not found",
			zap.String("user_id", userID),
			zap.String("resource", resource),
			zap.String("action", action))
		return false
	}

	if !user.Active {
		m.logger.Debug("Permission check failed - user inactive",
			zap.String("user_id", userID),
			zap.String("resource", resource),
			zap.String("action", action))
		return false
	}

	for _, roleID := range user.RoleIDs {
		role, exists := m.roles[roleID]
		if !exists {
			continue
		}
		if role.HasPermission(resource, action) {
			m.logger.Debug("Permission granted",
				zap.String("user_id", userID),
				zap.String("resource", resource),
				zap.String("action", action),
				zap.String("role_id", roleID))
			return true
		}
	}

	m.logger.Debug("Permission denied",
		zap.String("user_id", userID),
		zap.String("resource", resource),
		zap.String("action", action))
	return false
}

// HasAccess is an alias for CheckPermission for more intuitive API usage.
func (m *RBACManager) HasAccess(userID, resource, action string) bool {
	return m.CheckPermission(userID, resource, action)
}

// ListRoles returns a list of all roles.
func (m *RBACManager) ListRoles() []*Role {
	m.mu.RLock()
	defer m.mu.RUnlock()

	roles := make([]*Role, 0, len(m.roles))
	for _, role := range m.roles {
		roleCopy := *role
		roleCopy.Permissions = make([]Permission, len(role.Permissions))
		copy(roleCopy.Permissions, role.Permissions)
		roles = append(roles, &roleCopy)
	}

	return roles
}

// ListUsers returns a list of all users.
func (m *RBACManager) ListUsers() []*User {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]*User, 0, len(m.users))
	for _, user := range m.users {
		userCopy := *user
		userCopy.RoleIDs = make([]string, len(user.RoleIDs))
		copy(userCopy.RoleIDs, user.RoleIDs)
		users = append(users, &userCopy)
	}

	return users
}

// CreateUser creates a new user in the system.
func (m *RBACManager) CreateUser(id, username, email string) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[id]; exists {
		m.logger.Warn("Failed to create user - already exists", zap.String("user_id", id))
		return nil, errors.New("user already exists")
	}

	user := &User{
		ID:        id,
		Username:  username,
		Email:     email,
		RoleIDs:   []string{},
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	m.users[id] = user
	m.logger.Info("User created",
		zap.String("user_id", id),
		zap.String("username", username),
		zap.String("email", email))

	return user, nil
}

// GetUser retrieves a user by their ID.
func (m *RBACManager) GetUser(id string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[id]
	if !exists {
		m.logger.Debug("User not found", zap.String("user_id", id))
		return nil, ErrUserNotFound
	}

	userCopy := *user
	userCopy.RoleIDs = make([]string, len(user.RoleIDs))
	copy(userCopy.RoleIDs, user.RoleIDs)

	return &userCopy, nil
}

// UpdateUser updates a user's details.
func (m *RBACManager) UpdateUser(id, username, email string, active bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[id]
	if !exists {
		m.logger.Warn("Failed to update user - not found", zap.String("user_id", id))
		return ErrUserNotFound
	}

	user.Username = username
	user.Email = email
	user.Active = active
	user.UpdatedAt = time.Now()

	m.logger.Info("User updated",
		zap.String("user_id", id),
		zap.String("username", username),
		zap.Bool("active", active))

	return nil
}

// DeleteUser removes a user from the system.
func (m *RBACManager) DeleteUser(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[id]; !exists {
		m.logger.Warn("Failed to delete user - not found", zap.String("user_id", id))
		return ErrUserNotFound
	}

	delete(m.users, id)
	m.logger.Info("User deleted", zap.String("user_id", id))

	return nil
}

// SetUserActive sets the active status of a user.
func (m *RBACManager) SetUserActive(id string, active bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[id]
	if !exists {
		m.logger.Warn("Failed to set user active status - not found", zap.String("user_id", id))
		return ErrUserNotFound
	}

	user.Active = active
	user.UpdatedAt = time.Now()

	m.logger.Info("User active status updated",
		zap.String("user_id", id),
		zap.Bool("active", active))

	return nil
}

// GetUserRoles returns all roles assigned to a user.
func (m *RBACManager) GetUserRoles(userID string) ([]*Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[userID]
	if !exists {
		m.logger.Debug("User not found for role lookup", zap.String("user_id", userID))
		return nil, ErrUserNotFound
	}

	roles := make([]*Role, 0, len(user.RoleIDs))
	for _, roleID := range user.RoleIDs {
		role, exists := m.roles[roleID]
		if !exists {
			continue
		}
		roleCopy := *role
		roleCopy.Permissions = make([]Permission, len(role.Permissions))
		copy(roleCopy.Permissions, role.Permissions)
		roles = append(roles, &roleCopy)
	}

	return roles, nil
}

// GetRoleByName finds a role by its name (case-insensitive).
func (m *RBACManager) GetRoleByName(name string) (*Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	nameLower := strings.ToLower(name)
	for _, role := range m.roles {
		if strings.ToLower(role.Name) == nameLower {
			roleCopy := *role
			roleCopy.Permissions = make([]Permission, len(role.Permissions))
			copy(roleCopy.Permissions, role.Permissions)
			return &roleCopy, nil
		}
	}

	return nil, ErrRoleNotFound
}

// HasAnyPermission checks if a user has any of the specified permissions.
func (m *RBACManager) HasAnyPermission(userID, resource string, actions ...string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[userID]
	if !exists {
		return false
	}

	if !user.Active {
		return false
	}

	for _, roleID := range user.RoleIDs {
		role, exists := m.roles[roleID]
		if !exists {
			continue
		}
		for _, action := range actions {
			if role.HasPermission(resource, action) {
				return true
			}
		}
	}

	return false
}

// GetAllPermissions returns all unique permissions a user has across all roles.
func (m *RBACManager) GetAllPermissions(userID string) []Permission {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[userID]
	if !exists {
		return nil
	}

	permMap := make(map[string]Permission)
	keyFunc := func(p Permission) string {
		return p.Resource + ":" + p.Action + ":" + p.Scope
	}

	for _, roleID := range user.RoleIDs {
		role, exists := m.roles[roleID]
		if !exists {
			continue
		}
		for _, perm := range role.Permissions {
			permMap[keyFunc(perm)] = perm
		}
	}

	permissions := make([]Permission, 0, len(permMap))
	for _, perm := range permMap {
		permissions = append(permissions, perm)
	}

	return permissions
} 
