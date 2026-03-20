// Users & Roles - Role-Based Access Control (RBAC)
package rbac

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"novabackup/internal/database"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Roles
const (
	RoleAdmin       = "admin"
	RoleBackupAdmin = "backup_admin"
	RoleBackupUser  = "backup_user"
	RoleReadOnly    = "readonly"
)

// Permissions
const (
	PermBackupCreate   = "backup:create"
	PermBackupRead     = "backup:read"
	PermBackupUpdate   = "backup:update"
	PermBackupDelete   = "backup:delete"
	PermBackupRun      = "backup:run"
	PermRestoreCreate  = "restore:create"
	PermRestoreRead    = "restore:read"
	PermJobsCreate     = "jobs:create"
	PermJobsRead       = "jobs:read"
	PermJobsUpdate     = "jobs:update"
	PermJobsDelete     = "jobs:delete"
	PermSettingsRead   = "settings:read"
	PermSettingsUpdate = "settings:update"
	PermUsersCreate    = "users:create"
	PermUsersRead      = "users:read"
	PermUsersUpdate    = "users:update"
	PermUsersDelete    = "users:delete"
	PermLogsRead       = "logs:read"
)

// RolePermissions defines permissions for each role
var RolePermissions = map[string][]string{
	RoleAdmin: {
		PermBackupCreate, PermBackupRead, PermBackupUpdate, PermBackupDelete, PermBackupRun,
		PermRestoreCreate, PermRestoreRead,
		PermJobsCreate, PermJobsRead, PermJobsUpdate, PermJobsDelete,
		PermSettingsRead, PermSettingsUpdate,
		PermUsersCreate, PermUsersRead, PermUsersUpdate, PermUsersDelete,
		PermLogsRead,
	},
	RoleBackupAdmin: {
		PermBackupCreate, PermBackupRead, PermBackupUpdate, PermBackupRun,
		PermRestoreCreate, PermRestoreRead,
		PermJobsCreate, PermJobsRead, PermJobsUpdate,
		PermSettingsRead,
		PermLogsRead,
	},
	RoleBackupUser: {
		PermBackupRead, PermBackupRun,
		PermRestoreRead,
		PermJobsRead,
	},
	RoleReadOnly: {
		PermBackupRead,
		PermRestoreRead,
		PermJobsRead,
		PermSettingsRead,
		PermLogsRead,
	},
}

// User represents a system user
type User struct {
	ID              string     `json:"id"`
	Username        string     `json:"username"`
	PasswordHash    string     `json:"-"`
	Email           string     `json:"email"`
	FullName        string     `json:"full_name"`
	Role            string     `json:"role"`
	Enabled         bool       `json:"enabled"`
	CreatedAt       time.Time  `json:"created_at"`
	LastLogin       *time.Time `json:"last_login,omitempty"`
	PasswordExpires *time.Time `json:"password_expires,omitempty"`
}

