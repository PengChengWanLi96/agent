package service

import (
	"agent/internal/client/metrics"
	"agent/internal/model"
)

type MetricsService struct {
	collector *metrics.NodeExporterCollector
}

func NewMetricsService(collector *metrics.NodeExporterCollector) *MetricsService {
	return &MetricsService{collector: collector}
}

func (s *MetricsService) Collect() (*model.SystemMetrics, error) {
	return s.collector.Collect()
}

func (s *MetricsService) CollectRaw() (string, error) {
	return s.collector.CollectRaw()
}
