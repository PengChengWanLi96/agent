package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	nat "github.com/docker/go-connections/nat"

	"agent/internal/client/docker"
	"agent/internal/model"
)

type DockerService struct {
	client *docker.Client
}

func NewDockerService(client *docker.Client) *DockerService {
	return &DockerService{client: client}
}

func (s *DockerService) ListContainers(ctx context.Context, all bool, filters map[string][]string) ([]model.Container, error) {
	containers, err := s.client.ListContainers(ctx, all, filters)
	if err != nil {
		return nil, err
	}

	result := make([]model.Container, len(containers))
	for i, c := range containers {
		result[i] = convertContainer(c)
	}
	return result, nil
}

func (s *DockerService) InspectContainer(ctx context.Context, id string) (*model.ContainerDetail, error) {
	inspect, err := s.client.InspectContainer(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertContainerInspect(inspect), nil
}

func (s *DockerService) CreateContainer(ctx context.Context, req *model.ContainerCreateRequest) (map[string]string, error) {
	config := &container.Config{
		Image:      req.Image,
		Cmd:        req.Cmd,
		Entrypoint: req.Entrypoint,
		Env:        req.Env,
		Labels:     req.Labels,
		Hostname:   req.Hostname,
		User:       req.User,
		WorkingDir: req.WorkingDir,
		Tty:        req.Tty,
		OpenStdin:  req.StdinOpen,
	}

	if len(req.ExposedPorts) > 0 {
		config.ExposedPorts = make(nat.PortSet)
		for port := range req.ExposedPorts {
			config.ExposedPorts[nat.Port(port)] = struct{}{}
		}
	}

	if req.Volumes != nil {
		config.Volumes = req.Volumes
	}

	hostConfig := &container.HostConfig{
		AutoRemove:  req.AutoRemove,
		Privileged:  req.Privileged,
		NetworkMode: container.NetworkMode(req.NetworkMode),
		DNS:         req.DNS,
		Resources: container.Resources{
			Memory:    req.Memory,
			CPUQuota:  req.CPUQuota,
			CPUPeriod: req.CPUPeriod,
			CPUShares: req.CPUShares,
		},
	}

	if len(req.PortBindings) > 0 {
		hostConfig.PortBindings = make(map[nat.Port][]nat.PortBinding)
		for port, bindings := range req.PortBindings {
			natPort := nat.Port(port)
			for _, b := range bindings {
				hostIP := b.HostIP
				if hostIP == "" {
					hostIP = "0.0.0.0"
				}
				hostConfig.PortBindings[natPort] = append(hostConfig.PortBindings[natPort], nat.PortBinding{
					HostIP:   hostIP,
					HostPort: b.HostPort,
				})
			}
		}
	}

	if len(req.Mounts) > 0 {
		hostConfig.Mounts = make([]mount.Mount, len(req.Mounts))
		for i, m := range req.Mounts {
			hostConfig.Mounts[i] = mount.Mount{
				Type:     mount.Type(m.Type),
				Source:   m.Source,
				Target:   m.Target,
				ReadOnly: m.ReadOnly,
			}
			if m.BindOptions != nil {
				hostConfig.Mounts[i].BindOptions = &mount.BindOptions{
					Propagation: mount.Propagation(m.BindOptions.Propagation),
				}
			}
		}
	}

	if req.RestartPolicy != "" {
		hostConfig.RestartPolicy = container.RestartPolicy{
			Name:              req.RestartPolicy,
			MaximumRetryCount: 0,
		}
		if req.RestartRetries != nil {
			hostConfig.RestartPolicy.MaximumRetryCount = *req.RestartRetries
		}
	}

	var networkingConfig *network.NetworkingConfig
	if len(req.Networks) > 0 {
		networkingConfig = &network.NetworkingConfig{
			EndpointsConfig: make(map[string]*network.EndpointSettings),
		}
		for _, net := range req.Networks {
			networkName := net.NetworkName
			if networkName == "" {
				networkName = net.NetworkID
			}
			networkingConfig.EndpointsConfig[networkName] = &network.EndpointSettings{
				Aliases: net.Aliases,
			}
			if net.IPv4Address != "" {
				networkingConfig.EndpointsConfig[networkName].IPAMConfig = &network.EndpointIPAMConfig{
					IPv4Address: net.IPv4Address,
				}
			}
			if net.IPv6Address != "" {
				if networkingConfig.EndpointsConfig[networkName].IPAMConfig == nil {
					networkingConfig.EndpointsConfig[networkName].IPAMConfig = &network.EndpointIPAMConfig{}
				}
				networkingConfig.EndpointsConfig[networkName].IPAMConfig.IPv6Address = net.IPv6Address
			}
		}
	}

	resp, err := s.client.CreateContainer(ctx, config, hostConfig, networkingConfig, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	return map[string]string{
		"id":       resp.ID,
		"warnings": strings.Join(resp.Warnings, ", "),
	}, nil
}

func (s *DockerService) StartContainer(ctx context.Context, id string) error {
	return s.client.StartContainer(ctx, id)
}

func (s *DockerService) StopContainer(ctx context.Context, id string, timeout int) error {
	var t *int
	if timeout >= 0 {
		t = &timeout
	}
	return s.client.StopContainer(ctx, id, t)
}

func (s *DockerService) RestartContainer(ctx context.Context, id string, timeout int) error {
	var t *int
	if timeout >= 0 {
		t = &timeout
	}
	return s.client.RestartContainer(ctx, id, t)
}

func (s *DockerService) KillContainer(ctx context.Context, id, signal string) error {
	if signal == "" {
		signal = "SIGKILL"
	}
	return s.client.KillContainer(ctx, id, signal)
}

func (s *DockerService) RemoveContainer(ctx context.Context, id string, force bool) error {
	return s.client.RemoveContainer(ctx, id, force)
}

func (s *DockerService) PauseContainer(ctx context.Context, id string) error {
	return s.client.PauseContainer(ctx, id)
}

func (s *DockerService) UnpauseContainer(ctx context.Context, id string) error {
	return s.client.UnpauseContainer(ctx, id)
}

func (s *DockerService) RenameContainer(ctx context.Context, id string, newName string) error {
	return s.client.RenameContainer(ctx, id, newName)
}

func (s *DockerService) ContainerLogs(ctx context.Context, id string, tail string, timestamps bool, follow bool, since string, until string) (string, error) {
	logs, err := s.client.GetContainerLogs(ctx, id, tail, timestamps, follow, since, until)
	if err != nil {
		return "", err
	}
	defer logs.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, logs); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *DockerService) PruneContainers(ctx context.Context) (map[string]interface{}, error) {
	report, err := s.client.PruneContainers(ctx, nil)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"containers_deleted": report.ContainersDeleted,
		"space_reclaimed":    report.SpaceReclaimed,
	}, nil
}

