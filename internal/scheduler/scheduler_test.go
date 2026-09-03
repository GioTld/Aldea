package scheduler_test

import (
	"testing"

	"github.com/GioTld/aldea/internal/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduler_SelectNode(t *testing.T) {
	s := scheduler.NewScheduler()

	manifest := scheduler.WorkloadManifest{
		WorkloadID: "wl-web-1",
		Name:       "web-service",
		Image:      "caddy:alpine",
		Type:       scheduler.WorkloadStateless,
		CPUCores:   1.0,
		MemoryMB:   512,
	}

	t.Run("selects least-loaded eligible Linux node", func(t *testing.T) {
		candidates := []scheduler.NodeComputeCapability{
			{
				NodeID:            "linux-loaded",
				OS:                "linux",
				ComputeEnabled:    true,
				IsHealthy:         true,
				TotalCPUCores:     4.0,
				AllocatedCPUCores: 3.5, // 0.5 CPU free (too small for 1.0)
				TotalMemoryMB:     4096,
				AllocatedMemoryMB: 2048,
			},
			{
				NodeID:            "linux-busy",
				OS:                "linux",
				ComputeEnabled:    true,
				IsHealthy:         true,
				TotalCPUCores:     4.0,
				AllocatedCPUCores: 2.0, // 2.0 CPU free, 1024 RAM free
				TotalMemoryMB:     2096,
				AllocatedMemoryMB: 1072,
			},
			{
				NodeID:            "linux-optimal",
				OS:                "linux",
				ComputeEnabled:    true,
				IsHealthy:         true,
				TotalCPUCores:     8.0,
				AllocatedCPUCores: 1.0, // 7.0 CPU free, 6144 RAM free (least loaded)
				TotalMemoryMB:     8192,
				AllocatedMemoryMB: 2048,
			},
		}

		node, err := s.SelectNode(manifest, candidates)
		require.NoError(t, err)
		assert.Equal(t, "linux-optimal", node.NodeID)
	})

	t.Run("rejects non-Linux nodes per RNF-20", func(t *testing.T) {
		candidates := []scheduler.NodeComputeCapability{
			{
				NodeID:            "windows-node",
				OS:                "windows",
				ComputeEnabled:    true,
				IsHealthy:         true,
				TotalCPUCores:     16.0,
				AllocatedCPUCores: 0,
				TotalMemoryMB:     32768,
				AllocatedMemoryMB: 0,
			},
			{
				NodeID:            "mac-node",
				OS:                "darwin",
				ComputeEnabled:    true,
				IsHealthy:         true,
				TotalCPUCores:     10.0,
				AllocatedCPUCores: 0,
				TotalMemoryMB:     16384,
				AllocatedMemoryMB: 0,
			},
		}

		_, err := s.SelectNode(manifest, candidates)
		require.ErrorIs(t, err, scheduler.ErrNoEligibleComputeNode)
	})

	t.Run("rejects nodes where ComputeEnabled is false", func(t *testing.T) {
		candidates := []scheduler.NodeComputeCapability{
			{
				NodeID:            "linux-storage-only",
				OS:                "linux",
				ComputeEnabled:    false, // Storage only!
				IsHealthy:         true,
				TotalCPUCores:     8.0,
				AllocatedCPUCores: 0,
				TotalMemoryMB:     8192,
				AllocatedMemoryMB: 0,
			},
		}

		_, err := s.SelectNode(manifest, candidates)
		require.ErrorIs(t, err, scheduler.ErrNoEligibleComputeNode)
	})

	t.Run("rejects unhealthy nodes", func(t *testing.T) {
		candidates := []scheduler.NodeComputeCapability{
			{
				NodeID:            "linux-dead",
				OS:                "linux",
				ComputeEnabled:    true,
				IsHealthy:         false,
				TotalCPUCores:     8.0,
				AllocatedCPUCores: 0,
				TotalMemoryMB:     8192,
				AllocatedMemoryMB: 0,
			},
		}

		_, err := s.SelectNode(manifest, candidates)
		require.ErrorIs(t, err, scheduler.ErrNoEligibleComputeNode)
	})
}
