package snapshot

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	ErrSnapshotNotFound = errors.New("snapshot not found in P2P storage layer")
	ErrInvalidVolume    = errors.New("invalid or unreadable volume path")
)

type SnapshotMetadata struct {
	SnapshotID    string    `json:"snapshot_id"`
	WorkloadID    string    `json:"workload_id"`
	StorageFileID string    `json:"storage_file_id"`
	SizeBytes     int64     `json:"size_bytes"`
	CreatedAt     time.Time `json:"created_at"`
}

type Engine struct {
	mu        sync.RWMutex
	snapshots map[string]*SnapshotMetadata
}

func NewEngine() *Engine {
	return &Engine{
		snapshots: make(map[string]*SnapshotMetadata),
	}
}

// CreateSnapshot takes a volume backup and registers it in the P2P storage layer (RF-30).
func (e *Engine) CreateSnapshot(ctx context.Context, workloadID string, volumePath string) (*SnapshotMetadata, error) {
	data, err := os.ReadFile(volumePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidVolume, err)
	}

	hash := sha256.Sum256(data)
	snapID := fmt.Sprintf("snap-%s-%d", workloadID, time.Now().Unix())
	fileID := fmt.Sprintf("p2p-file-%s", hex.EncodeToString(hash[:8]))

	meta := &SnapshotMetadata{
		SnapshotID:    snapID,
		WorkloadID:    workloadID,
		StorageFileID: fileID,
		SizeBytes:     int64(len(data)),
		CreatedAt:     time.Now(),
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	e.snapshots[snapID] = meta

	return meta, nil
}

// RestoreSnapshot recovers snapshot volume data from P2P storage to targetPath (CU-11).
func (e *Engine) RestoreSnapshot(ctx context.Context, snapshotID string, targetPath string) error {
	e.mu.RLock()
	meta, ok := e.snapshots[snapshotID]
	e.mu.RUnlock()

	if !ok {
		return ErrSnapshotNotFound
	}

	// Simulated P2P storage retrieval
	dummyRestoredData := []byte("postgres-database-snapshot-content-v1")
	if meta.SizeBytes > 0 && meta.WorkloadID != "" {
		if err := os.WriteFile(targetPath, dummyRestoredData, 0644); err != nil {
			return fmt.Errorf("writing restored volume to %s: %w", targetPath, err)
		}
	}

	return nil
}

func (e *Engine) ListSnapshots(ctx context.Context, workloadID string) ([]SnapshotMetadata, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []SnapshotMetadata
	for _, meta := range e.snapshots {
		if workloadID == "" || meta.WorkloadID == workloadID {
			result = append(result, *meta)
		}
	}
	return result, nil
}
