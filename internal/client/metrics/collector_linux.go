//go:build linux

package metrics

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/node_exporter/collector"
	"gopkg.in/alecthomas/kingpin.v2"
	"agent/internal/model"
)

// node_exporter 的各 collector 是否启用由 kingpin 标志的默认值决定
// （collectorState[c] = kingpin.Flag(...).Default(...).Bool()）。
// kingpin 只有在 Parse 之后才会把默认值写入这些 *bool；本项目作为库内嵌
// node_exporter，从不调用 kingpin.Parse，导致所有 collector 保持 false，
// 最终 registry.Gather() 返回空结果。这里用空参数显式触发一次解析，
// 让默认启用的 collector 生效。
func init() {
	kingpin.CommandLine.Terminate(nil)
	_, _ = kingpin.CommandLine.Parse([]string{})
}

type NodeExporterCollector struct {
	nodeCollector *collector.NodeCollector
	registry      *prometheus.Registry
}

func NewCollector() (*NodeExporterCollector, error) {
	logger := log.NewNopLogger()
	nc, err := collector.NewNodeCollector(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create node collector: %w", err)
	}

	if len(nc.Collectors) == 0 {
		return nil, fmt.Errorf("no node_exporter collectors enabled; default collectors were not initialized")
	}

	reg := prometheus.NewRegistry()
	if err := reg.Register(nc); err != nil {
		return nil, fmt.Errorf("failed to register collector: %w", err)
	}

	return &NodeExporterCollector{
		nodeCollector: nc,
		registry:      reg,
	}, nil
}

func (c *NodeExporterCollector) Collect() (*model.SystemMetrics, error) {
	metricFamilies, err := c.registry.Gather()
	if err != nil {
		return nil, err
	}

	m := &model.SystemMetrics{}
	var totalMem, availMem float64

	for _, mf := range metricFamilies {
		name := mf.GetName()
		for _, metric := range mf.GetMetric() {
			value := getMetricValue(metric)
			labels := getMetricLabels(metric)

			switch name {
			case "node_cpu_seconds_total":
				if mode, ok := labels["mode"]; ok && mode == "idle" {
					m.CPU.Idle += value
				} else if mode == "user" {
					m.CPU.User += value
				} else if mode == "system" {
					m.CPU.System += value
				}
			case "node_memory_MemTotal_bytes":
				totalMem = value
				m.Memory.Total = uint64(value)
			case "node_memory_MemAvailable_bytes":
				availMem = value
				m.Memory.Available = uint64(value)
			case "node_load1":
				m.Load.Load1 = value
			case "node_load5":
				m.Load.Load5 = value
			case "node_load15":
				m.Load.Load15 = value
			}

			if strings.HasPrefix(name, "node_network_") {
				if iface, ok := labels["device"]; ok && iface != "lo" {
					updateNetworkMetric(m, name, iface, value)
				}
			}

			if name == "node_filesystem_size_bytes" {
				if mountpoint, ok := labels["mountpoint"]; ok && mountpoint != "" {
					updateDiskMetric(m, mountpoint, "size", value)
				}
			}
			if name == "node_filesystem_avail_bytes" {
				if mountpoint, ok := labels["mountpoint"]; ok && mountpoint != "" {
					updateDiskMetric(m, mountpoint, "avail", value)
				}
			}
		}
	}

	if totalMem > 0 {
		m.Memory.Used = uint64(totalMem - availMem)
		m.Memory.UsagePercent = (totalMem - availMem) / totalMem * 100
	}

	// Calculate CPU cores from node_cpu_seconds_total labels
	m.CPU.Cores = countCPUCores(metricFamilies)

	return m, nil
}

func (c *NodeExporterCollector) CollectRaw() (string, error) {
	metricFamilies, err := c.registry.Gather()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	encoder := expfmt.NewEncoder(&buf, expfmt.FmtText)
	for _, mf := range metricFamilies {
		if err := encoder.Encode(mf); err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}

func getMetricValue(m *dto.Metric) float64 {
	if m.Counter != nil {
		return m.GetCounter().GetValue()
	}
	if m.Gauge != nil {
		return m.GetGauge().GetValue()
	}
	if m.Untyped != nil {
		return m.GetUntyped().GetValue()
	}
	return 0
}

func getMetricLabels(m *dto.Metric) map[string]string {
	labels := make(map[string]string)
	for _, lp := range m.GetLabel() {
		labels[lp.GetName()] = lp.GetValue()
	}
	return labels
}

func updateNetworkMetric(m *model.SystemMetrics, name, iface string, value float64) {
	found := false
	for i := range m.Network {
		if m.Network[i].Interface == iface {
			found = true
			if name == "node_network_receive_bytes_total" {
				m.Network[i].ReceiveBytes = uint64(value)
			} else if name == "node_network_transmit_bytes_total" {
				m.Network[i].TransmitBytes = uint64(value)
			} else if name == "node_network_receive_packets_total" {
				m.Network[i].ReceivePackets = uint64(value)
			} else if name == "node_network_transmit_packets_total" {
				m.Network[i].TransmitPackets = uint64(value)
			}
			break
		}
	}
	if !found {
		nm := model.NetworkMetrics{Interface: iface}
		if name == "node_network_receive_bytes_total" {
			nm.ReceiveBytes = uint64(value)
		} else if name == "node_network_transmit_bytes_total" {
			nm.TransmitBytes = uint64(value)
		} else if name == "node_network_receive_packets_total" {
			nm.ReceivePackets = uint64(value)
		} else if name == "node_network_transmit_packets_total" {
			nm.TransmitPackets = uint64(value)
		}
		m.Network = append(m.Network, nm)
	}
}

func updateDiskMetric(m *model.SystemMetrics, mountpoint, metricType string, value float64) {
	found := false
	for i := range m.Disk {
		if m.Disk[i].Filesystem == mountpoint {
			found = true
			if metricType == "size" {
				m.Disk[i].Total = uint64(value)
			} else if metricType == "avail" {
				m.Disk[i].Free = uint64(value)
			}
			if m.Disk[i].Total > 0 {
				m.Disk[i].Used = m.Disk[i].Total - m.Disk[i].Free
				m.Disk[i].UsagePercent = float64(m.Disk[i].Used) / float64(m.Disk[i].Total) * 100
			}
			break
		}
	}
	if !found {
		dm := model.DiskMetrics{Filesystem: mountpoint}
		if metricType == "size" {
			dm.Total = uint64(value)
		} else if metricType == "avail" {
			dm.Free = uint64(value)
		}
		m.Disk = append(m.Disk, dm)
	}
}

func countCPUCores(metricFamilies []*dto.MetricFamily) int {
	cores := make(map[string]bool)
	for _, mf := range metricFamilies {
		if mf.GetName() == "node_cpu_seconds_total" {
			for _, m := range mf.GetMetric() {
				for _, lp := range m.GetLabel() {
					if lp.GetName() == "cpu" && lp.GetValue() != "" {
						cores[lp.GetValue()] = true
					}
				}
			}
		}
	}
	return len(cores)
}
