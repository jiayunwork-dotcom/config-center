package handlers

import (
	"net/http"
	"strconv"

	"config-center/internal/middleware"
	"config-center/internal/models"
	"config-center/internal/services"
	"config-center/internal/validator"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	configService    *services.ConfigService
	namespaceService *services.NamespaceService
	groupService     *services.GroupService
	auditService     *services.AuditService
	approvalService  *services.ApprovalService
	roleService      *services.RoleService
}

func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{
		configService:    services.NewConfigService(),
		namespaceService: services.NewNamespaceService(),
		groupService:     services.NewGroupService(),
		auditService:     services.NewAuditService(),
		approvalService:  services.NewApprovalService(),
		roleService:      services.NewRoleService(),
	}
}

func (h *ConfigHandler) CreateConfigItem(c *gin.Context) {
	var item models.ConfigItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	operator := middleware.GetUsername(c)
	if operator == "" {
		operator = "anonymous"
	}

	result, err := h.configService.CreateConfigItem(&item, operator)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := middleware.GetUserID(c)
	uname := operator
	rid := result.ID
	go h.auditService.CreateLog(&uid, uname, models.ActionCreate, models.ResourceConfig, &rid, result.Key, "", item.Value, getClientIP(c))

	c.JSON(http.StatusOK, result)
}

func (h *ConfigHandler) GetConfigItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	item, err := h.configService.GetConfigItem(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *ConfigHandler) ListConfigItems(c *gin.Context) {
	namespaceID, _ := strconv.ParseUint(c.Query("namespace_id"), 10, 32)
	groupID, _ := strconv.ParseUint(c.Query("group_id"), 10, 32)
	environment := c.DefaultQuery("environment", "dev")

	items, err := h.configService.GetConfigItems(uint(namespaceID), uint(groupID), environment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *ConfigHandler) UpdateConfigItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	oldItem, _ := h.configService.GetConfigItem(uint(id))
	oldValue := ""
	configKey := ""
	if oldItem != nil {
		oldValue = oldItem.Value
		configKey = oldItem.Key
	}

	var req struct {
		Value       string `json:"value"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	operator := middleware.GetUsername(c)
	if operator == "" {
		operator = "anonymous"
	}

	userID := middleware.GetUserID(c)
	isGlobalAdmin := h.roleService.IsGlobalAdmin(userID)

	if oldItem != nil && oldItem.Environment == "prod" && !isGlobalAdmin {
		hasPending, _ := h.approvalService.HasPendingApproval(uint(id))
		if hasPending {
			c.JSON(http.StatusConflict, gin.H{"error": "该配置已有待审批的变更申请"})
			return
		}

		approval, err := h.approvalService.CreateApproval(
			userID, operator, uint(id), configKey, req.Value, oldValue, oldItem.Environment, req.Description,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "已提交审批，等待管理员审批",
			"approval_id":   approval.ID,
			"requires_approval": true,
		})
		return
	}

	result, err := h.configService.UpdateConfigItem(uint(id), req.Value, operator, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := userID
	uname := operator
	rid := result.ID
	go h.auditService.CreateLog(&uid, uname, models.ActionUpdate, models.ResourceConfig, &rid, configKey, oldValue, req.Value, getClientIP(c))

	c.JSON(http.StatusOK, result)
}

func (h *ConfigHandler) DeleteConfigItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	oldItem, _ := h.configService.GetConfigItem(uint(id))
	oldValue := ""
	configKey := ""
	if oldItem != nil {
		oldValue = oldItem.Value
		configKey = oldItem.Key
	}
	rid := uint(id)

	if err := h.configService.DeleteConfigItem(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uid := middleware.GetUserID(c)
	uname := middleware.GetUsername(c)
	go h.auditService.CreateLog(&uid, uname, models.ActionDelete, models.ResourceConfig, &rid, configKey, oldValue, "", getClientIP(c))

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *ConfigHandler) BatchDeleteConfigItems(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids is required"})
		return
	}

	uid := middleware.GetUserID(c)
	uname := middleware.GetUsername(c)
	ip := getClientIP(c)

	for _, id := range req.IDs {
		oldItem, _ := h.configService.GetConfigItem(id)
		oldValue := ""
		configKey := ""
		if oldItem != nil {
			oldValue = oldItem.Value
			configKey = oldItem.Key
		}
		rid := id
		go h.auditService.CreateLog(&uid, uname, models.ActionDelete, models.ResourceConfig, &rid, configKey, oldValue, "", ip)
	}

	if err := h.configService.BatchDeleteConfigItems(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted", "count": len(req.IDs)})
}

func (h *ConfigHandler) BatchCopyConfigItems(c *gin.Context) {
	var req struct {
		SourceIDs       []uint `json:"source_ids"`
		TargetEnvironment string `json:"target_environment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.SourceIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source_ids is required"})
		return
	}
	if req.TargetEnvironment == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_environment is required"})
		return
	}

	operator := middleware.GetUsername(c)
	if operator == "" {
		operator = "anonymous"
	}

	results, err := h.configService.BatchCopyConfigItems(req.SourceIDs, req.TargetEnvironment, operator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uid := middleware.GetUserID(c)
	uname := operator
	ip := getClientIP(c)

	for _, r := range results {
		if r.Status == "success" {
			rid := r.ID
			go h.auditService.CreateLog(&uid, uname, models.ActionCreate, models.ResourceConfig, &rid, r.Key, "", "", ip)
		}
	}

	successCount := 0
	skippedCount := 0
	failedCount := 0
	for _, r := range results {
		switch r.Status {
		case "success":
			successCount++
		case "skipped":
			skippedCount++
		case "failed":
			failedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"results":       results,
		"success_count": successCount,
		"skipped_count": skippedCount,
		"failed_count":  failedCount,
	})
}

