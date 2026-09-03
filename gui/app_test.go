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
}
