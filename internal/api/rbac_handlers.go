package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"novabackup/pkg/rbac"
)

var rbacManager *rbac.RBACManager

func init() {
	rbacManager = rbac.NewRBACManager(zap.NewNop())
}

func (s *Server) listUsers(c *gin.Context) {
	users := rbacManager.ListUsers()
	c.JSON(http.StatusOK, users)
}

func (s *Server) createUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &rbac.User{
		ID:       fmt.Sprintf("user_%s", uuid.New().String()[:8]),
		Username: req.Username,
		Email:    req.Email,
		Active:   true,
	}
	if err := rbacManager.CreateUser(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (s *Server) getUser(c *gin.Context) {
	id := c.Param("id")
	user, err := rbacManager.GetUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (s *Server) updateUser(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Email  string `json:"email"`
		Active bool   `json:"active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := rbacManager.GetUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	user.Email = req.Email
	user.Active = req.Active
	rbacManager.UpdateUser(user)
	c.JSON(http.StatusOK, user)
}

func (s *Server) deleteUser(c *gin.Context) {
	id := c.Param("id")
	rbacManager.DeleteUser(id)
	c.JSON(http.StatusOK, gin.H{"message": "User deleted", "id": id})
}

func (s *Server) assignRoleToUser(c *gin.Context) {
	userID := c.Param("id")
	var req struct {
		RoleID string `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := rbacManager.AssignRoleToUser(userID, req.RoleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Role assigned"})
}

func (s *Server) removeRoleFromUser(c *gin.Context) {
	userID := c.Param("id")
	roleID := c.Param("roleId")
	rbacManager.RemoveRoleFromUser(userID, roleID)
	c.JSON(http.StatusOK, gin.H{"message": "Role removed"})
}

func (s *Server) listRoles(c *gin.Context) {
	roles := rbacManager.ListRoles()
	c.JSON(http.StatusOK, roles)
}

func (s *Server) createRole(c *gin.Context) {
	var req struct {
		ID          string   `json:"id" binding:"required"`
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	perms := make([]rbac.Permission, len(req.Permissions))
	for i, p := range req.Permissions {
		perms[i] = rbac.Permission(p)
	}

	role := &rbac.Role{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Permissions: perms,
	}
	if err := rbacManager.CreateRole(role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, role)
}

func (s *Server) getRole(c *gin.Context) {
	id := c.Param("id")
	role, err := rbacManager.GetRole(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, role)
}

func (s *Server) updateRole(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role, err := rbacManager.GetRole(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	if len(req.Permissions) > 0 {
		perms := make([]rbac.Permission, len(req.Permissions))
		for i, p := range req.Permissions {
			perms[i] = rbac.Permission(p)
		}
		role.Permissions = perms
	}

	rbacManager.UpdateRole(role)
	c.JSON(http.StatusOK, role)
}

func (s *Server) deleteRole(c *gin.Context) {
	id := c.Param("id")
	rbacManager.DeleteRole(id)
	c.JSON(http.StatusOK, gin.H{"message": "Role deleted", "id": id})
}

func (s *Server) listPermissions(c *gin.Context) {
	permissions := []string{
		"job:create", "job:read", "job:update", "job:delete", "job:execute",
		"backup:create", "backup:read", "backup:delete", "backup:restore",
		"vm:read", "vm:snapshot", "vm:restore",
		"storage:read", "storage:write", "storage:delete",
		"replication:read", "replication:write", "replication:delete",
		"monitoring:read", "monitoring:admin",
		"user:create", "user:read", "user:update", "user:delete",
		"system:config", "system:logs", "system:admin",
	}
	c.JSON(http.StatusOK, permissions)
}

func (s *Server) runReplicationJob(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Replication job started", "id": id})
}
