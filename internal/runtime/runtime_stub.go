//go:build !linux

package runtime

import (
	"context"

	"github.com/GioTld/aldea/internal/scheduler"
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) StartWorkload(ctx context.Context, manifest scheduler.WorkloadManifest) (*ContainerStatus, error) {
	return nil, ErrComputeNotSupportedOnPlatform
}

func (e *Engine) StopWorkload(ctx context.Context, workloadID string) error {
	return ErrComputeNotSupportedOnPlatform
}

func (e *Engine) GetWorkloadStatus(ctx context.Context, workloadID string) (*ContainerStatus, error) {
	return nil, ErrComputeNotSupportedOnPlatform
}

func (e *Engine) ListWorkloads(ctx context.Context) ([]ContainerStatus, error) {
	return nil, ErrComputeNotSupportedOnPlatform
}
