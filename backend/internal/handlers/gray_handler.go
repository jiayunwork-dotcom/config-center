package handlers

import (
	"net/http"
	"strconv"

	"config-center/internal/middleware"
	"config-center/internal/models"
	"config-center/internal/services"

	"github.com/gin-gonic/gin"
)

type GrayHandler struct {
	grayService  *services.GrayReleaseService
	auditService *services.AuditService
}

func NewGrayHandler() *GrayHandler {
	return &GrayHandler{
		grayService:  services.NewGrayReleaseService(),
		auditService: services.NewAuditService(),
	}
}

func (h *GrayHandler) CreateGrayRelease(c *gin.Context) {
	var release models.GrayRelease
	if err := c.ShouldBindJSON(&release); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if release.TenantID == 0 {
		release.TenantID = 1
	}

	result, err := h.grayService.CreateGrayRelease(&release)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uid := middleware.GetUserID(c)
	uname := middleware.GetUsername(c)
	rid := result.ID
	name := "gray-" + strconv.Itoa(int(result.ID))
	go h.auditService.CreateLog(&uid, uname, models.ActionCreate, models.ResourceGray, &rid, name, "", "", getClientIP(c))

	c.JSON(http.StatusOK, result)
}

func (h *GrayHandler) StartGrayRelease(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	result, err := h.grayService.StartGrayRelease(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uid := middleware.GetUserID(c)
	uname := middleware.GetUsername(c)
	rid := result.ID
	name := "gray-" + strconv.Itoa(int(result.ID))
	go h.auditService.CreateLog(&uid, uname, models.ActionStart, models.ResourceGray, &rid, name, "", "", getClientIP(c))

	c.JSON(http.StatusOK, result)
}

func (h *GrayHandler) FullPush(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	result, err := h.grayService.FullPush(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uid := middleware.GetUserID(c)
	uname := middleware.GetUsername(c)
	rid := result.ID
	name := "gray-" + strconv.Itoa(int(result.ID))
	go h.auditService.CreateLog(&uid, uname, models.ActionFullPush, models.ResourceGray, &rid, name, "", "", getClientIP(c))

	c.JSON(http.StatusOK, result)
}

func (h *GrayHandler) RollbackGrayRelease(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	result, err := h.grayService.RollbackGrayRelease(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uid := middleware.GetUserID(c)
	uname := middleware.GetUsername(c)
	rid := result.ID
	name := "gray-" + strconv.Itoa(int(result.ID))
	go h.auditService.CreateLog(&uid, uname, models.ActionRollback, models.ResourceGray, &rid, name, "", "", getClientIP(c))

	c.JSON(http.StatusOK, result)
}

func (h *GrayHandler) ListGrayReleases(c *gin.Context) {
	configItemID, _ := strconv.ParseUint(c.Query("config_item_id"), 10, 32)

	releases, err := h.grayService.GetGrayReleases(uint(configItemID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, releases)
}

func (h *GrayHandler) GetGrayRelease(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	release, err := h.grayService.GetGrayRelease(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, release)
}
