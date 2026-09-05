package runtime_test

import (
	"context"
	"runtime"
	"testing"

	aldearuntime "github.com/GioTld/aldea/internal/runtime"
	"github.com/GioTld/aldea/internal/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_WorkloadLifecycle(t *testing.T) {
	ctx := context.Background()
	engine := aldearuntime.NewEngine()

	manifest := scheduler.WorkloadManifest{
		WorkloadID: "wl-test-runtime-1",
		Name:       "test-container",
		Image:      "alpine:latest",
		Type:       scheduler.WorkloadStateless,
		CPUCores:   1.0,
		MemoryMB:   256,
	}

	t.Run("behavior matches current OS platform", func(t *testing.T) {
		if runtime.GOOS == "linux" {
			// On Linux, engine should accept starting and stopping workloads
			status, err := engine.StartWorkload(ctx, manifest)
			require.NoError(t, err)
			assert.Equal(t, "wl-test-runtime-1", status.WorkloadID)
			assert.Equal(t, aldearuntime.StateRunning, status.State)

			// List workloads
			list, err := engine.ListWorkloads(ctx)
			require.NoError(t, err)
			assert.Len(t, list, 1)

			// Stop workload
			err = engine.StopWorkload(ctx, "wl-test-runtime-1")
			require.NoError(t, err)

			statusAfter, err := engine.GetWorkloadStatus(ctx, "wl-test-runtime-1")
			require.NoError(t, err)
			assert.Equal(t, aldearuntime.StateStopped, statusAfter.State)
		} else {
			// On non-Linux, engine returns platform not supported error
			_, err := engine.StartWorkload(ctx, manifest)
			require.ErrorIs(t, err, aldearuntime.ErrComputeNotSupportedOnPlatform)
		}
	})
}
