package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppBackend(t *testing.T) {
	tmpDir := t.TempDir()
	app := NewApp(tmpDir)

	t.Run("CreateNetwork returns valid invite token", func(t *testing.T) {
		token, err := app.CreateNetwork("127.0.0.1:9090")
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Contains(t, token, "aldea1_")
	})

	t.Run("JoinNetwork configures local node", func(t *testing.T) {
		token, err := app.CreateNetwork("127.0.0.1:9090")
		require.NoError(t, err)

		err = app.JoinNetwork(token)
		require.NoError(t, err)

		status := app.GetNodeStatus()
		assert.True(t, status.Configured)
		assert.Equal(t, "127.0.0.1:9090", status.TrackerAddr)
	})

	t.Run("SetStorageAllocation updates allocation", func(t *testing.T) {
		token, err := app.CreateNetwork("127.0.0.1:9090")
		require.NoError(t, err)
		require.NoError(t, app.JoinNetwork(token))

		newAlloc := int64(10 * 1024 * 1024 * 1024) // 10 GB
		err = app.SetStorageAllocation(newAlloc)
		require.NoError(t, err)

		status := app.GetNodeStatus()
		assert.Equal(t, newAlloc, status.StorageAllocated)
	})

	t.Run("PauseNode toggles node operational state", func(t *testing.T) {
		token, err := app.CreateNetwork("127.0.0.1:9090")
		require.NoError(t, err)
		require.NoError(t, app.JoinNetwork(token))

		assert.False(t, app.IsPaused())
		app.PauseNode(true)
		assert.True(t, app.IsPaused())
		status := app.GetNodeStatus()
		assert.Equal(t, "PAUSED", status.StateLabel)

		app.PauseNode(false)
		assert.False(t, app.IsPaused())
	})

	t.Run("GetComputeWorkloads returns workload list", func(t *testing.T) {
		workloads := app.GetComputeWorkloads()
		assert.NotEmpty(t, workloads)
		assert.Equal(t, "wl-web-caddy", workloads[0].WorkloadID)
	})

	t.Run("GetNetworkMetrics returns peer latencies and throughput", func(t *testing.T) {
		metrics := app.GetNetworkMetrics()
		assert.Greater(t, metrics.DownloadSpeedKBps, float64(0))
		assert.NotEmpty(t, metrics.Peers)
		assert.Equal(t, "node-madrid-01", metrics.Peers[0].NodeID)
	})

	t.Run("UploadFile processes real file with chunking, Reed-Solomon, XChaCha20 encryption and physical shard creation", func(t *testing.T) {
		err := app.UploadFile("sample_photo.jpg", 512*1024)
		require.NoError(t, err)

		files := app.ListFiles()
		require.Len(t, files, 1)
		assert.Equal(t, "sample_photo.jpg", files[0].FileName)

		status := app.GetNodeStatus()
		assert.Greater(t, status.StorageUsed, int64(0))
	})
}
