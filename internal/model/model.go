package model

import (
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

// ========== Docker Models ==========

type Container struct {
	ID      string   `json:"id"`
	Names   []string `json:"names"`
	Image   string   `json:"image"`
	State   string   `json:"state"`
	Status  string   `json:"status"`
	Ports   []Port   `json:"ports,omitempty"`
	Created int64    `json:"created"`
}

type Port struct {
	IP          string `json:"ip,omitempty"`
	PrivatePort uint16 `json:"private_port"`
	PublicPort  uint16 `json:"public_port,omitempty"`
	Type        string `json:"type"`
}

type ContainerDetail struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Image           string                 `json:"image"`
	State           ContainerState         `json:"state"`
	Config          map[string]interface{} `json:"config,omitempty"`
	Network         map[string]interface{} `json:"network,omitempty"`
	HostConfig      *container.HostConfig  `json:"host_config,omitempty"`
	NetworkSettings *NetworkSettingsDetail `json:"network_settings,omitempty"`
	Mounts          []MountPoint           `json:"mounts,omitempty"`
}

type ContainerState struct {
	Status     string    `json:"status"`
	Running    bool      `json:"running"`
	Paused     bool      `json:"paused"`
	Restarting bool      `json:"restarting"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
}

type NetworkSettingsDetail struct {
	Bridge                 string                               `json:"bridge,omitempty"`
	SandboxID              string                               `json:"sandbox_id,omitempty"`
	HairpinMode            bool                                 `json:"hairpin_mode,omitempty"`
	LinkLocalIPv6Address   string                               `json:"link_local_ipv6_address,omitempty"`
	LinkLocalIPv6PrefixLen int                                  `json:"link_local_ipv6_prefix_len,omitempty"`
	Ports                  map[string][]PortBinding             `json:"ports,omitempty"`
	SandboxKey             string                               `json:"sandbox_key,omitempty"`
	SecondaryIPAddresses   interface{}                          `json:"secondary_ip_addresses,omitempty"`
	SecondaryIPv6Addresses interface{}                          `json:"secondary_ipv6_addresses,omitempty"`
	EndpointID             string                               `json:"endpoint_id,omitempty"`
	Gateway                string                               `json:"gateway,omitempty"`
	GlobalIPv6Address      string                               `json:"global_ipv6_address,omitempty"`
	GlobalIPv6PrefixLen    int                                  `json:"global_ipv6_prefix_len,omitempty"`
	IPAddress              string                               `json:"ip_address,omitempty"`
	IPPrefixLen            int                                  `json:"ip_prefix_len,omitempty"`
	IPv6Gateway            string                               `json:"ipv6_gateway,omitempty"`
	MacAddress             string                               `json:"mac_address,omitempty"`
	Networks               map[string]*network.EndpointSettings `json:"networks,omitempty"`
}

type PortBinding struct {
	HostIP   string `json:"host_ip,omitempty"`
	HostPort string `json:"host_port,omitempty"`
}

type MountPoint struct {
	Type        string `json:"type"`
	Name        string `json:"name,omitempty"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Mode        string `json:"mode,omitempty"`
	RW          bool   `json:"rw"`
}

// ========== Image Models ==========

type Image struct {
	ID          string            `json:"id"`
	RepoTags    []string          `json:"repo_tags"`
	RepoDigests []string          `json:"repo_digests"`
	Created     int64             `json:"created"`
	Size        int64             `json:"size"`
	VirtualSize int64             `json:"virtual_size"`
	Labels      map[string]string `json:"labels,omitempty"`
}

type ImageDetail struct {
	ID           string            `json:"id"`
	RepoTags     []string          `json:"repo_tags"`
	RepoDigests  []string          `json:"repo_digests"`
	Parent       string            `json:"parent,omitempty"`
	Comment      string            `json:"comment,omitempty"`
	Created      string            `json:"created"`
	Container    string            `json:"container,omitempty"`
	Size         int64             `json:"size"`
	VirtualSize  int64             `json:"virtual_size"`
	Labels       map[string]string `json:"labels,omitempty"`
	Architecture string            `json:"architecture,omitempty"`
	Os           string            `json:"os,omitempty"`
}

type ImagePullRequest struct {
	Image        string `json:"image" binding:"required"`
	RegistryAuth string `json:"registry_auth,omitempty"`
	Insecure     bool   `json:"insecure,omitempty"`
	Platform     string `json:"platform,omitempty"`
}

