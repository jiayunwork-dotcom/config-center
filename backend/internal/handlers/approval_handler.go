package handlers

import (
	"net/http"
	"strconv"

	"config-center/internal/middleware"
	"config-center/internal/services"

	"github.com/gin-gonic/gin"
)

type ApprovalHandler struct {
	approvalService *services.ApprovalService
	auditService    *services.AuditService
}

func NewApprovalHandler() *ApprovalHandler {
	return &ApprovalHandler{
		approvalService: services.NewApprovalService(),
		auditService:    services.NewAuditService(),
	}
}

func (h *ApprovalHandler) ListApprovals(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	status := c.Query("status")

	approvals, total, err := h.approvalService.ListApprovals(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": approvals,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func (h *ApprovalHandler) Approve(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	reviewerID := middleware.GetUserID(c)
	reviewer := middleware.GetUsername(c)

	result, err := h.approvalService.Approve(uint(id), reviewerID, reviewer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := reviewerID
	uname := reviewer
	rid := result.ID
	go h.auditService.CreateLog(&uid, uname, "update", "config", &rid, result.Key, "", result.Value, getClientIP(c))

	c.JSON(http.StatusOK, gin.H{
		"message": "approved and config updated",
		"config":  result,
	})
}

func (h *ApprovalHandler) Reject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		ReviewNote string `json:"review_note"`
	}
	c.ShouldBindJSON(&req)

	reviewerID := middleware.GetUserID(c)
	reviewer := middleware.GetUsername(c)

	if err := h.approvalService.Reject(uint(id), reviewerID, reviewer, req.ReviewNote); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rejected"})
}
