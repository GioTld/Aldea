package metrics_test

import (
	"testing"
	"time"

	"github.com/GioTld/aldea/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsCollector(t *testing.T) {
	collector := metrics.NewCollector()

	t.Run("records storage usage metrics", func(t *testing.T) {
		collector.RecordStorage(100*1024*1024*1024, 25*1024*1024*1024)
		snap := collector.GetSnapshot()

		assert.Equal(t, int64(100*1024*1024*1024), snap.StorageAllocated)
		assert.Equal(t, int64(25*1024*1024*1024), snap.StorageUsed)
		assert.Equal(t, 25.0, snap.StoragePercent)
	})

	t.Run("records network bandwidth transfers", func(t *testing.T) {
		collector.RecordTransfer(1024, 2048)
		collector.RecordTransfer(512, 1024)

		snap := collector.GetSnapshot()
		assert.Equal(t, int64(1536), snap.BytesSentTotal)
		assert.Equal(t, int64(3072), snap.BytesReceivedTotal)
	})

	t.Run("records peer health and latencies", func(t *testing.T) {
		collector.RecordPeerStatus("peer-1", 15*time.Millisecond, true)
		collector.RecordPeerStatus("peer-2", 45*time.Millisecond, true)
		collector.RecordPeerStatus("peer-3", 0, false)

		snap := collector.GetSnapshot()
		require.Len(t, snap.Peers, 3)
		assert.True(t, snap.Peers["peer-1"].IsHealthy)
		assert.Equal(t, int64(15), snap.Peers["peer-1"].LatencyMs)
		assert.False(t, snap.Peers["peer-3"].IsHealthy)
	})
}
