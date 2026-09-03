package tracker_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/GioTld/aldea/internal/tracker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrackerStore(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tracker_test.db")

	store, err := tracker.NewStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	t.Run("register and list nodes", func(t *testing.T) {
		nodeA := tracker.NodeMetadata{
			NodeID:           "node-a",
			Address:          "192.168.1.10:9090",
			StorageAllocated: 1000000,
			StorageUsed:      10000,
			LastSeen:         time.Now().Unix(),
			IsHealthy:        true,
		}
		nodeB := tracker.NodeMetadata{
			NodeID:           "node-b",
			Address:          "192.168.1.11:9090",
			StorageAllocated: 2000000,
			StorageUsed:      20000,
			LastSeen:         time.Now().Unix(),
			IsHealthy:        true,
		}

		require.NoError(t, store.RegisterNode(nodeA))
		require.NoError(t, store.RegisterNode(nodeB))

		nodes, err := store.ListNodes()
		require.NoError(t, err)
		assert.Len(t, nodes, 2)
	})

	t.Run("select distinct placement nodes", func(t *testing.T) {
		for i := 1; i <= 5; i++ {
			node := tracker.NodeMetadata{
				NodeID:           "node-" + string(rune('a'+i)),
				Address:          "10.0.0." + string(rune('0'+i)) + ":9090",
				StorageAllocated: 5000000,
				StorageUsed:      5000,
				LastSeen:         time.Now().Unix(),
				IsHealthy:        true,
			}
			require.NoError(t, store.RegisterNode(node))
		}

		selected, err := store.SelectPlacementNodes(4, 100000)
		require.NoError(t, err)
		assert.Len(t, selected, 4)

		unique := make(map[string]bool)
		for _, n := range selected {
			unique[n.NodeID] = true
		}
		assert.Len(t, unique, 4)
	})

	t.Run("select placement nodes fails if not enough nodes", func(t *testing.T) {
		_, err := store.SelectPlacementNodes(100, 10000)
		assert.ErrorIs(t, err, tracker.ErrInsufficientNodes)
	})

	t.Run("save and get file placement", func(t *testing.T) {
		fileMeta := tracker.FileMetadata{
			FileID:     "file-101",
			FileName:   "photo.jpg",
			Size:       204800,
			ChunkCount: 1,
			CreatedAt:  time.Now().Unix(),
		}
		placements := []tracker.ShardPlacement{
			{ShardID: "shard-1", FileID: "file-101", ChunkIndex: 0, ShardIndex: 0, NodeID: "node-b"},
			{ShardID: "shard-2", FileID: "file-101", ChunkIndex: 0, ShardIndex: 1, NodeID: "node-c"},
		}

		require.NoError(t, store.SaveFilePlacement(fileMeta, placements))

		file, gotPlacements, err := store.GetFilePlacement("file-101")
		require.NoError(t, err)
		assert.Equal(t, "photo.jpg", file.FileName)
		assert.Len(t, gotPlacements, 2)
	})

	t.Run("get non existent file returns error", func(t *testing.T) {
		_, _, err := store.GetFilePlacement("non-existent")
		assert.ErrorIs(t, err, tracker.ErrFileNotFound)
	})
}
