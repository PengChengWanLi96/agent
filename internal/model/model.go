package model

import "time"

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
	ID      string                 `json:"id"`
	Name    string                 `json:"name"`
	Image   string                 `json:"image"`
	State   ContainerState         `json:"state"`
	Config  map[string]interface{} `json:"config,omitempty"`
	Network map[string]interface{} `json:"network,omitempty"`
}

type ContainerState struct {
	Status     string    `json:"status"`
	Running    bool      `json:"running"`
	Paused     bool      `json:"paused"`
	Restarting bool      `json:"restarting"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
}

// ========== System Metrics Models ==========

type SystemMetrics struct {
	CPU       CPUMetrics       `json:"cpu"`
	Memory    MemoryMetrics    `json:"memory"`
	Disk      []DiskMetrics    `json:"disk"`
	Network   []NetworkMetrics `json:"network"`
	Load      LoadMetrics      `json:"load"`
	Uptime    float64          `json:"uptime_seconds"`
	Timestamp time.Time        `json:"timestamp"`
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

// ========== Common Response ==========

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