func (s *DockerService) ExecContainer(ctx context.Context, id string, req *model.ContainerExecRequest) (*model.ContainerExecResponse, error) {
	if !req.AttachStdin && !req.AttachStdout && !req.AttachStderr {
		req.AttachStdout = true
		req.AttachStderr = true
	}

	execResp, err := s.client.CreateExec(ctx, id, req.Cmd, req.WorkingDir, req.Env, req.User, req.Privileged, req.Tty, req.AttachStdin, req.AttachStdout, req.AttachStderr)
	if err != nil {
		return nil, err
	}

	if !req.AttachStdout && !req.AttachStderr {
		return &model.ContainerExecResponse{ID: execResp.ID}, nil
	}

	attachResp, err := s.client.ExecAttach(ctx, execResp.ID, req.Tty)
	if err != nil {
		return nil, err
	}
	defer attachResp.Close()

	var output bytes.Buffer
	if req.Tty {
		io.Copy(&output, attachResp.Reader)
	} else {
		docker.StdCopy(&output, &output, attachResp.Reader)
	}

	inspect, err := s.client.ExecInspect(ctx, execResp.ID)
	if err != nil {
		return &model.ContainerExecResponse{
			ID:     execResp.ID,
			Output: output.String(),
		}, nil
	}

	return &model.ContainerExecResponse{
		ID:       execResp.ID,
		Output:   output.String(),
		ExitCode: inspect.ExitCode,
	}, nil
}

// ==================== Image Service ====================

func (s *DockerService) ListImages(ctx context.Context, all bool, filters map[string][]string) ([]model.Image, error) {
	images, err := s.client.ListImages(ctx, all, filters)
	if err != nil {
		return nil, err
	}

	result := make([]model.Image, len(images))
	for i, img := range images {
		result[i] = model.Image{
			ID:          img.ID,
			RepoTags:    img.RepoTags,
			RepoDigests: img.RepoDigests,
			Created:     img.Created,
			Size:        img.Size,
			VirtualSize: img.VirtualSize,
			Labels:      img.Labels,
		}
	}
	return result, nil
}

