package runtime

import (
	"context"
	"errors"
	"time"

	"github.com/GioTld/aldea/internal/scheduler"
)

type ContainerState string

const (
	StateRunning ContainerState = "running"
	StateStopped ContainerState = "stopped"
	StateError   ContainerState = "error"
)

var (
	ErrComputeNotSupportedOnPlatform = errors.New("compute runtime execution is restricted to Linux operating systems (RNF-20)")
	ErrWorkloadNotFound              = errors.New("workload not found in local runtime engine")
)

type ContainerStatus struct {
	WorkloadID string         `json:"workload_id"`
	Name       string         `json:"name"`
	Image      string         `json:"image"`
	State      ContainerState `json:"state"`
	Pid        int            `json:"pid"`
	StartedAt  time.Time      `json:"started_at"`
	IPAddress  string         `json:"ip_address"`
}

type Runtime interface {
	StartWorkload(ctx context.Context, manifest scheduler.WorkloadManifest) (*ContainerStatus, error)
	StopWorkload(ctx context.Context, workloadID string) error
	GetWorkloadStatus(ctx context.Context, workloadID string) (*ContainerStatus, error)
	ListWorkloads(ctx context.Context) ([]ContainerStatus, error)
}