func (h *ConfigHandler) RollbackVersion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	oldItem, _ := h.configService.GetConfigItem(uint(id))
	oldValue := ""
	configKey := ""
	if oldItem != nil {
		oldValue = oldItem.Value
		configKey = oldItem.Key
	}

	var req struct {
		Version int `json:"version"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	operator := middleware.GetUsername(c)
	if operator == "" {
		operator = "anonymous"
	}

	result, err := h.configService.RollbackToVersion(uint(id), req.Version, operator)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := middleware.GetUserID(c)
	uname := operator
	rid := result.ID
	newValue := result.Value
	go h.auditService.CreateLog(&uid, uname, models.ActionRollback, models.ResourceConfig, &rid, configKey, oldValue, newValue, getClientIP(c))

	c.JSON(http.StatusOK, result)
}

func (h *ConfigHandler) GetVersionHistory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	versions, total, err := h.configService.GetVersionHistory(uint(id), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": versions,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func (h *ConfigHandler) CompareVersions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	v1, _ := strconv.Atoi(c.Query("version1"))
	v2, _ := strconv.Atoi(c.Query("version2"))

	diff, err := h.configService.CompareVersions(uint(id), v1, v2)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"diff": diff})
}

func (h *ConfigHandler) GetMergedConfig(c *gin.Context) {
	tenantID, _ := strconv.ParseUint(c.DefaultQuery("tenant_id", "1"), 10, 32)
	namespaceID, _ := strconv.ParseUint(c.Query("namespace_id"), 10, 32)
	groupID, _ := strconv.ParseUint(c.Query("group_id"), 10, 32)
	environment := c.DefaultQuery("environment", "dev")

	if namespaceID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace_id is required"})
		return
	}

	merged, err := h.configService.GetMergedConfig(uint(tenantID), uint(namespaceID), uint(groupID), environment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, merged)
}

func (h *ConfigHandler) ValidateConfig(c *gin.Context) {
	var req struct {
		Value  string `json:"value"`
		Format string `json:"format"`
		Schema string `json:"schema"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "valid": false})
		return
	}

	type ValidationResult struct {
		Valid   bool   `json:"valid"`
		Message string `json:"message,omitempty"`
	}

	if err := validator.ValidateFormat(req.Value, req.Format); err != nil {
		c.JSON(http.StatusOK, ValidationResult{Valid: false, Message: err.Error()})
		return
	}

	if req.Schema != "" && req.Format == "json" {
		if err := validator.ValidateWithSchema(req.Value, req.Schema); err != nil {
			c.JSON(http.StatusOK, ValidationResult{Valid: false, Message: err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, ValidationResult{Valid: true})
}
