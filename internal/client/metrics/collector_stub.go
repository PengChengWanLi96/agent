//go:build !linux

package metrics

import (
	"errors"
	"agent/internal/model"
)

type NodeExporterCollector struct{}

func NewCollector() (*NodeExporterCollector, error) {
	return &NodeExporterCollector{}, nil
}

func (c *NodeExporterCollector) Collect() (*model.SystemMetrics, error) {
	return nil, errors.New("node_exporter collector is only available on Linux")
}

func (c *NodeExporterCollector) CollectRaw() (string, error) {
	return "", errors.New("node_exporter collector is only available on Linux")
}
