package rbac

import (
	"testing"
	"time"
)

func TestHashPassword(t *testing.T) {
	password := "SecurePass1!"
	hash := HashPassword(password)

	// Check that hash is not empty
	if hash == "" {
		t.Error("Expected non-empty hash")
	}

	// Check that hash is different from password
	if hash == password {
		t.Error("Hash should be different from password")
	}

	// Check that same password produces different hashes (bcrypt uses salt)
	hash2 := HashPassword(password)
	if hash == hash2 {
		t.Error("Same password should produce different hashes due to salt")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "SecurePass1!"
	hash := HashPassword(password)

	// Check correct password
	if !CheckPassword(password, hash) {
		t.Error("CheckPassword should return true for correct password")
	}

	// Check wrong password
	if CheckPassword("wrongPassword", hash) {
		t.Error("CheckPassword should return false for wrong password")
	}
}

func TestPasswordPolicy(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"Too short", "Short1!", true},
		{"No uppercase", "password123!", true},
		{"No lowercase", "PASSWORD123!", true},
		{"No digit", "Password!!!!", true},
		{"No special", "Password123", true},
		{"Valid password", "SecurePass1!", false},
		{"Valid with special char", "MyStr0ngP@ss!", false},
		{"Valid complex", "Xy9#mP@ssw0rd", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PasswordPolicy(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("PasswordPolicy(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

func TestRBACEngine_Authenticate(t *testing.T) {
	engine := NewRBACEngine()

	// Test default admin
	user, err := engine.Authenticate("admin", "SecurePass1!")
	if err != nil {
		t.Errorf("Expected successful authentication, got error: %v", err)
	}
	if user.Username != "admin" {
		t.Errorf("Expected username 'admin', got '%s'", user.Username)
	}

	// Test wrong password
	_, err = engine.Authenticate("admin", "wrongpassword")
	if err == nil {
		t.Error("Expected error for wrong password")
	}

	// Test non-existent user
	_, err = engine.Authenticate("nonexistent", "password")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

func TestRBACEngine_CreateUser(t *testing.T) {
	engine := NewRBACEngine()

	// Create valid user
	user, err := engine.CreateUser("testuser", "MyStr0ngP@ss!", "test@example.com", "Test User", RoleBackupUser)
	if err != nil {
		t.Errorf("Expected successful user creation, got error: %v", err)
	}
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}
	if user.Role != RoleBackupUser {
		t.Errorf("Expected role '%s', got '%s'", RoleBackupUser, user.Role)
	}

	// Create duplicate user
	_, err = engine.CreateUser("testuser", "Xy9#mP@ssw0rd!", "another@example.com", "Another User", RoleReadOnly)
	if err == nil {
		t.Error("Expected error for duplicate username")
	}

	// Create user with weak password
	_, err = engine.CreateUser("weakuser", "weak", "weak@example.com", "Weak User", RoleReadOnly)
	if err == nil {
		t.Error("Expected error for weak password")
	}

	// Create user with invalid role
	_, err = engine.CreateUser("invalidroleuser", "SecurePass1!", "invalid@example.com", "Invalid User", "invalid_role")
	if err == nil {
		t.Error("Expected error for invalid role")
	}
}

func TestRBACEngine_DeleteUser(t *testing.T) {
	engine := NewRBACEngine()

	// Try to delete admin
	err := engine.DeleteUser("admin")
	if err == nil {
		t.Error("Expected error when deleting admin user")
	}

	// Create and delete a regular user
	user, _ := engine.CreateUser("tempuser", "Xy9#mP@ssw0rd", "temp@example.com", "Temp User", RoleReadOnly)
	err = engine.DeleteUser(user.ID)
	if err != nil {
		t.Errorf("Expected successful deletion, got error: %v", err)
	}

	// Try to delete already deleted user
	err = engine.DeleteUser(user.ID)
	if err == nil {
		t.Error("Expected error when deleting non-existent user")
	}
}

func TestRBACEngine_Session(t *testing.T) {
	engine := NewRBACEngine()

	// Create session
	session, err := engine.CreateSession("admin", "127.0.0.1", "test-agent")
	if err != nil {
		t.Errorf("Expected successful session creation, got error: %v", err)
	}

	// Validate session
	user, err := engine.ValidateSession(session.Token)
	if err != nil {
		t.Errorf("Expected valid session, got error: %v", err)
	}
	if user.Username != "admin" {
		t.Errorf("Expected admin user, got '%s'", user.Username)
	}

	// Logout
	engine.Logout(session.Token)

	// Validate logged out session
	_, err = engine.ValidateSession(session.Token)
	if err == nil {
		t.Error("Expected error for logged out session")
	}
}

func TestRBACEngine_CheckPermission(t *testing.T) {
	engine := NewRBACEngine()

	// Get admin user
	admin, _ := engine.Authenticate("admin", "SecurePass1!")

	// Admin should have all permissions
	if !engine.CheckPermission(admin, PermUsersCreate) {
		t.Error("Admin should have PermUsersCreate permission")
	}
	if !engine.CheckPermission(admin, PermBackupRun) {
		t.Error("Admin should have PermBackupRun permission")
	}

	// Create read-only user
	readonly, _ := engine.CreateUser("readonly", "Xy9#mP@ssw0rd", "readonly@example.com", "Read Only", RoleReadOnly)

	// Read-only should not have write permissions
	if engine.CheckPermission(readonly, PermUsersCreate) {
		t.Error("ReadOnly should not have PermUsersCreate permission")
	}
	if engine.CheckPermission(readonly, PermBackupRun) {
		t.Error("ReadOnly should not have PermBackupRun permission")
	}

	// Read-only should have read permissions
	if !engine.CheckPermission(readonly, PermBackupRead) {
		t.Error("ReadOnly should have PermBackupRead permission")
	}
}

func TestRBACEngine_ChangePassword(t *testing.T) {
	engine := NewRBACEngine()

	// Create test user
	user, _ := engine.CreateUser("changepasstest", "MyStr0ngP@ss!", "change@example.com", "Change Test", RoleBackupUser)

	// Change password with correct old password
	err := engine.ChangePassword(user.ID, "MyStr0ngP@ss!", "C0mpl3x!Pass")
	if err != nil {
		t.Errorf("Expected successful password change, got error: %v", err)
	}

	// Verify new password works
	newUser, _ := engine.GetUser(user.ID)
	if !CheckPassword("C0mpl3x!Pass", newUser.PasswordHash) {
		t.Error("New password should be set correctly")
	}

	// Try to change with wrong old password
	err = engine.ChangePassword(user.ID, "WrongOldPass", "Xy9#mP@ssw0rd")
	if err == nil {
		t.Error("Expected error for wrong old password")
	}

	// Try to change to weak password
	err = engine.ChangePassword(user.ID, "C0mpl3x!Pass", "weak")
	if err == nil {
		t.Error("Expected error for weak new password")
	}
}

func TestRBACEngine_DisableUser(t *testing.T) {
	engine := NewRBACEngine()

	// Create test user
	user, _ := engine.CreateUser("disabletest", "MyStr0ngP@ss!", "disable@example.com", "Disable Test", RoleBackupUser)

	// Disable user
	err := engine.DisableUser(user.ID)
	if err != nil {
		t.Errorf("Expected successful disable, got error: %v", err)
	}

	// Try to authenticate disabled user
	_, err = engine.Authenticate("disabletest", "MyStr0ngP@ss!")
	if err == nil {
		t.Error("Expected error when authenticating disabled user")
	}

	// Enable user back
	err = engine.EnableUser(user.ID)
	if err != nil {
		t.Errorf("Expected successful enable, got error: %v", err)
	}

	// Now should be able to authenticate
	_, err = engine.Authenticate("disabletest", "MyStr0ngP@ss!")
	if err != nil {
		t.Errorf("Expected successful authentication after enable, got error: %v", err)
	}
}

func TestGetRoleDescription(t *testing.T) {
	tests := []struct {
		role     string
		expected string
	}{
		{RoleAdmin, "Адміністратор - повний доступ до всіх функцій"},
		{RoleBackupAdmin, "Адміністратор резервних копій - управління бекапами"},
		{RoleBackupUser, "Користувач резервних копій - виконання бекапів"},
		{RoleReadOnly, "Тільки читання - перегляд без змін"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			desc := GetRoleDescription(tt.role)
			if desc != tt.expected {
				t.Errorf("GetRoleDescription(%q) = %q, want %q", tt.role, desc, tt.expected)
			}
		})
	}
}

func TestListRoles(t *testing.T) {
	roles := ListRoles()

	expectedRoles := []string{RoleAdmin, RoleBackupAdmin, RoleBackupUser, RoleReadOnly}
	for _, role := range expectedRoles {
		if _, exists := roles[role]; !exists {
			t.Errorf("Expected role %q to exist", role)
		}
	}
}

func TestGetRolePermissions(t *testing.T) {
	adminPerms := GetRolePermissions(RoleAdmin)
	if len(adminPerms) == 0 {
		t.Error("Admin role should have permissions")
	}

	readOnlyPerms := GetRolePermissions(RoleReadOnly)
	if len(readOnlyPerms) >= len(adminPerms) {
		t.Error("ReadOnly role should have fewer permissions than Admin")
	}

	// Unknown role should return empty
	unknownPerms := GetRolePermissions("unknown_role")
	if len(unknownPerms) != 0 {
		t.Error("Unknown role should return empty permissions")
	}
}

func TestAuditEngine(t *testing.T) {
	engine := NewAuditEngine()

	// Log some actions
	engine.Log("user1", "testuser", "login", "/api/auth/login", "127.0.0.1", true, map[string]interface{}{
		"method": "POST",
	})
	engine.Log("user1", "testuser", "create_job", "/api/jobs", "127.0.0.1", true, map[string]interface{}{
		"job_id": "job123",
	})

	// Get logs
	logs := engine.GetLogs(10)
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs, got %d", len(logs))
	}

	// Check log content
	if logs[0].Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", logs[0].Username)
	}
	if logs[0].Success != true {
		t.Error("Expected success to be true")
	}

	// Test limit
	limitedLogs := engine.GetLogs(1)
	if len(limitedLogs) != 1 {
		t.Errorf("Expected 1 log with limit, got %d", len(limitedLogs))
	}
}

func TestSessionExpiration(t *testing.T) {
	engine := NewRBACEngine()

	// Create session
	session, err := engine.CreateSession("admin", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Manually expire the session
	session.ExpiresAt = time.Now().Add(-1 * time.Hour)

	// Try to validate expired session
	_, err = engine.ValidateSession(session.Token)
	if err == nil {
		t.Error("Expected error for expired session")
	}
}

func BenchmarkHashPassword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		HashPassword("SecurePass1!")
	}
}

func BenchmarkCheckPassword(b *testing.B) {
	password := "SecurePass1!"
	hash := HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckPassword(password, hash)
	}
}
