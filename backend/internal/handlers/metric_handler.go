package handlers

import (
	"net/http"
	"strconv"

	"config-center/internal/services"

	"github.com/gin-gonic/gin"
)

type MetricHandler struct {
	metricService *services.MetricService
}

func NewMetricHandler() *MetricHandler {
	return &MetricHandler{
		metricService: services.NewMetricService(),
	}
}

func (h *MetricHandler) GetMetrics(c *gin.Context) {
	namespaceID, _ := strconv.ParseUint(c.Query("namespace_id"), 10, 32)
	metricType := c.Query("metric_type")
	duration := c.DefaultQuery("duration", "1h")

	metrics, err := h.metricService.GetMetrics(uint(namespaceID), metricType, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *MetricHandler) GetLatestMetrics(c *gin.Context) {
	namespaceID, _ := strconv.ParseUint(c.Query("namespace_id"), 10, 32)

	metrics, err := h.metricService.GetLatestMetrics(uint(namespaceID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}
