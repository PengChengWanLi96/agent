package service

import (
	"context"
	"agent/internal/client/docker"
	"agent/internal/client/metrics"
	"agent/internal/model"
)

type DockerService struct {
	client *docker.Client
}

func NewDockerService(client *docker.Client) *DockerService {
	return &DockerService{client: client}
}

func (s *DockerService) ListContainers(ctx context.Context, all bool) ([]model.Container, error) {
	return s.client.ListContainers(ctx, all)
}

func (s *DockerService) InspectContainer(ctx context.Context, id string) (*model.ContainerDetail, error) {
	return s.client.InspectContainer(ctx, id)
}

func (s *DockerService) StartContainer(ctx context.Context, id string) error {
	return s.client.StartContainer(ctx, id)
}

func (s *DockerService) StopContainer(ctx context.Context, id string, timeout int) error {
	return s.client.StopContainer(ctx, id, timeout)
}

func (s *DockerService) RemoveContainer(ctx context.Context, id string, force bool) error {
	return s.client.RemoveContainer(ctx, id, force)
}

func (s *DockerService) ContainerLogs(ctx context.Context, id string, tail string) (string, error) {
	return s.client.ContainerLogs(ctx, id, tail)
}

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
