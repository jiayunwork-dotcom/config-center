package handlers

import (
	"net/http"
	"strconv"

	"config-center/internal/services"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	auditService *services.AuditService
}

func NewAuditHandler() *AuditHandler {
	return &AuditHandler{
		auditService: services.NewAuditService(),
	}
}

func (h *AuditHandler) ListLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var userID *uint
	uidStr := c.Query("user_id")
	if uidStr != "" {
		if uid, err := strconv.ParseUint(uidStr, 10, 32); err == nil {
			u := uint(uid)
			userID = &u
		}
	}

	action := c.Query("action")

	logs, total, err := h.auditService.ListLogs(page, pageSize, userID, action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": logs,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}