// ========== Network Models ==========

type Network struct {
	Name       string            `json:"name"`
	ID         string            `json:"id"`
	Created    string            `json:"created"`
	Scope      string            `json:"scope"`
	Driver     string            `json:"driver"`
	EnableIPv6 bool              `json:"enable_ipv6"`
	Internal   bool              `json:"internal"`
	Attachable bool              `json:"attachable"`
	Ingress    bool              `json:"ingress"`
	Labels     map[string]string `json:"labels,omitempty"`
}

type NetworkCreateRequest struct {
	Name       string            `json:"name" binding:"required"`
	Driver     string            `json:"driver,omitempty"`
	Options    map[string]string `json:"options,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Internal   bool              `json:"internal,omitempty"`
	Attachable bool              `json:"attachable,omitempty"`
}

type NetworkConnectRequest struct {
	ContainerID string   `json:"container_id" binding:"required"`
	Aliases     []string `json:"aliases,omitempty"`
	IPv4Address string   `json:"ipv4_address,omitempty"`
	IPv6Address string   `json:"ipv6_address,omitempty"`
}

type NetworkDisconnectRequest struct {
	ContainerID string `json:"container_id" binding:"required"`
	Force       bool   `json:"force,omitempty"`
}

// ========== Volume Models ==========

type Volume struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Mountpoint string            `json:"mountpoint"`
	CreatedAt  string            `json:"created_at,omitempty"`
	Status     map[string]interface{} `json:"status,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Scope      string            `json:"scope"`
	Options    map[string]string `json:"options,omitempty"`
}

type VolumeListResponse struct {
	Volumes  []*Volume `json:"volumes"`
	Warnings []string  `json:"warnings,omitempty"`
}