// Session represents an active user session
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	LastUsed  time.Time `json:"last_used"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
}

// RBACEngine handles authentication and authorization
type RBACEngine struct {
	users    map[string]*User
	sessions map[string]*Session
	mu       sync.RWMutex
	DB       interface {
		CreateUserSession(*database.UserSession) error
		GetUserSessionByToken(string) (*database.UserSession, error)
		UpdateUserSessionLastUsed(string, time.Time) error
		DeleteUserSession(string) error
		CreateUser(*database.User) error
		UpdateUser(*database.User) error
		UpdateUserPassword(string, string, *time.Time) error
		UpdateUserLastLogin(string, time.Time) error
		DeleteUser(string) error
		GetUserByID(string) (*database.User, error)
		GetUserByUsername(string) (*database.User, error)
		ListUsers() ([]database.User, error)
		CountUsers() (int, error)
	}
}

// NewRBACEngine creates a new RBAC engine
func NewRBACEngine() *RBACEngine {
	engine := &RBACEngine{
		users:    make(map[string]*User),
		sessions: make(map[string]*Session),
	}

	// Create default admin user
	engine.CreateDefaultAdmin()

	return engine
}

// LoadUsersFromDB replaces in-memory users with database users (if DB is available).
// If no users exist, it seeds the default admin and persists it.
func (e *RBACEngine) LoadUsersFromDB() error {
	if e.DB == nil {
		return nil
	}

	count, err := e.DB.CountUsers()
	if err != nil {
		return err
	}

	if count == 0 {
		// Seed default admin in DB and memory
		admin := &User{
			ID:           "admin",
			Username:     "admin",
			PasswordHash: HashPassword("admin123"),
			Email:        "admin@novabackup.local",
			FullName:     "Administrator",
			Role:         RoleAdmin,
			Enabled:      true,
			CreatedAt:    time.Now(),
		}

		e.mu.Lock()
		e.users = map[string]*User{admin.ID: admin}
		e.mu.Unlock()

		_ = e.DB.CreateUser(&database.User{
			ID:           admin.ID,
			Username:     admin.Username,
			PasswordHash: admin.PasswordHash,
			Email:        admin.Email,
			FullName:     admin.FullName,
			Role:         admin.Role,
			Enabled:      admin.Enabled,
			CreatedAt:    admin.CreatedAt,
		})
		return nil
	}

	users, err := e.DB.ListUsers()
	if err != nil {
		return err
	}

	e.mu.Lock()
	e.users = make(map[string]*User, len(users))
	for _, u := range users {
		user := u
		e.users[user.ID] = &User{
			ID:              user.ID,
			Username:        user.Username,
			PasswordHash:    user.PasswordHash,
			Email:           user.Email,
			FullName:        user.FullName,
			Role:            user.Role,
			Enabled:         user.Enabled,
			CreatedAt:       user.CreatedAt,
			LastLogin:       user.LastLogin,
			PasswordExpires: user.PasswordExpires,
		}
	}
	e.mu.Unlock()

	return nil
}

// CreateDefaultAdmin creates the default admin user
func (e *RBACEngine) CreateDefaultAdmin() {
	admin := &User{
		ID:           "admin",
		Username:     "admin",
		PasswordHash: HashPassword("admin123"),
		Email:        "admin@novabackup.local",
		FullName:     "Administrator",
		Role:         RoleAdmin,
		Enabled:      true,
		CreatedAt:    time.Now(),
	}
	e.users[admin.ID] = admin
}

// Authenticate authenticates a user
func (e *RBACEngine) Authenticate(username, password string) (*User, error) {
	e.mu.RLock()
	var user *User
	for _, u := range e.users {
		if u.Username == username {
			user = u
			break
		}
	}
	e.mu.RUnlock()

	if user == nil {
		return nil, errors.New("користувача не знайдено")
	}

	if !user.Enabled {
		return nil, errors.New("користувача вимкнено")
	}

	if !CheckPassword(password, user.PasswordHash) {
		return nil, errors.New("невірний пароль")
	}

	now := time.Now()
	e.mu.Lock()
	if stored, ok := e.users[user.ID]; ok {
		stored.LastLogin = &now
	}
	e.mu.Unlock()

	if e.DB != nil {
		_ = e.DB.UpdateUserLastLogin(user.ID, now)
	}

	return user, nil
}

// CreateSession creates a new session for a user
func (e *RBACEngine) CreateSession(userID, ipAddress, userAgent string) (*Session, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	_, exists := e.users[userID]
	if !exists {
		return nil, errors.New("користувача не знайдено")
	}

	session := &Session{
		ID:        generateSessionID(),
		UserID:    userID,
		Token:     generateToken(),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour session
		LastUsed:  time.Now(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	// Store in memory by token (for lookup by token)
	e.sessions[session.Token] = session

	// Also store in database if available
	if e.DB != nil {
		dbSession := &database.UserSession{
			ID:        session.ID,
			UserID:    session.UserID,
			Token:     session.Token,
			CreatedAt: session.CreatedAt,
			ExpiresAt: session.ExpiresAt,
			LastUsed:  session.LastUsed,
			IPAddress: session.IPAddress,
			UserAgent: session.UserAgent,
		}
		_ = e.DB.CreateUserSession(dbSession) // Ignore error if DB fails
	}

	return session, nil
}

// ValidateSession validates a session token
func (e *RBACEngine) ValidateSession(token string) (*User, error) {
	e.mu.RLock()
	session, exists := e.sessions[token]
	e.mu.RUnlock()

	// Check memory first
	if exists {
		e.mu.Lock()
		defer e.mu.Unlock()

		if time.Now().After(session.ExpiresAt) {
			return nil, errors.New("сесія закінчилася")
		}

		user, userExists := e.users[session.UserID]
		if !userExists {
			return nil, errors.New("користувача не знайдено")
		}

		if !user.Enabled {
			return nil, errors.New("користувача вимкнено")
		}

		session.LastUsed = time.Now()
		return user, nil
	}

	// If not in memory, try database (for persistence across restarts)
	if e.DB != nil {
		dbSession, err := e.DB.GetUserSessionByToken(token)
		if err == nil {
			// Found in database
			if time.Now().After(dbSession.ExpiresAt) {
				return nil, errors.New("сесія закінчилася")
			}

			user, userExists := e.users[dbSession.UserID]
			if !userExists {
				return nil, errors.New("користувача не знайдено")
			}

			if !user.Enabled {
				return nil, errors.New("користувача вимкнено")
			}

			// Update last used
			_ = e.DB.UpdateUserSessionLastUsed(token, time.Now())

			// Recreate in-memory session
			newSession := &Session{
				ID:        dbSession.ID,
				UserID:    dbSession.UserID,
				Token:     dbSession.Token,
				CreatedAt: dbSession.CreatedAt,
				ExpiresAt: dbSession.ExpiresAt,
				LastUsed:  dbSession.LastUsed,
				IPAddress: dbSession.IPAddress,
				UserAgent: dbSession.UserAgent,
			}

			e.mu.Lock()
			e.sessions[token] = newSession
			e.mu.Unlock()

			return user, nil
		}
	}

	return nil, errors.New("недійсна сесія")
}

// Logout invalidates a session
func (e *RBACEngine) Logout(token string) {
	e.mu.Lock()
	// Delete by token directly (since we store by token now)
	delete(e.sessions, token)
	e.mu.Unlock()

	// Also delete from database if available
	if e.DB != nil {
		_ = e.DB.DeleteUserSession(token)
	}
}

// CheckPermission checks if a user has a specific permission
func (e *RBACEngine) CheckPermission(user *User, permission string) bool {
	permissions, exists := RolePermissions[user.Role]
	if !exists {
		return false
	}

	for _, perm := range permissions {
		if perm == permission {
			return true
		}
	}

	return false
}

// CheckPermissions checks if a user has all specified permissions
func (e *RBACEngine) CheckPermissions(user *User, permissions []string) bool {
	for _, perm := range permissions {
		if !e.CheckPermission(user, perm) {
			return false
		}
	}
	return true
}

// CreateUser creates a new user
func (e *RBACEngine) CreateUser(username, password, email, fullName, role string) (*User, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check if username already exists
	for _, user := range e.users {
		if user.Username == username {
			return nil, errors.New("користувач вже існує")
		}
	}

	// Validate role
	if _, exists := RolePermissions[role]; !exists {
		return nil, errors.New("невідома роль")
	}

	// Validate password with policy
	if err := PasswordPolicy(password); err != nil {
		return nil, err
	}

	user := &User{
		ID:           generateUserID(),
		Username:     username,
		PasswordHash: HashPassword(password),
		Email:        email,
		FullName:     fullName,
		Role:         role,
		Enabled:      true,
		CreatedAt:    time.Now(),
	}

	e.users[user.ID] = user

	if e.DB != nil {
		if err := e.DB.CreateUser(&database.User{
			ID:           user.ID,
			Username:     user.Username,
			PasswordHash: user.PasswordHash,
			Email:        user.Email,
			FullName:     user.FullName,
			Role:         user.Role,
			Enabled:      user.Enabled,
			CreatedAt:    user.CreatedAt,
		}); err != nil {
			delete(e.users, user.ID)
			return nil, err
		}
	}

	return user, nil
}

// UpdateUser updates a user
func (e *RBACEngine) UpdateUser(userID, email, fullName, role string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	user, exists := e.users[userID]
	if !exists {
		return errors.New("користувача не знайдено")
	}

	user.Email = email
	user.FullName = fullName
	user.Role = role

	if e.DB != nil {
		return e.DB.UpdateUser(&database.User{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
			Role:     user.Role,
			Enabled:  user.Enabled,
		})
	}

	return nil
}

// DeleteUser deletes a user
func (e *RBACEngine) DeleteUser(userID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if userID == "admin" {
		return errors.New("не можна видалити адміністратора")
	}

	_, exists := e.users[userID]
	if !exists {
		return errors.New("користувача не знайдено")
	}

	delete(e.users, userID)

	// Invalidate all user sessions
	for id, session := range e.sessions {
		if session.UserID == userID {
			delete(e.sessions, id)
		}
	}

	if e.DB != nil {
		return e.DB.DeleteUser(userID)
	}

	return nil
}

// GetUser returns a user by ID
func (e *RBACEngine) GetUser(userID string) (*User, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	user, exists := e.users[userID]
	if !exists {
		return nil, errors.New("користувача не знайдено")
	}

	return user, nil
}

// ListUsers returns all users
func (e *RBACEngine) ListUsers() []*User {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var users []*User
	for _, user := range e.users {
		users = append(users, user)
	}
	return users
}

// ChangePassword changes a user's password
func (e *RBACEngine) ChangePassword(userID, oldPassword, newPassword string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	user, exists := e.users[userID]
	if !exists {
		return errors.New("користувача не знайдено")
	}

	// Verify old password (except for admin creating users)
	if !CheckPassword(oldPassword, user.PasswordHash) {
		return errors.New("невірний старий пароль")
	}

	// Validate new password with policy
	if err := PasswordPolicy(newPassword); err != nil {
		return err
	}

	user.PasswordHash = HashPassword(newPassword)

	// Set password expiration (90 days)
	expires := time.Now().Add(90 * 24 * time.Hour)
	user.PasswordExpires = &expires

	if e.DB != nil {
		return e.DB.UpdateUserPassword(user.ID, user.PasswordHash, &expires)
	}

	return nil
}

// EnableUser enables a user
func (e *RBACEngine) EnableUser(userID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	user, exists := e.users[userID]
	if !exists {
		return errors.New("користувача не знайдено")
	}

	user.Enabled = true

	if e.DB != nil {
		return e.DB.UpdateUser(&database.User{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
			Role:     user.Role,
			Enabled:  user.Enabled,
		})
	}

	return nil
}

// DisableUser disables a user
func (e *RBACEngine) DisableUser(userID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	user, exists := e.users[userID]
	if !exists {
		return errors.New("користувача не знайдено")
	}

	user.Enabled = false

	// Invalidate all user sessions
	for id, session := range e.sessions {
		if session.UserID == userID {
			delete(e.sessions, id)
		}
	}

	if e.DB != nil {
		return e.DB.UpdateUser(&database.User{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
			Role:     user.Role,
			Enabled:  user.Enabled,
		})
	}

	return nil
}

// GetRolePermissions returns permissions for a role
func GetRolePermissions(role string) []string {
	if perms, exists := RolePermissions[role]; exists {
		return perms
	}
	return []string{}
}

// ListRoles returns all available roles
func ListRoles() map[string][]string {
	return RolePermissions
}

// GetRoleDescription returns human-readable role description
func GetRoleDescription(role string) string {
	descriptions := map[string]string{
		RoleAdmin:       "Адміністратор - повний доступ до всіх функцій",
		RoleBackupAdmin: "Адміністратор резервних копій - управління бекапами",
		RoleBackupUser:  "Користувач резервних копій - виконання бекапів",
		RoleReadOnly:    "Тільки читання - перегляд без змін",
	}
	if desc, exists := descriptions[role]; exists {
		return desc
	}
	return role
}

// Helper functions

// HashPassword creates a bcrypt hash of a password
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// Fallback to SHA-256 in case of error (should never happen)
		hash := sha256.Sum256([]byte(password))
		return hex.EncodeToString(hash[:])
	}
	return string(bytes)
}

// CheckPassword verifies a password against a bcrypt hash
func CheckPassword(password, hash string) bool {
	// Try bcrypt first
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {
		return true
	}

	// Fallback to SHA-256 for backward compatibility with old hashes
	shaHash := HashPasswordSHA256(password)
	return shaHash == hash
}

// HashPasswordSHA256 creates a SHA-256 hash (for backward compatibility only)
func HashPasswordSHA256(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func generateUserID() string {
	return fmt.Sprintf("user_%d", time.Now().UnixNano())
}

func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

func generateToken() string {
	// Use crypto/rand for secure random token
	bytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		// Fallback to time-based (less secure, but better than nothing)
		hash := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
		return hex.EncodeToString(hash[:])
	}
	return hex.EncodeToString(bytes)
}

// PasswordPolicy validates password strength
func PasswordPolicy(password string) error {
	if len(password) < 8 {
		return errors.New("пароль має бути не менше 8 символів")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return errors.New("пароль має містити великі літери, малі літери та цифри")
	}

	return nil
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    string                 `json:"user_id"`
	Username  string                 `json:"username"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Details   map[string]interface{} `json:"details,omitempty"`
	IPAddress string                 `json:"ip_address"`
	Success   bool                   `json:"success"`
}

// AuditEngine handles audit logging
type AuditEngine struct {
	logs []AuditLog
	mu   sync.Mutex
}

// NewAuditEngine creates a new audit engine
func NewAuditEngine() *AuditEngine {
	return &AuditEngine{
		logs: make([]AuditLog, 0),
	}
}

// Log adds an audit log entry
func (e *AuditEngine) Log(userID, username, action, resource, ipAddress string, success bool, details map[string]interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()

	log := AuditLog{
		ID:        fmt.Sprintf("audit_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		UserID:    userID,
		Username:  username,
		Action:    action,
		Resource:  resource,
		IPAddress: ipAddress,
		Success:   success,
		Details:   details,
	}

	e.logs = append(e.logs, log)

	// Keep only last 10000 entries
	if len(e.logs) > 10000 {
		e.logs = e.logs[len(e.logs)-10000:]
	}
}

// GetLogs returns audit logs
func (e *AuditEngine) GetLogs(limit int) []AuditLog {
	e.mu.Lock()
	defer e.mu.Unlock()

	if limit <= 0 || limit > len(e.logs) {
		return e.logs
	}

	return e.logs[len(e.logs)-limit:]
}
