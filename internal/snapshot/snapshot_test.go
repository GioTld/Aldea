package snapshot_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/GioTld/aldea/internal/snapshot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSnapshotEngine(t *testing.T) {
	ctx := context.Background()
	engine := snapshot.NewEngine()

	tempDir := t.TempDir()
	sourceVol := filepath.Join(tempDir, "postgres_data.db")
	err := os.WriteFile(sourceVol, []byte("postgres-database-snapshot-content-v1"), 0644)
	require.NoError(t, err)

	t.Run("creates periodic volume snapshot to P2P storage (RF-30)", func(t *testing.T) {
		meta, err := engine.CreateSnapshot(ctx, "wl-db-01", sourceVol)
		require.NoError(t, err)
		assert.Equal(t, "wl-db-01", meta.WorkloadID)
		assert.NotEmpty(t, meta.SnapshotID)
		assert.NotEmpty(t, meta.StorageFileID)
	})

	t.Run("restores snapshot to replacement node target volume (CU-11)", func(t *testing.T) {
		meta, err := engine.CreateSnapshot(ctx, "wl-db-02", sourceVol)
		require.NoError(t, err)

		targetVol := filepath.Join(tempDir, "restored_postgres.db")
		err = engine.RestoreSnapshot(ctx, meta.SnapshotID, targetVol)
		require.NoError(t, err)

		data, err := os.ReadFile(targetVol)
		require.NoError(t, err)
		assert.Equal(t, "postgres-database-snapshot-content-v1", string(data))
	})
}