type VolumeCreateRequest struct {
	Name       string            `json:"name" binding:"required"`
	Driver     string            `json:"driver,omitempty"`
	DriverOpts map[string]string `json:"driver_opts,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// ========== Container Create Models ==========

type ContainerCreateRequest struct {
	Name           string                      `json:"name" binding:"required"`
	Image          string                      `json:"image" binding:"required"`
	Cmd            []string                    `json:"cmd,omitempty"`
	Entrypoint     []string                    `json:"entrypoint,omitempty"`
	Env            []string                    `json:"env,omitempty"`
	Labels         map[string]string           `json:"labels,omitempty"`
	ExposedPorts   map[string]struct{}         `json:"exposed_ports,omitempty"`
	PortBindings   map[string][]PortBinding    `json:"port_bindings,omitempty"`
	Volumes        map[string]struct{}         `json:"volumes,omitempty"`
	Mounts         []MountConfig               `json:"mounts,omitempty"`
	NetworkMode    string                      `json:"network_mode,omitempty"`
	Networks       []ContainerNetworkConfig    `json:"networks,omitempty"`
	RestartPolicy  string `json:"restart_policy,omitempty"`
	RestartRetries *int   `json:"restart_retries,omitempty"`
	AutoRemove     bool                        `json:"auto_remove,omitempty"`
	Privileged     bool                        `json:"privileged,omitempty"`
	User           string                      `json:"user,omitempty"`
	WorkingDir     string                      `json:"working_dir,omitempty"`
	Hostname       string                      `json:"hostname,omitempty"`
	DNS            []string                    `json:"dns,omitempty"`
	Memory         int64                       `json:"memory,omitempty"`
	CPUQuota       int64                       `json:"cpu_quota,omitempty"`
	CPUPeriod      int64                       `json:"cpu_period,omitempty"`
	CPUShares      int64                       `json:"cpu_shares,omitempty"`
	StdinOpen      bool                        `json:"stdin_open,omitempty"`
	Tty            bool                        `json:"tty,omitempty"`
}

type MountConfig struct {
	Type        string   `json:"type,omitempty"`
	Source      string   `json:"source,omitempty"`
	Target      string   `json:"target" binding:"required"`
	ReadOnly    bool     `json:"read_only,omitempty"`
	BindOptions *BindOptions `json:"bind_options,omitempty"`
}

type BindOptions struct {
	Propagation string `json:"propagation,omitempty"`
}

type ContainerNetworkConfig struct {
	NetworkID   string   `json:"network_id,omitempty"`
	NetworkName string   `json:"network_name,omitempty"`
	Aliases     []string `json:"aliases,omitempty"`
	IPv4Address string   `json:"ipv4_address,omitempty"`
	IPv6Address string   `json:"ipv6_address,omitempty"`
}

type ContainerRenameRequest struct {
	NewName string `json:"new_name" binding:"required"`
}

type ContainerExecRequest struct {
	Cmd          []string `json:"cmd" binding:"required"`
	WorkingDir   string   `json:"working_dir,omitempty"`
	Env          []string `json:"env,omitempty"`
	User         string   `json:"user,omitempty"`
	Privileged   bool     `json:"privileged,omitempty"`
	Tty          bool     `json:"tty,omitempty"`
	AttachStdin  bool     `json:"attach_stdin,omitempty"`
	AttachStdout bool     `json:"attach_stdout,omitempty"`
	AttachStderr bool     `json:"attach_stderr,omitempty"`
}

type ContainerExecResponse struct {
	ID       string `json:"id"`
	Output   string `json:"output,omitempty"`
	ExitCode int    `json:"exit_code,omitempty"`
}

// ========== System Metrics Models ==========

type SystemMetrics struct {
	CPU       CPUMetrics       `json:"cpu"`
	Memory    MemoryMetrics    `json:"memory"`
	Disk      []DiskMetrics    `json:"disk"`
	Network   []NetworkMetrics `json:"network"`
	Load      LoadMetrics      `json:"load"`
	Timestamp time.Time        `json:"timestamp"`
}

// MetricsCollectResponse 是 /api/v1/metrics/collect 的专用返回结构。
// 在 SystemMetrics 基础上增加运行时长字段。
type MetricsCollectResponse struct {
	CPU             CPUMetrics       `json:"cpu"`
	Memory          MemoryMetrics    `json:"memory"`
	Disk            []DiskMetrics    `json:"disk"`
	Network         []NetworkMetrics `json:"network"`
	Load            LoadMetrics      `json:"load"`
	Timestamp       time.Time        `json:"timestamp"`
	UptimeSeconds   float64          `json:"uptime_seconds"`
	UptimeFormatted string           `json:"uptime_formatted"`
}

type CPUMetrics struct {
	UsagePercent float64 `json:"usage_percent"`
	User         float64 `json:"user_seconds"`
	System       float64 `json:"system_seconds"`
	Idle         float64 `json:"idle_seconds"`
	Cores        int     `json:"cores"`
}

type MemoryMetrics struct {
	Total       uint64  `json:"total_bytes"`
	Available   uint64  `json:"available_bytes"`
	Used        uint64  `json:"used_bytes"`
	Free        uint64  `json:"free_bytes"`
	UsagePercent float64 `json:"usage_percent"`
}

type DiskMetrics struct {
	Filesystem string  `json:"filesystem"`
	Total      uint64  `json:"total_bytes"`
	Used       uint64  `json:"used_bytes"`
	Free       uint64  `json:"free_bytes"`
	UsagePercent float64 `json:"usage_percent"`
}

type NetworkMetrics struct {
	Interface   string `json:"interface"`
	ReceiveBytes   uint64 `json:"receive_bytes"`
	TransmitBytes  uint64 `json:"transmit_bytes"`
	ReceivePackets uint64 `json:"receive_packets"`
	TransmitPackets uint64 `json:"transmit_packets"`
}

type LoadMetrics struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

// ========== SSH File Management Models ==========

type SSHConnectRequest struct {
	Host       string `json:"host" binding:"required"`
	Port       int    `json:"port"`
	User       string `json:"user" binding:"required"`
	Password   string `json:"password,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
}

type SSHSessionResponse struct {
	ID        string `json:"id"`
	Host      string `json:"host"`
	User      string `json:"user"`
	CreatedAt int64  `json:"created_at"`
}

type SSHFileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"is_dir"`
	Mode    string `json:"mode"`
	ModTime int64  `json:"mod_time"`
}

type SSHExecRequest struct {
	Command string `json:"command" binding:"required"`
}

type SSHExecResponse struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
}

type SSHRenameRequest struct {
	OldPath string `json:"old_path" binding:"required"`
	NewPath string `json:"new_path" binding:"required"`
}

// ========== Common Response ==========

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
