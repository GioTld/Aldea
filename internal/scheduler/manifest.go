package scheduler

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type WorkloadType string

const (
	WorkloadStateless WorkloadType = "stateless"
	WorkloadStateful  WorkloadType = "stateful"
)

var (
	ErrInvalidManifest     = errors.New("invalid workload manifest: missing required fields")
	ErrInvalidWorkloadType = errors.New("invalid workload type: must be 'stateless' or 'stateful'")
)

type PortMapping struct {
	HostPort      int    `yaml:"host_port" json:"host_port"`
	ContainerPort int    `yaml:"container_port" json:"container_port"`
	Protocol      string `yaml:"protocol" json:"protocol"`
}

type WorkloadManifest struct {
	WorkloadID          string            `yaml:"workload_id" json:"workload_id"`
	Name                string            `yaml:"name" json:"name"`
	Image               string            `yaml:"image" json:"image"`
	Type                WorkloadType      `yaml:"type" json:"type"`
	CPUCores            float64           `yaml:"cpu_cores" json:"cpu_cores"`
	MemoryMB            int64             `yaml:"memory_mb" json:"memory_mb"`
	PortMappings        []PortMapping     `yaml:"ports,omitempty" json:"ports,omitempty"`
	Env                 map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	SnapshotIntervalSec int               `yaml:"snapshot_interval_sec,omitempty" json:"snapshot_interval_sec,omitempty"`
}

func (m *WorkloadManifest) Validate() error {
	if m.WorkloadID == "" || m.Name == "" || m.Image == "" {
		return ErrInvalidManifest
	}
	if m.CPUCores <= 0 || m.MemoryMB <= 0 {
		return fmt.Errorf("%w: cpu_cores and memory_mb must be greater than zero", ErrInvalidManifest)
	}

	wt := WorkloadType(strings.ToLower(string(m.Type)))
	if wt != WorkloadStateless && wt != WorkloadStateful {
		return ErrInvalidWorkloadType
	}
	m.Type = wt

	return nil
}

func ParseManifestYAML(data []byte) (*WorkloadManifest, error) {
	var m WorkloadManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing workload manifest yaml: %w", err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}
