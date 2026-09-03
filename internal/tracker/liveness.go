package tracker

import "time"

// NodeStatus classifies a node into three liveness buckets.
// Healthy  — seen recently, fully trusted.
// Transient — silent for a short while (router restart, sleep, etc.); kept
//             in the routing table but not selected for new placements.
// Dead      — silent long enough that permanent failure is assumed; removed
//             from healthy sets and triggers shard repair.
type NodeStatus uint8

const (
	StatusHealthy   NodeStatus = iota
	StatusTransient            // brief absence: exclude from placements, do not repair yet
	StatusDead                 // prolonged absence: mark unhealthy, trigger repair
)

// LivenessConfig holds the two thresholds that separate the three zones.
type LivenessConfig struct {
	TransientAfter time.Duration // silence beyond which a node is Transient
	DeadAfter      time.Duration // silence beyond which a node is Dead
}

// DefaultLivenessConfig returns tunings suitable for residential home
// connections where router reboots and short outages are common.
// A node silent for up to 60 s is considered transient (router restart).
// Beyond 600 s (10 min) the node is considered permanently gone.
func DefaultLivenessConfig() LivenessConfig {
	return LivenessConfig{
		TransientAfter: 60 * time.Second,
		DeadAfter:      600 * time.Second,
	}
}

// LivenessEvaluator classifies a NodeMetadata snapshot against a wall-clock
// reference time without touching any storage layer.
type LivenessEvaluator struct {
	cfg LivenessConfig
}

func NewLivenessEvaluator(cfg LivenessConfig) *LivenessEvaluator {
	return &LivenessEvaluator{cfg: cfg}
}

// Evaluate returns the NodeStatus of node relative to nowUnix (Unix seconds).
func (e *LivenessEvaluator) Evaluate(node NodeMetadata, nowUnix int64) NodeStatus {
	if !node.IsHealthy {
		return StatusDead
	}

	silent := time.Duration(nowUnix-node.LastSeen) * time.Second
	switch {
	case silent < e.cfg.TransientAfter:
		return StatusHealthy
	case silent < e.cfg.DeadAfter:
		return StatusTransient
	default:
		return StatusDead
	}
}

// ApplyLiveness runs liveness evaluation against all registered nodes and
// persists any health-status changes back to the store. Returns the count of
// nodes whose IsHealthy field changed.
func (s *Store) ApplyLiveness(cfg LivenessConfig) (int, error) {
	nodes, err := s.ListNodes()
	if err != nil {
		return 0, err
	}

	eval := NewLivenessEvaluator(cfg)
	now := timeNow()
	changed := 0

	for _, n := range nodes {
		status := eval.Evaluate(n, now)
		shouldBeHealthy := status == StatusHealthy

		if n.IsHealthy != shouldBeHealthy {
			n.IsHealthy = shouldBeHealthy
			if err := s.RegisterNode(n); err != nil {
				return changed, err
			}
			changed++
		}
	}

	return changed, nil
}

// timeNow is a package-level hook so tests can override wall time if needed.
// In production it always returns time.Now().Unix().
var timeNow = func() int64 { return time.Now().Unix() }
