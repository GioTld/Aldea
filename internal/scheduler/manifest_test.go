package scheduler_test

import (
	"testing"

	"github.com/GioTld/aldea/internal/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkloadManifest(t *testing.T) {
	t.Run("parses valid stateless workload manifest", func(t *testing.T) {
		yamlData := `
workload_id: wl-api-01
name: my-web-api
image: nginx:alpine
type: stateless
cpu_cores: 1.5
memory_mb: 512
env:
  PORT: "8080"
`
		m, err := scheduler.ParseManifestYAML([]byte(yamlData))
		require.NoError(t, err)
		assert.Equal(t, "wl-api-01", m.WorkloadID)
		assert.Equal(t, scheduler.WorkloadStateless, m.Type)
		assert.Equal(t, 1.5, m.CPUCores)
		assert.Equal(t, int64(512), m.MemoryMB)
	})

	t.Run("parses valid stateful workload manifest", func(t *testing.T) {
		yamlData := `
workload_id: wl-db-01
name: postgres-db
image: postgres:15-alpine
type: stateful
cpu_cores: 2.0
memory_mb: 1024
snapshot_interval_sec: 300
`
		m, err := scheduler.ParseManifestYAML([]byte(yamlData))
		require.NoError(t, err)
		assert.Equal(t, scheduler.WorkloadStateful, m.Type)
		assert.Equal(t, 300, m.SnapshotIntervalSec)
	})

	t.Run("rejects manifest with invalid workload type", func(t *testing.T) {
		yamlData := `
workload_id: wl-bad
name: bad-app
image: ubuntu:latest
type: unknown_type
cpu_cores: 1.0
memory_mb: 256
`
		_, err := scheduler.ParseManifestYAML([]byte(yamlData))
		require.ErrorIs(t, err, scheduler.ErrInvalidWorkloadType)
	})

	t.Run("rejects manifest with missing required fields", func(t *testing.T) {
		yamlData := `
name: no-image-app
type: stateless
`
		_, err := scheduler.ParseManifestYAML([]byte(yamlData))
		require.ErrorIs(t, err, scheduler.ErrInvalidManifest)
	})
}
