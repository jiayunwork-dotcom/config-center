package handlers

import (
	"net/http"
	"strconv"

	"config-center/internal/database"
	"config-center/internal/middleware"
	"config-center/internal/models"
	"config-center/internal/services"

	"github.com/gin-gonic/gin"
)

type NamespaceHandler struct {
	namespaceService *services.NamespaceService
	groupService     *services.GroupService
	roleService      *services.RoleService
	auditService     *services.AuditService
}

func NewNamespaceHandler() *NamespaceHandler {
	return &NamespaceHandler{
		namespaceService: services.NewNamespaceService(),
		groupService:     services.NewGroupService(),
		roleService:      services.NewRoleService(),
		auditService:     services.NewAuditService(),
	}
}

func getClientIP(c *gin.Context) string {
	ip := c.GetHeader("X-Forwarded-For")
	if ip != "" {
		return ip
	}
	ip = c.GetHeader("X-Real-IP")
	if ip != "" {
		return ip
	}
	return c.ClientIP()
}

func (h *NamespaceHandler) CreateNamespace(c *gin.Context) {
	var ns models.Namespace
	if err := c.ShouldBindJSON(&ns); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if ns.TenantID == 0 {
		ns.TenantID = 1
	}

	result, err := h.namespaceService.CreateNamespace(&ns)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uid := middleware.GetUserID(c)
	uname := middleware.GetUsername(c)
	rid := result.ID
	go h.auditService.CreateLog(&uid, uname, models.ActionCreate, models.ResourceNamespace, &rid, result.Name, "", result.Name, getClientIP(c))

	c.JSON(http.StatusOK, result)
}

func (h *NamespaceHandler) ListNamespaces(c *gin.Context) {
	tenantID, _ := strconv.ParseUint(c.DefaultQuery("tenant_id", "1"), 10, 32)
	accessibleIDs := middleware.GetAccessibleNamespaces(c)

	namespaces, err := h.namespaceService.GetNamespaces(uint(tenantID), accessibleIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, namespaces)
}

func (h *NamespaceHandler) GetNamespace(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ns, err := h.namespaceService.GetNamespace(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ns)
}

func (h *NamespaceHandler) UpdateNamespace(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	oldNs, _ := h.namespaceService.GetNamespace(uint(id))
	oldName := ""
	if oldNs != nil {
		oldName = oldNs.Name
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.namespaceService.UpdateNamespace(uint(id), req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uid := middleware.GetUserID(c)
	uname := middleware.GetUsername(c)
	rid := result.ID
	go h.auditService.CreateLog(&uid, uname, models.ActionUpdate, models.ResourceNamespace, &rid, result.Name, oldName, result.Name, getClientIP(c))

	c.JSON(http.StatusOK, result)
}

func (h *NamespaceHandler) DeleteNamespace(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	oldNs, _ := h.namespaceService.GetNamespace(uint(id))
	oldName := ""
	if oldNs != nil {
		oldName = oldNs.Name
	}
	rid := uint(id)

	if err := h.namespaceService.DeleteNamespace(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uid := middleware.GetUserID(c)
	uname := middleware.GetUsername(c)
	go h.auditService.CreateLog(&uid, uname, models.ActionDelete, models.ResourceNamespace, &rid, oldName, oldName, "", getClientIP(c))

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *NamespaceHandler) CreateGroup(c *gin.Context) {
	var group models.Group
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if group.TenantID == 0 {
		group.TenantID = 1
	}

	result, err := h.groupService.CreateGroup(&group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uid := middleware.GetUserID(c)
	uname := middleware.GetUsername(c)
	rid := result.ID
	go h.auditService.CreateLog(&uid, uname, models.ActionCreate, models.ResourceGroup, &rid, result.Name, "", result.Name, getClientIP(c))

	c.JSON(http.StatusOK, result)
}

func (h *NamespaceHandler) ListGroups(c *gin.Context) {
	namespaceID, err := strconv.ParseUint(c.Query("namespace_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid namespace_id"})
		return
	}

	groups, err := h.groupService.GetGroups(uint(namespaceID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, groups)
}

func (h *NamespaceHandler) GetGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	group, err := h.groupService.GetGroup(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

func (h *NamespaceHandler) UpdateGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	oldGroup, _ := h.groupService.GetGroup(uint(id))
	oldName := ""
	if oldGroup != nil {
		oldName = oldGroup.Name
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.groupService.UpdateGroup(uint(id), req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uid := middleware.GetUserID(c)
	uname := middleware.GetUsername(c)
	rid := result.ID
	go h.auditService.CreateLog(&uid, uname, models.ActionUpdate, models.ResourceGroup, &rid, result.Name, oldName, result.Name, getClientIP(c))

	c.JSON(http.StatusOK, result)
}

func (h *NamespaceHandler) DeleteGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	oldGroup, _ := h.groupService.GetGroup(uint(id))
	oldName := ""
	if oldGroup != nil {
		oldName = oldGroup.Name
	}
	rid := uint(id)

	if err := h.groupService.DeleteGroup(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uid := middleware.GetUserID(c)
	uname := middleware.GetUsername(c)
	go h.auditService.CreateLog(&uid, uname, models.ActionDelete, models.ResourceGroup, &rid, oldName, oldName, "", getClientIP(c))

	_ = database.DB
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
