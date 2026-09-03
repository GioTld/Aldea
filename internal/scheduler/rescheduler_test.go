package scheduler_test

import (
	"testing"

	"github.com/GioTld/aldea/internal/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduler_RescheduleOnNodeFailure(t *testing.T) {
	s := scheduler.NewScheduler()

	statelessWorkload := scheduler.WorkloadManifest{
		WorkloadID: "wl-api-failover",
		Name:       "web-api",
		Image:      "node:18-alpine",
		Type:       scheduler.WorkloadStateless,
		CPUCores:   1.0,
		MemoryMB:   512,
	}

	statefulWorkload := scheduler.WorkloadManifest{
		WorkloadID: "wl-redis-pinned",
		Name:       "redis-cache",
		Image:      "redis:7-alpine",
		Type:       scheduler.WorkloadStateful,
		CPUCores:   1.0,
		MemoryMB:   512,
	}

	candidates := []scheduler.NodeComputeCapability{
		{
			NodeID:            "dead-node-1",
			OS:                "linux",
			ComputeEnabled:    true,
			IsHealthy:         false, // Dead!
			TotalCPUCores:     4.0,
			AllocatedCPUCores: 0,
			TotalMemoryMB:     4096,
			AllocatedMemoryMB: 0,
		},
		{
			NodeID:            "healthy-node-2",
			OS:                "linux",
			ComputeEnabled:    true,
			IsHealthy:         true,
			TotalCPUCores:     8.0,
			AllocatedCPUCores: 1.0,
			TotalMemoryMB:     8192,
			AllocatedMemoryMB: 1024,
		},
	}

	t.Run("automatically reschedules stateless workload on node failure (RF-28, RF-32)", func(t *testing.T) {
		newTargetNode, err := s.RescheduleOnNodeFailure(statelessWorkload, "dead-node-1", candidates)
		require.NoError(t, err)
		assert.Equal(t, "healthy-node-2", newTargetNode.NodeID)
	})

	t.Run("rejects auto-rescheduling for stateful workload per RF-29", func(t *testing.T) {
		_, err := s.RescheduleOnNodeFailure(statefulWorkload, "dead-node-1", candidates)
		require.ErrorIs(t, err, scheduler.ErrStatefulWorkloadPinned)
	})
}
