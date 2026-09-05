package metrics

import (
	"sync"
	"time"
)

type PeerMetric struct {
	NodeID    string    `json:"node_id"`
	LatencyMs int64     `json:"latency_ms"`
	IsHealthy bool      `json:"is_healthy"`
	LastSeen  time.Time `json:"last_seen"`
}

type MetricsSnapshot struct {
	StorageAllocated   int64                 `json:"storage_allocated"`
	StorageUsed        int64                 `json:"storage_used"`
	StoragePercent     float64               `json:"storage_percent"`
	BytesSentTotal     int64                 `json:"bytes_sent_total"`
	BytesReceivedTotal int64                 `json:"bytes_received_total"`
	Peers              map[string]PeerMetric `json:"peers"`
}

type Collector struct {
	mu            sync.RWMutex
	storageAlloc  int64
	storageUsed   int64
	bytesSent     int64
	bytesReceived int64
	peers         map[string]PeerMetric
}

func NewCollector() *Collector {
	return &Collector{
		peers: make(map[string]PeerMetric),
	}
}

func (c *Collector) RecordStorage(allocated, used int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.storageAlloc = allocated
	c.storageUsed = used
}

func (c *Collector) RecordTransfer(sent, received int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bytesSent += sent
	c.bytesReceived += received
}

func (c *Collector) RecordPeerStatus(nodeID string, latency time.Duration, isHealthy bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.peers[nodeID] = PeerMetric{
		NodeID:    nodeID,
		LatencyMs: latency.Milliseconds(),
		IsHealthy: isHealthy,
		LastSeen:  time.Now(),
	}
}

func (c *Collector) GetSnapshot() MetricsSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	pct := 0.0
	if c.storageAlloc > 0 {
		pct = (float64(c.storageUsed) / float64(c.storageAlloc)) * 100.0
	}

	peersCopy := make(map[string]PeerMetric, len(c.peers))
	for k, v := range c.peers {
		peersCopy[k] = v
	}

	return MetricsSnapshot{
		StorageAllocated:   c.storageAlloc,
		StorageUsed:        c.storageUsed,
		StoragePercent:     pct,
		BytesSentTotal:     c.bytesSent,
		BytesReceivedTotal: c.bytesReceived,
		Peers:              peersCopy,
	}
}
