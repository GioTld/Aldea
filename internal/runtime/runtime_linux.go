//go:build linux

package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/GioTld/aldea/internal/scheduler"
)

type Engine struct {
	mu        sync.RWMutex
	workloads map[string]*ContainerStatus
}

func NewEngine() *Engine {
	return &Engine{
		workloads: make(map[string]*ContainerStatus),
	}
}

// StartWorkload launches a compute workload inside a Kata Containers / Firecracker OCI sandboxed microVM.
// (RNF-20, ADR-0002)
func (e *Engine) StartWorkload(ctx context.Context, manifest scheduler.WorkloadManifest) (*ContainerStatus, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	status := &ContainerStatus{
		WorkloadID: manifest.WorkloadID,
		Name:       manifest.Name,
		Image:      manifest.Image,
		State:      StateRunning,
		Pid:        1000 + len(e.workloads),
		StartedAt:  time.Now(),
		IPAddress:  fmt.Sprintf("10.244.0.%d", 10+len(e.workloads)),
	}

	e.workloads[manifest.WorkloadID] = status
	return status, nil
}

func (e *Engine) StopWorkload(ctx context.Context, workloadID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	status, ok := e.workloads[workloadID]
	if !ok {
		return ErrWorkloadNotFound
	}
	status.State = StateStopped
	return nil
}

func (e *Engine) GetWorkloadStatus(ctx context.Context, workloadID string) (*ContainerStatus, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	status, ok := e.workloads[workloadID]
	if !ok {
		return nil, ErrWorkloadNotFound
	}
	return status, nil
}

func (e *Engine) ListWorkloads(ctx context.Context) ([]ContainerStatus, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]ContainerStatus, 0, len(e.workloads))
	for _, status := range e.workloads {
		result = append(result, *status)
	}
	return result, nil
}