func (s *DockerService) InspectImage(ctx context.Context, id string) (*model.ImageDetail, error) {
	inspect, _, err := s.client.InspectImage(ctx, id)
	if err != nil {
		return nil, err
	}

	labels := inspect.Config.Labels
	return &model.ImageDetail{
		ID:           inspect.ID,
		RepoTags:     inspect.RepoTags,
		RepoDigests:  inspect.RepoDigests,
		Parent:       inspect.Parent,
		Comment:      inspect.Comment,
		Created:      inspect.Created,
		Container:    inspect.Container,
		Size:         inspect.Size,
		VirtualSize:  inspect.VirtualSize,
		Labels:       labels,
		Architecture: inspect.Architecture,
		Os:           inspect.Os,
	}, nil
}

func (s *DockerService) PullImage(ctx context.Context, req *model.ImagePullRequest) (io.ReadCloser, error) {
	return s.client.PullImage(ctx, req.Image, req.RegistryAuth, req.Insecure, req.Platform)
}

func (s *DockerService) RemoveImage(ctx context.Context, id string, force bool) ([]types.ImageDeleteResponseItem, error) {
	return s.client.RemoveImage(ctx, id, force)
}

func (s *DockerService) PruneImages(ctx context.Context) (map[string]interface{}, error) {
	report, err := s.client.PruneImages(ctx, nil)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"images_deleted":  report.ImagesDeleted,
		"space_reclaimed": report.SpaceReclaimed,
	}, nil
}

// ==================== Network Service ====================

func (s *DockerService) ListNetworks(ctx context.Context, filters map[string][]string) ([]model.Network, error) {
	networks, err := s.client.ListNetworks(ctx, filters)
	if err != nil {
		return nil, err
	}

	result := make([]model.Network, len(networks))
	for i, n := range networks {
		result[i] = model.Network{
			Name:       n.Name,
			ID:         n.ID,
			Created:    n.Created.String(),
			Scope:      n.Scope,
			Driver:     n.Driver,
			EnableIPv6: n.EnableIPv6,
			Internal:   n.Internal,
			Attachable: n.Attachable,
			Ingress:    n.Ingress,
			Labels:     n.Labels,
		}
	}
	return result, nil
}

func (s *DockerService) CreateNetwork(ctx context.Context, req *model.NetworkCreateRequest) (map[string]string, error) {
	driver := req.Driver
	if driver == "" {
		driver = "bridge"
	}
	resp, err := s.client.CreateNetwork(ctx, req.Name, driver, req.Options, req.Labels, req.Internal, req.Attachable)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"id":       resp.ID,
		"warnings": resp.Warning,
	}, nil
}

func (s *DockerService) RemoveNetwork(ctx context.Context, id string) error {
	return s.client.RemoveNetwork(ctx, id)
}

func (s *DockerService) ConnectNetwork(ctx context.Context, id string, req *model.NetworkConnectRequest) error {
	return s.client.ConnectNetwork(ctx, id, req.ContainerID, req.Aliases, req.IPv4Address, req.IPv6Address)
}

func (s *DockerService) DisconnectNetwork(ctx context.Context, id string, req *model.NetworkDisconnectRequest) error {
	return s.client.DisconnectNetwork(ctx, id, req.ContainerID, req.Force)
}

func (s *DockerService) PruneNetworks(ctx context.Context) (map[string]interface{}, error) {
	report, err := s.client.PruneNetworks(ctx, nil)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"networks_deleted": report.NetworksDeleted,
	}, nil
}

// ==================== Volume Service ====================

func (s *DockerService) ListVolumes(ctx context.Context) (*model.VolumeListResponse, error) {
	volList, err := s.client.ListVolumes(ctx, nil)
	if err != nil {
		return nil, err
	}

	result := &model.VolumeListResponse{
		Warnings: volList.Warnings,
	}
	for _, v := range volList.Volumes {
		result.Volumes = append(result.Volumes, &model.Volume{
			Name:       v.Name,
			Driver:     v.Driver,
			Mountpoint: v.Mountpoint,
			CreatedAt:  v.CreatedAt,
			Status:     v.Status,
			Labels:     v.Labels,
			Scope:      v.Scope,
			Options:    v.Options,
		})
	}
	return result, nil
}

func (s *DockerService) CreateVolume(ctx context.Context, req *model.VolumeCreateRequest) (*model.Volume, error) {
	driver := req.Driver
	if driver == "" {
		driver = "local"
	}
	v, err := s.client.CreateVolume(ctx, req.Name, driver, req.DriverOpts, req.Labels)
	if err != nil {
		return nil, err
	}
	return &model.Volume{
		Name:       v.Name,
		Driver:     v.Driver,
		Mountpoint: v.Mountpoint,
		CreatedAt:  v.CreatedAt,
		Status:     v.Status,
		Labels:     v.Labels,
		Scope:      v.Scope,
		Options:    v.Options,
	}, nil
}

