package tracker_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/GioTld/aldea/internal/tracker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLivenessEvaluator(t *testing.T) {
	now := time.Now().Unix()

	t.Run("recently seen node is healthy", func(t *testing.T) {
		node := tracker.NodeMetadata{
			NodeID:   "node-a",
			LastSeen: now - 10,
			IsHealthy: true,
		}
		e := tracker.NewLivenessEvaluator(tracker.DefaultLivenessConfig())
		status := e.Evaluate(node, now)
		assert.Equal(t, tracker.StatusHealthy, status)
	})

	t.Run("node silent within transient window is transient", func(t *testing.T) {
		node := tracker.NodeMetadata{
			NodeID:   "node-b",
			LastSeen: now - 90,
			IsHealthy: true,
		}
		e := tracker.NewLivenessEvaluator(tracker.DefaultLivenessConfig())
		status := e.Evaluate(node, now)
		assert.Equal(t, tracker.StatusTransient, status)
	})

	t.Run("node silent beyond dead threshold is dead", func(t *testing.T) {
		node := tracker.NodeMetadata{
			NodeID:   "node-c",
			LastSeen: now - 700,
			IsHealthy: true,
		}
		e := tracker.NewLivenessEvaluator(tracker.DefaultLivenessConfig())
		status := e.Evaluate(node, now)
		assert.Equal(t, tracker.StatusDead, status)
	})

	t.Run("already marked unhealthy node is dead regardless of lastSeen", func(t *testing.T) {
		node := tracker.NodeMetadata{
			NodeID:    "node-d",
			LastSeen:  now - 5,
			IsHealthy: false,
		}
		e := tracker.NewLivenessEvaluator(tracker.DefaultLivenessConfig())
		status := e.Evaluate(node, now)
		assert.Equal(t, tracker.StatusDead, status)
	})

	t.Run("custom thresholds respected", func(t *testing.T) {
		cfg := tracker.LivenessConfig{
			TransientAfter: 30 * time.Second,
			DeadAfter:      120 * time.Second,
		}
		e := tracker.NewLivenessEvaluator(cfg)

		healthy := tracker.NodeMetadata{LastSeen: now - 10, IsHealthy: true}
		assert.Equal(t, tracker.StatusHealthy, e.Evaluate(healthy, now))

		transient := tracker.NodeMetadata{LastSeen: now - 60, IsHealthy: true}
		assert.Equal(t, tracker.StatusTransient, e.Evaluate(transient, now))

		dead := tracker.NodeMetadata{LastSeen: now - 200, IsHealthy: true}
		assert.Equal(t, tracker.StatusDead, e.Evaluate(dead, now))
	})
}

func TestStore_ApplyLiveness(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := tracker.NewStore(filepath.Join(tmpDir, "tracker.db"))
	require.NoError(t, err)
	defer store.Close()

	now := time.Now().Unix()

	nodes := []tracker.NodeMetadata{
		{NodeID: "alive", Address: "10.0.0.1:9001", StorageAllocated: 1e9, LastSeen: now - 5, IsHealthy: true},
		{NodeID: "flapping", Address: "10.0.0.2:9001", StorageAllocated: 1e9, LastSeen: now - 90, IsHealthy: true},
		{NodeID: "gone", Address: "10.0.0.3:9001", StorageAllocated: 1e9, LastSeen: now - 800, IsHealthy: true},
	}
	for _, n := range nodes {
		require.NoError(t, store.RegisterNode(n))
	}

	cfg := tracker.DefaultLivenessConfig()
	changed, err := store.ApplyLiveness(cfg)
	require.NoError(t, err)
	assert.Equal(t, 2, changed)

	all, err := store.ListNodes()
	require.NoError(t, err)

	byID := make(map[string]tracker.NodeMetadata)
	for _, n := range all {
		byID[n.NodeID] = n
	}

	assert.True(t, byID["alive"].IsHealthy)
	assert.False(t, byID["flapping"].IsHealthy)
	assert.False(t, byID["gone"].IsHealthy)
}
