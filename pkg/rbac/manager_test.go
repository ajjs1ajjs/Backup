package rbac

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestRBACManager_CreateUser(t *testing.T) {
	logger := zap.NewNop()
	manager := NewRBACManager(logger)

	user := &User{
		ID:       "user_001",
		Username: "testuser",
		Email:    "test@example.com",
		Active:   true,
	}

	err := manager.CreateUser(user)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	retrieved, err := manager.GetUser("user_001")
	if err != nil {
		t.Errorf("Expected user to exist, got error: %v", err)
	}
	if retrieved.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", retrieved.Username)
	}
}

func TestRBACManager_CreateRole(t *testing.T) {
	logger := zap.NewNop()
	manager := NewRBACManager(logger)

	role := &Role{
		ID:          "role_custom",
		Name:        "Custom Role",
		Description: "A custom role",
		Permissions: []Permission{PermJobRead, PermBackupRead},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := manager.CreateRole(role)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	retrieved, err := manager.GetRole("role_custom")
	if err != nil {
		t.Errorf("Expected role to exist, got error: %v", err)
	}
	if retrieved.Name != "Custom Role" {
		t.Errorf("Expected role name 'Custom Role', got '%s'", retrieved.Name)
	}
}

func TestRBACManager_AssignRoleToUser(t *testing.T) {
	logger := zap.NewNop()
	manager := NewRBACManager(logger)

	user := &User{
		ID:       "user_002",
		Username: "operator",
		Email:    "op@example.com",
		Active:   true,
	}
	manager.CreateUser(user)

	err := manager.AssignRoleToUser("user_002", "role_operator")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	perms, err := manager.GetUserPermissions("user_002")
	if err != nil {
		t.Errorf("Expected no error getting permissions, got %v", err)
	}
	if len(perms) == 0 {
		t.Error("Expected permissions to be assigned")
	}
}

func TestRBACManager_HasPermission(t *testing.T) {
	logger := zap.NewNop()
	manager := NewRBACManager(logger)

	user := &User{
		ID:       "user_003",
		Username: "admin",
		Email:    "admin@example.com",
		Active:   true,
	}
	manager.CreateUser(user)
	manager.AssignRoleToUser("user_003", "role_admin")

	if !manager.HasPermission("user_003", PermJobCreate) {
		t.Error("Admin user should have job:create permission")
	}

	if !manager.HasPermission("user_003", PermBackupRestore) {
		t.Error("Admin user should have backup:restore permission")
	}
}

func TestRBACManager_ListRoles(t *testing.T) {
	logger := zap.NewNop()
	manager := NewRBACManager(logger)

	roles := manager.ListRoles()
	if len(roles) < 3 {
		t.Errorf("Expected at least 3 default roles, got %d", len(roles))
	}
}
