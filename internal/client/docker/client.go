package docker

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// Client 封装 Docker 官方 SDK，提供容器、镜像、网络、卷等管理能力。
type Client struct {
	cli *client.Client
}

// NewClient 根据配置创建 Docker 客户端。
func NewClient(host string, tlsVerify bool, certPath string) (*Client, error) {
	opts := []client.Opt{
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	}

	if host != "" {
		opts = append(opts, client.WithHost(host))
	} else {
		opts = append(opts, client.WithHost("unix:///var/run/docker.sock"))
	}

	if tlsVerify && certPath != "" {
		opts = append(opts, client.WithTLSClientConfig(
			certPath+"/ca.pem",
			certPath+"/cert.pem",
			certPath+"/key.pem",
		))
	} else if tlsVerify {
		// 仅启用 TLS 验证，不指定证书路径时走系统根证书
		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{},
			},
		}
		opts = append(opts, client.WithHTTPClient(httpClient))
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}
	return &Client{cli: cli}, nil
}

// Close 关闭 Docker 客户端。
func (c *Client) Close() error {
	if c.cli != nil {
		return c.cli.Close()
	}
	return nil
}

// Ping 测试 Docker 连接。
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.cli.Ping(ctx)
	return err
}

// toFilters 将 map 转换为 Docker filters.Args。
func toFilters(m map[string][]string) filters.Args {
	args := filters.NewArgs()
	for key, values := range m {
		for _, value := range values {
			args.Add(key, value)
		}
	}
	return args
}

// ==================== 容器管理 ====================

func (c *Client) ListContainers(ctx context.Context, all bool, filterMap map[string][]string) ([]types.Container, error) {
	opts := types.ContainerListOptions{All: all}
	if filterMap != nil {
		opts.Filters = toFilters(filterMap)
	}
	return c.cli.ContainerList(ctx, opts)
}

func (c *Client) InspectContainer(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return c.cli.ContainerInspect(ctx, containerID)
}

func (c *Client) CreateContainer(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (container.CreateResponse, error) {
	return c.cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, containerName)
}

func (c *Client) StartContainer(ctx context.Context, containerID string) error {
	return c.cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
}

func (c *Client) StopContainer(ctx context.Context, containerID string, timeout *int) error {
	var stopOpts container.StopOptions
	if timeout != nil {
		stopOpts.Timeout = timeout
	}
	return c.cli.ContainerStop(ctx, containerID, stopOpts)
}

func (c *Client) RestartContainer(ctx context.Context, containerID string, timeout *int) error {
	var stopOpts container.StopOptions
	if timeout != nil {
		stopOpts.Timeout = timeout
	}
	return c.cli.ContainerRestart(ctx, containerID, stopOpts)
}

func (c *Client) KillContainer(ctx context.Context, containerID, signal string) error {
	return c.cli.ContainerKill(ctx, containerID, signal)
}

func (c *Client) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	return c.cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: force})
}

func (c *Client) PauseContainer(ctx context.Context, containerID string) error {
	return c.cli.ContainerPause(ctx, containerID)
}

func (c *Client) UnpauseContainer(ctx context.Context, containerID string) error {
	return c.cli.ContainerUnpause(ctx, containerID)
}

func (c *Client) RenameContainer(ctx context.Context, containerID, newName string) error {
	return c.cli.ContainerRename(ctx, containerID, newName)
}

func (c *Client) GetContainerStats(ctx context.Context, containerID string, stream bool) (types.ContainerStats, error) {
	return c.cli.ContainerStats(ctx, containerID, stream)
}

func (c *Client) GetContainerLogs(ctx context.Context, containerID string, tail string, timestamps bool, follow bool, since string, until string) (io.ReadCloser, error) {
	opts := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: timestamps,
		Follow:     follow,
	}
	if tail != "" {
		opts.Tail = tail
	}
	if since != "" {
		opts.Since = since
	}
	if until != "" {
		opts.Until = until
	}
	return c.cli.ContainerLogs(ctx, containerID, opts)
}

func (c *Client) PruneContainers(ctx context.Context, filterMap map[string][]string) (types.ContainersPruneReport, error) {
	var args filters.Args
	if filterMap != nil {
		args = toFilters(filterMap)
	}
	return c.cli.ContainersPrune(ctx, args)
}

// ==================== 镜像管理 ====================

func (c *Client) ListImages(ctx context.Context, all bool, filterMap map[string][]string) ([]types.ImageSummary, error) {
	opts := types.ImageListOptions{All: all}
	if filterMap != nil {
		opts.Filters = toFilters(filterMap)
	}
	return c.cli.ImageList(ctx, opts)
}

func (c *Client) InspectImage(ctx context.Context, imageID string) (types.ImageInspect, []byte, error) {
	return c.cli.ImageInspectWithRaw(ctx, imageID)
}

func (c *Client) PullImage(ctx context.Context, refStr string, registryAuth string, insecure bool, platform string) (io.ReadCloser, error) {
	opts := types.ImagePullOptions{}
	if registryAuth != "" {
		opts.RegistryAuth = registryAuth
	}
	if platform != "" {
		opts.Platform = platform
	}

	if insecure {
		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
		cli, err := client.NewClientWithOpts(
			client.FromEnv,
			client.WithAPIVersionNegotiation(),
			client.WithHTTPClient(httpClient),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create insecure docker client: %w", err)
		}
		defer cli.Close()
		return cli.ImagePull(ctx, refStr, opts)
	}

	return c.cli.ImagePull(ctx, refStr, opts)
}