func (s *DockerService) RemoveVolume(ctx context.Context, name string, force bool) error {
	return s.client.RemoveVolume(ctx, name, force)
}

func (s *DockerService) PruneVolumes(ctx context.Context) (map[string]interface{}, error) {
	report, err := s.client.PruneVolumes(ctx, nil)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"volumes_deleted": report.VolumesDeleted,
		"space_reclaimed": report.SpaceReclaimed,
	}, nil
}

// ==================== 转换函数 ====================

func convertContainer(c types.Container) model.Container {
	ports := make([]model.Port, len(c.Ports))
	for i, p := range c.Ports {
		ports[i] = model.Port{
			IP:          p.IP,
			PrivatePort: p.PrivatePort,
			PublicPort:  p.PublicPort,
			Type:        p.Type,
		}
	}

	return model.Container{
		ID:      c.ID,
		Names:   c.Names,
		Image:   c.Image,
		State:   c.State,
		Status:  c.Status,
		Ports:   ports,
		Created: c.Created,
	}
}

func convertContainerInspect(c types.ContainerJSON) *model.ContainerDetail {
	resp := &model.ContainerDetail{
		ID:    c.ID,
		Name:  c.Name,
		Image: c.Config.Image,
	}

	if c.State != nil {
		resp.State = model.ContainerState{
			Status:     c.State.Status,
			Running:    c.State.Running,
			Paused:     c.State.Paused,
			Restarting: c.State.Restarting,
			StartedAt:  parseTime(c.State.StartedAt),
			FinishedAt: parseTime(c.State.FinishedAt),
		}
	}

	if c.Config != nil {
		resp.Config = map[string]interface{}{
			"hostname": c.Config.Hostname,
			"env":      c.Config.Env,
			"cmd":      c.Config.Cmd,
			"image":    c.Config.Image,
			"labels":   c.Config.Labels,
		}
	}

	resp.HostConfig = c.HostConfig

	if len(c.Mounts) > 0 {
		mounts := make([]model.MountPoint, len(c.Mounts))
		for i, m := range c.Mounts {
			mounts[i] = model.MountPoint{
				Type:        string(m.Type),
				Name:        m.Name,
				Source:      m.Source,
				Destination: m.Destination,
				Mode:        m.Mode,
				RW:          m.RW,
			}
		}
		resp.Mounts = mounts
	}

	if c.NetworkSettings != nil {
		ports := make(map[string][]model.PortBinding)
		for port, bindings := range c.NetworkSettings.Ports {
			portKey := string(port)
			ports[portKey] = make([]model.PortBinding, len(bindings))
			for i, b := range bindings {
				ports[portKey][i] = model.PortBinding{
					HostIP:   b.HostIP,
					HostPort: b.HostPort,
				}
			}
		}

		resp.NetworkSettings = &model.NetworkSettingsDetail{
			Bridge:                 c.NetworkSettings.Bridge,
			SandboxID:              c.NetworkSettings.SandboxID,
			HairpinMode:            c.NetworkSettings.HairpinMode,
			LinkLocalIPv6Address:   c.NetworkSettings.LinkLocalIPv6Address,
			LinkLocalIPv6PrefixLen: c.NetworkSettings.LinkLocalIPv6PrefixLen,
			Ports:                  ports,
			SandboxKey:             c.NetworkSettings.SandboxKey,
			SecondaryIPAddresses:   c.NetworkSettings.SecondaryIPAddresses,
			SecondaryIPv6Addresses: c.NetworkSettings.SecondaryIPv6Addresses,
			EndpointID:             c.NetworkSettings.EndpointID,
			Gateway:                c.NetworkSettings.Gateway,
			GlobalIPv6Address:      c.NetworkSettings.GlobalIPv6Address,
			GlobalIPv6PrefixLen:    c.NetworkSettings.GlobalIPv6PrefixLen,
			IPAddress:              c.NetworkSettings.IPAddress,
			IPPrefixLen:            c.NetworkSettings.IPPrefixLen,
			IPv6Gateway:            c.NetworkSettings.IPv6Gateway,
			MacAddress:             c.NetworkSettings.MacAddress,
			Networks:               c.NetworkSettings.Networks,
		}
	}

	return resp
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, s)
	return t
}

func timeoutPtr(timeout int) *int {
	if timeout < 0 {
		return nil
	}
	return &timeout
}

var _ = strconv.Atoi // keep import if needed
var _ = timeoutPtr
