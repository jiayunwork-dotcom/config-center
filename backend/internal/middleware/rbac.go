package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"config-center/internal/models"
	"config-center/internal/services"

	"github.com/gin-gonic/gin"
)

type PermissionConfig struct {
	RequireGlobalAdmin bool
	WriteOperation     bool
	NamespaceIDSource  string
}

type auditContextKey string

const (
	AuditDataKey auditContextKey = "audit_data"
)

type AuditData struct {
	ResourceType string
	Action       string
	OldValue     string
	NewValue     string
	ResourceID   *uint
	ResourceName string
}

func RBACMiddleware(cfg PermissionConfig) gin.HandlerFunc {
	roleService := services.NewRoleService()

	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if cfg.RequireGlobalAdmin {
			if !roleService.IsGlobalAdmin(userID) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin role required"})
				return
			}
			c.Next()
			return
		}

		namespaceID := extractNamespaceID(c, cfg.NamespaceIDSource)

		requiredRole := models.RoleViewer
		if cfg.WriteOperation {
			requiredRole = models.RoleEditor
		}

		if namespaceID > 0 {
			if !roleService.HasPermission(userID, namespaceID, requiredRole) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
		} else {
			if !roleService.IsGlobalAdmin(userID) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
		}

		c.Next()
	}
}

func extractNamespaceID(c *gin.Context, source string) uint {
	switch source {
	case "param_id":
		return 0
	case "param_namespace":
		idStr := c.Param("id")
		id, _ := strconv.ParseUint(idStr, 10, 32)
		return uint(id)
	case "query_namespace":
		idStr := c.Query("namespace_id")
		id, _ := strconv.ParseUint(idStr, 10, 32)
		return uint(id)
	case "body_namespace":
		body, ok := ParseJSONBody(c)
		if ok {
			if v, found := body["namespace_id"]; found {
				switch val := v.(type) {
				case float64:
					return uint(val)
				case string:
					id, _ := strconv.ParseUint(val, 10, 32)
					return uint(id)
				}
			}
		}
	}
	return 0
}

func FilterNamespacesMiddleware() gin.HandlerFunc {
	roleService := services.NewRoleService()
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			c.Next()
			return
		}
		nsIDs := roleService.GetAccessibleNamespaceIDs(userID)
		c.Set("accessible_namespaces", nsIDs)
		c.Next()
	}
}

func GetAccessibleNamespaces(c *gin.Context) []uint {
	val, exists := c.Get("accessible_namespaces")
	if !exists {
		return nil
	}
	return val.([]uint)
}

func IsWriteMethod(c *gin.Context) bool {
	method := strings.ToUpper(c.Request.Method)
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete
}

func SetAuditData(c *gin.Context, data *AuditData) {
	c.Set(string(AuditDataKey), data)
}

func GetAuditData(c *gin.Context) *AuditData {
	val, exists := c.Get(string(AuditDataKey))
	if !exists {
		return nil
	}
	return val.(*AuditData)
}
