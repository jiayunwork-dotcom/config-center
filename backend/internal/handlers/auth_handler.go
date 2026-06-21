package handlers

import (
	"net/http"
	"strconv"

	"config-center/internal/auth"
	"config-center/internal/middleware"
	"config-center/internal/models"
	"config-center/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
	roleService *services.RoleService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: services.NewAuthService(),
		roleService: services.NewRoleService(),
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
		},
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	username := middleware.GetUsername(c)

	roles, _ := h.roleService.GetUserRoles(userID)
	accessibleNS := h.roleService.GetAccessibleNamespaceIDs(userID)
	isAdmin := h.roleService.IsGlobalAdmin(userID)

	c.JSON(http.StatusOK, gin.H{
		"id":                    userID,
		"username":              username,
		"roles":                 roles,
		"accessible_namespaces": accessibleNS,
		"is_global_admin":       isAdmin,
	})
}

func (h *AuthHandler) CreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.CreateUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) ListUsers(c *gin.Context) {
	users, err := h.authService.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]map[string]interface{}, 0, len(users))
	for _, u := range users {
		roles, _ := h.roleService.GetRoleBindings(u.ID)
		result = append(result, map[string]interface{}{
			"id":         u.ID,
			"username":   u.Username,
			"created_at": u.CreatedAt,
			"roles":      roles,
		})
	}

	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.authService.DeleteUser(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *AuthHandler) GrantRole(c *gin.Context) {
	var req struct {
		UserID      uint   `json:"user_id" binding:"required"`
		NamespaceID *uint  `json:"namespace_id"`
		Role        string `json:"role" binding:"required,oneof=admin editor viewer"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.roleService.GrantRole(req.UserID, req.NamespaceID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role granted"})
}

func (h *AuthHandler) RevokeRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.roleService.RevokeRole(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role revoked"})
}

func (h *AuthHandler) GetRoles(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	roles, err := h.roleService.GetRoleBindings(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, roles)
}

func (h *AuthHandler) ListNamespacePermissions(c *gin.Context) {
	namespaceID, err := strconv.ParseUint(c.Query("namespace_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid namespace_id"})
		return
	}

	var bindings []models.UserRole
	services.NewRoleService()
	h.roleService = services.NewRoleService()

	_ = namespaceID
	_ = bindings

	c.JSON(http.StatusOK, gin.H{"message": "use /users endpoint"})
}
