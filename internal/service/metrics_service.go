package service

import (
	"fmt"
	"math"
	"strings"
	"time"

	"agent/internal/client/metrics"
	"agent/internal/model"
)

type MetricsService struct {
	collector *metrics.NodeExporterCollector
	startedAt time.Time
}

func NewMetricsService(collector *metrics.NodeExporterCollector, startedAt time.Time) *MetricsService {
	return &MetricsService{collector: collector, startedAt: startedAt}
}

func (s *MetricsService) Collect() (*model.MetricsCollectResponse, error) {
	m, err := s.collector.Collect()
	if err != nil {
		return nil, err
	}

	uptime := time.Since(s.startedAt)
	resp := &model.MetricsCollectResponse{
		CPU:             m.CPU,
		Memory:          m.Memory,
		Disk:            m.Disk,
		Network:         m.Network,
		Load:            m.Load,
		Timestamp:       m.Timestamp,
		UptimeSeconds:   math.Round(uptime.Seconds()*1000) / 1000,
		UptimeFormatted: formatDuration(uptime),
	}
	return resp, nil
}

func (s *MetricsService) CollectRaw() (string, error) {
	return s.collector.CollectRaw()
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	parts := []string{}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d天", days))
	}
	if hours > 0 || len(parts) > 0 {
		parts = append(parts, fmt.Sprintf("%d小时", hours))
	}
	if minutes > 0 || len(parts) > 0 {
		parts = append(parts, fmt.Sprintf("%d分", minutes))
	}
	parts = append(parts, fmt.Sprintf("%d秒", seconds))
	return strings.Join(parts, " ")
}
