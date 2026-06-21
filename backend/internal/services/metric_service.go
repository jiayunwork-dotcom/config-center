package services

import (
	"time"

	"config-center/internal/database"
	"config-center/internal/models"
)

type MetricService struct{}

func NewMetricService() *MetricService {
	return &MetricService{}
}

func (s *MetricService) RecordMetric(tenantID, namespaceID uint, metricType string, value float64) error {
	metric := &models.Metric{
		TenantID:    tenantID,
		NamespaceID: namespaceID,
		MetricType:  metricType,
		Value:       value,
		Timestamp:   time.Now(),
	}
	return database.DB.Create(metric).Error
}

func (s *MetricService) GetMetrics(namespaceID uint, metricType string, duration string) ([]models.Metric, error) {
	var metrics []models.Metric

	var startTime time.Time
	switch duration {
	case "1h":
		startTime = time.Now().Add(-1 * time.Hour)
	case "24h":
		startTime = time.Now().Add(-24 * time.Hour)
	default:
		startTime = time.Now().Add(-1 * time.Hour)
	}

	query := database.DB.Where("namespace_id = ? AND metric_type = ? AND timestamp >= ?",
		namespaceID, metricType, startTime).
		Order("timestamp ASC")

	err := query.Find(&metrics).Error
	return metrics, err
}

func (s *MetricService) GetLatestMetrics(namespaceID uint) (map[string]float64, error) {
	types := []string{"pull_qps", "push_success_rate", "avg_latency"}
	result := make(map[string]float64)

	for _, mt := range types {
		var metric models.Metric
		database.DB.Where("namespace_id = ? AND metric_type = ?", namespaceID, mt).
			Order("timestamp DESC").
			First(&metric)
		if metric.ID > 0 {
			result[mt] = metric.Value
		} else {
			result[mt] = 0
		}
	}

	return result, nil
}
