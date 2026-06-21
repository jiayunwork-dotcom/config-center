package handlers

import (
	"net/http"
	"strconv"

	"config-center/internal/merger"
	"config-center/internal/models"
	"config-center/internal/services"
	"config-center/internal/validator"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	configService    *services.ConfigService
	namespaceService *services.NamespaceService
	groupService     *services.GroupService
}

func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{
		configService:    services.NewConfigService(),
		namespaceService: services.NewNamespaceService(),
		groupService:     services.NewGroupService(),
	}
}

func (h *ConfigHandler) CreateConfigItem(c *gin.Context) {
	var item models.ConfigItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	operator := c.GetHeader("X-Operator")
	if operator == "" {
		operator = "anonymous"
	}

	result, err := h.configService.CreateConfigItem(&item, operator)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	var req struct {
		Value       string `json:"value"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	operator := c.GetHeader("X-Operator")
	if operator == "" {
		operator = "anonymous"
	}

	result, err := h.configService.UpdateConfigItem(uint(id), req.Value, operator, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *ConfigHandler) DeleteConfigItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.configService.DeleteConfigItem(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *ConfigHandler) RollbackVersion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Version int `json:"version"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	operator := c.GetHeader("X-Operator")
	if operator == "" {
		operator = "anonymous"
	}

	result, err := h.configService.RollbackToVersion(uint(id), req.Version, operator)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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
	namespaceID, _ := strconv.ParseUint(c.Query("namespace_id"), 10, 32)
	groupID, _ := strconv.ParseUint(c.Query("group_id"), 10, 32)
	environment := c.DefaultQuery("environment", "dev")

	publicItems, _ := h.configService.GetConfigItems(1, 1, environment)
	namespaceItems, _ := h.configService.GetConfigItems(uint(namespaceID), uint(groupID), environment)

	publicMap := make(map[string]string)
	for _, item := range publicItems {
		if item.Level == "public" {
			publicMap[item.Key] = item.Value
		}
	}

	namespaceMap := make(map[string]string)
	groupMap := make(map[string]string)
	for _, item := range namespaceItems {
		if item.Level == "namespace" {
			namespaceMap[item.Key] = item.Value
		} else {
			groupMap[item.Key] = item.Value
		}
	}

	format := c.DefaultQuery("format", "json")
	merged := merger.MergeConfigs(publicMap, namespaceMap, groupMap, format)

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
