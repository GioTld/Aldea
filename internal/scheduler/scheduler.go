package scheduler

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

var (
	ErrNoEligibleComputeNode  = errors.New("no eligible healthy compute node found for workload placement")
	ErrStatefulWorkloadPinned = errors.New("stateful workloads are pinned to their assigned node and cannot auto-reschedule (RF-29)")
)

type NodeComputeCapability struct {
	NodeID            string  `json:"node_id"`
	OS                string  `json:"os"`
	ComputeEnabled    bool    `json:"compute_enabled"`
	IsHealthy         bool    `json:"is_healthy"`
	TotalCPUCores     float64 `json:"total_cpu_cores"`
	AllocatedCPUCores float64 `json:"allocated_cpu_cores"`
	TotalMemoryMB     int64   `json:"total_memory_mb"`
	AllocatedMemoryMB int64   `json:"allocated_memory_mb"`
}

func (n *NodeComputeCapability) AvailableCPU() float64 {
	free := n.TotalCPUCores - n.AllocatedCPUCores
	if free < 0 {
		return 0
	}
	return free
}

func (n *NodeComputeCapability) AvailableMemoryMB() int64 {
	free := n.TotalMemoryMB - n.AllocatedMemoryMB
	if free < 0 {
		return 0
	}
	return free
}

type Scheduler struct {
	mu sync.RWMutex
}

func NewScheduler() *Scheduler {
	return &Scheduler{}
}

// SelectNode evaluates candidate nodes and selects the optimal Linux compute node
// using a least-loaded placement strategy.
//
// Rules enforced:
// 1. RNF-20: Node OS must be Linux ("linux"). Windows/macOS nodes are rejected.
// 2. RNF-19: Node must have explicitly opted into compute (ComputeEnabled == true).
// 3. Node must be healthy (IsHealthy == true).
// 4. Node must have sufficient free CPU and RAM margins for the workload.
func (s *Scheduler) SelectNode(manifest WorkloadManifest, candidates []NodeComputeCapability) (*NodeComputeCapability, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var bestNode *NodeComputeCapability
	var maxScore float64 = -1.0

	for i := range candidates {
		c := &candidates[i]

		// Enforce RNF-20 (Linux-only compute capability)
		if strings.ToLower(c.OS) != "linux" {
			continue
		}

		// Enforce RNF-19 (Explicit compute opt-in)
		if !c.ComputeEnabled || !c.IsHealthy {
			continue
		}

		// Check resource capacity constraints
		freeCPU := c.AvailableCPU()
		freeMem := c.AvailableMemoryMB()

		if freeCPU < manifest.CPUCores || freeMem < manifest.MemoryMB {
			continue
		}

		// Calculate least-loaded score based on CPU and RAM margins
		cpuRatio := freeCPU / c.TotalCPUCores
		memRatio := float64(freeMem) / float64(c.TotalMemoryMB)
		score := (cpuRatio + memRatio) / 2.0

		if score > maxScore {
			maxScore = score
			bestNode = c
		}
	}

	if bestNode == nil {
		return nil, fmt.Errorf("%w: requested CPU=%.2f, RAM=%dMB", ErrNoEligibleComputeNode, manifest.CPUCores, manifest.MemoryMB)
	}

	return bestNode, nil
}

// RescheduleOnNodeFailure automatically selects a replacement compute node for stateless workloads (RF-28, RF-32).
// Stateful workloads are explicitly rejected per RF-29 (pinned to their assigned node).
func (s *Scheduler) RescheduleOnNodeFailure(manifest WorkloadManifest, deadNodeID string, candidates []NodeComputeCapability) (*NodeComputeCapability, error) {
	if manifest.Type == WorkloadStateful {
		return nil, ErrStatefulWorkloadPinned
	}

	// Filter out the failed node from candidates
	filtered := make([]NodeComputeCapability, 0, len(candidates))
	for _, c := range candidates {
		if c.NodeID != deadNodeID {
			filtered = append(filtered, c)
		}
	}

	return s.SelectNode(manifest, filtered)
}