func (c *Client) RemoveImage(ctx context.Context, imageID string, force bool) ([]types.ImageDeleteResponseItem, error) {
	return c.cli.ImageRemove(ctx, imageID, types.ImageRemoveOptions{Force: force})
}

func (c *Client) PruneImages(ctx context.Context, filterMap map[string][]string) (types.ImagesPruneReport, error) {
	var args filters.Args
	if filterMap != nil {
		args = toFilters(filterMap)
	}
	return c.cli.ImagesPrune(ctx, args)
}

// ==================== 网络管理 ====================

func (c *Client) ListNetworks(ctx context.Context, filterMap map[string][]string) ([]types.NetworkResource, error) {
	opts := types.NetworkListOptions{}
	if filterMap != nil {
		opts.Filters = toFilters(filterMap)
	}
	return c.cli.NetworkList(ctx, opts)
}

func (c *Client) InspectNetwork(ctx context.Context, networkID string) (types.NetworkResource, error) {
	return c.cli.NetworkInspect(ctx, networkID, types.NetworkInspectOptions{})
}

func (c *Client) CreateNetwork(ctx context.Context, name string, driver string, options map[string]string, labels map[string]string, internal bool, attachable bool) (types.NetworkCreateResponse, error) {
	config := types.NetworkCreate{
		Driver:     driver,
		Options:    options,
		Labels:     labels,
		Internal:   internal,
		Attachable: attachable,
	}
	return c.cli.NetworkCreate(ctx, name, config)
}

func (c *Client) RemoveNetwork(ctx context.Context, networkID string) error {
	return c.cli.NetworkRemove(ctx, networkID)
}

func (c *Client) PruneNetworks(ctx context.Context, filterMap map[string][]string) (types.NetworksPruneReport, error) {
	var args filters.Args
	if filterMap != nil {
		args = toFilters(filterMap)
	}
	return c.cli.NetworksPrune(ctx, args)
}

func (c *Client) ConnectNetwork(ctx context.Context, networkID, containerID string, aliases []string, ipv4Address, ipv6Address string) error {
	config := &network.EndpointSettings{Aliases: aliases}
	if ipv4Address != "" {
		config.IPAMConfig = &network.EndpointIPAMConfig{IPv4Address: ipv4Address}
	}
	if ipv6Address != "" {
		if config.IPAMConfig == nil {
			config.IPAMConfig = &network.EndpointIPAMConfig{}
		}
		config.IPAMConfig.IPv6Address = ipv6Address
	}
	return c.cli.NetworkConnect(ctx, networkID, containerID, config)
}

func (c *Client) DisconnectNetwork(ctx context.Context, networkID, containerID string, force bool) error {
	return c.cli.NetworkDisconnect(ctx, networkID, containerID, force)
}

// ==================== 卷管理 ====================

func (c *Client) ListVolumes(ctx context.Context, filterMap map[string][]string) (volume.ListResponse, error) {
	opts := volume.ListOptions{}
	if filterMap != nil {
		opts.Filters = toFilters(filterMap)
	}
	return c.cli.VolumeList(ctx, opts)
}

func (c *Client) InspectVolume(ctx context.Context, volumeName string) (volume.Volume, error) {
	return c.cli.VolumeInspect(ctx, volumeName)
}

func (c *Client) CreateVolume(ctx context.Context, name, driver string, driverOpts map[string]string, labels map[string]string) (*volume.Volume, error) {
	config := volume.CreateOptions{
		Name:       name,
		Driver:     driver,
		DriverOpts: driverOpts,
		Labels:     labels,
	}
	v, err := c.cli.VolumeCreate(ctx, config)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (c *Client) RemoveVolume(ctx context.Context, volumeName string, force bool) error {
	return c.cli.VolumeRemove(ctx, volumeName, force)
}

func (c *Client) PruneVolumes(ctx context.Context, filterMap map[string][]string) (types.VolumesPruneReport, error) {
	var args filters.Args
	if filterMap != nil {
		args = toFilters(filterMap)
	}
	return c.cli.VolumesPrune(ctx, args)
}

// ==================== Exec ====================

func (c *Client) CreateExec(ctx context.Context, containerID string, cmd []string, workingDir string, env []string, user string, privileged, tty, attachStdin, attachStdout, attachStderr bool) (types.IDResponse, error) {
	config := types.ExecConfig{
		Cmd:          cmd,
		WorkingDir:   workingDir,
		Env:          env,
		User:         user,
		Privileged:   privileged,
		Tty:          tty,
		AttachStdin:  attachStdin,
		AttachStdout: attachStdout,
		AttachStderr: attachStderr,
	}
	return c.cli.ContainerExecCreate(ctx, containerID, config)
}

func (c *Client) ExecAttach(ctx context.Context, execID string, tty bool) (types.HijackedResponse, error) {
	return c.cli.ContainerExecAttach(ctx, execID, types.ExecStartCheck{Tty: tty})
}

func (c *Client) ExecStart(ctx context.Context, execID string) error {
	return c.cli.ContainerExecStart(ctx, execID, types.ExecStartCheck{})
}

func (c *Client) ExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error) {
	return c.cli.ContainerExecInspect(ctx, execID)
}

func (c *Client) ResizeExec(ctx context.Context, execID string, height, width uint) error {
	return c.cli.ContainerExecResize(ctx, execID, types.ResizeOptions{Height: height, Width: width})
}

// StdCopy 用于解析非 TTY exec 的多路复用输出。
var StdCopy = stdcopy.StdCopy
