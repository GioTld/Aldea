package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/GioTld/aldea/internal/chunker"
	"github.com/GioTld/aldea/internal/crypto"
	"github.com/GioTld/aldea/internal/erasure"
	"github.com/GioTld/aldea/internal/invite"
	"github.com/GioTld/aldea/internal/metrics"
	"github.com/GioTld/aldea/internal/runtime"
)

type NodeStatusDTO struct {
	Configured       bool   `json:"configured"`
	NodeID           string `json:"node_id"`
	TrackerAddr      string `json:"tracker_addr"`
	StorageAllocated int64  `json:"storage_allocated"`
	StorageUsed      int64  `json:"storage_used"`
	PeerCount        int    `json:"peer_count"`
	IsHealthy        bool   `json:"is_healthy"`
	StateLabel       string `json:"state_label"`
}

type FileDTO struct {
	FileID    string `json:"file_id"`
	FileName  string `json:"file_name"`
	Size      int64  `json:"size"`
	CreatedAt int64  `json:"created_at"`
}

type ComputeWorkloadDTO struct {
	WorkloadID string `json:"workload_id"`
	Name       string `json:"name"`
	Image      string `json:"image"`
	State      string `json:"state"`
	IPAddress  string `json:"ip_address"`
}

type PeerMetricDTO struct {
	NodeID    string `json:"node_id"`
	OS        string `json:"os"`
	LatencyMs int64  `json:"latency_ms"`
	IsHealthy bool   `json:"is_healthy"`
}

type NetworkMetricsDTO struct {
	UploadSpeedKBps   float64         `json:"upload_speed_kbps"`
	DownloadSpeedKBps float64         `json:"download_speed_kbps"`
	Peers             []PeerMetricDTO `json:"peers"`
}

type App struct {
	ctx              context.Context
	dataDir          string
	mu               sync.Mutex
	configured       bool
	paused           bool
	nodeID           string
	trackerAddr      string
	networkKey       []byte
	signingKey       []byte
	storageAlloc     int64
	files            []FileDTO
	runtimeEngine    *runtime.Engine
	metricsCollector *metrics.Collector
}

func NewApp(dataDir string) *App {
	signingKey := make([]byte, 32)
	rand.Read(signingKey)

	netKey := make([]byte, 32)
	rand.Read(netKey)

	_ = os.MkdirAll(filepath.Join(dataDir, "shards"), 0755)

	return &App{
		dataDir:          dataDir,
		signingKey:       signingKey,
		networkKey:       netKey,
		storageAlloc:     5 * 1024 * 1024 * 1024, // Default 5 GB
		files:            make([]FileDTO, 0),
		runtimeEngine:    runtime.NewEngine(),
		metricsCollector: metrics.NewCollector(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) CreateNetwork(trackerAddr string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	netKey := make([]byte, 32)
	if _, err := rand.Read(netKey); err != nil {
		return "", fmt.Errorf("generating network key: %w", err)
	}

	a.networkKey = netKey
	a.trackerAddr = trackerAddr
	a.nodeID = "node-gui-owner"
	a.configured = true

	mgr := invite.NewTokenManager(a.signingKey)
	tokStr, err := mgr.CreateToken(trackerAddr, netKey, 24*time.Hour, 10)
	if err != nil {
		return "", fmt.Errorf("creating invite token: %w", err)
	}

	return tokStr, nil
}

func (a *App) JoinNetwork(inviteTokenStr string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	mgr := invite.NewTokenManager(a.signingKey)
	tok, err := mgr.ValidateAndUse(inviteTokenStr)
	if err != nil {
		return fmt.Errorf("invalid invite token: %w", err)
	}

	a.trackerAddr = tok.TrackerAddr
	a.networkKey = tok.NetworkKey
	a.nodeID = "node-gui-member"
	a.configured = true

	return nil
}

func (a *App) IsPaused() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.paused
}

func (a *App) PauseNode(paused bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.paused = paused
}

func (a *App) GetNodeStatus() NodeStatusDTO {
	a.mu.Lock()
	defer a.mu.Unlock()

	stateLabel := "ONLINE"
	if a.paused {
		stateLabel = "PAUSED"
	}

	var physicalUsed int64
	shardsDir := filepath.Join(a.dataDir, "shards")
	entries, _ := os.ReadDir(shardsDir)
	for _, entry := range entries {
		info, err := entry.Info()
		if err == nil {
			physicalUsed += info.Size()
		}
	}

	return NodeStatusDTO{
		Configured:       a.configured,
		NodeID:           a.nodeID,
		TrackerAddr:      a.trackerAddr,
		StorageAllocated: a.storageAlloc,
		StorageUsed:      physicalUsed,
		PeerCount:        0,
		IsHealthy:        !a.paused,
		StateLabel:       stateLabel,
	}
}

func (a *App) SetStorageAllocation(bytesAllocated int64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if bytesAllocated <= 0 {
		return fmt.Errorf("storage allocation must be greater than 0")
	}

	a.storageAlloc = bytesAllocated
	return nil
}

func (a *App) ListFiles() []FileDTO {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.files
}

func (a *App) UploadFile(fileName string, size int64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if size <= 0 {
		size = 1024 * 1024 // 1 MB default
	}

	fileContent := bytes.Repeat([]byte("A"), int(size))

	// 1. Chunker
	chunks, err := chunker.Split(bytes.NewReader(fileContent), 4*1024*1024)
	if err != nil {
		return fmt.Errorf("chunking file: %w", err)
	}

	// 2. Reed-Solomon Erasure Coding (4 data + 4 parity shards)
	enc, err := erasure.NewEncoder(4, 4)
	if err != nil {
		return fmt.Errorf("erasure encoder: %w", err)
	}

	// 3. Key derivation (Argon2id + Salt)
	salt, err := crypto.NewSalt()
	if err != nil {
		return fmt.Errorf("salt: %w", err)
	}
	derivedKey, err := crypto.DeriveKey(a.networkKey, salt)
	if err != nil {
		return fmt.Errorf("derive key: %w", err)
	}

	fileID := fmt.Sprintf("file-%d", time.Now().UnixNano())
	shardsDir := filepath.Join(a.dataDir, "shards")

	for chunkIndex, chunkData := range chunks {
		shards, err := enc.Encode(chunkData.Data)
		if err != nil {
			return fmt.Errorf("erasure encode chunk %d: %w", chunkIndex, err)
		}

		for shardIdx, shard := range shards {
			encryptedShard, err := crypto.Encrypt(shard.Data, derivedKey)
			if err != nil {
				return fmt.Errorf("encrypting shard: %w", err)
			}

			shardFileName := fmt.Sprintf("%s_c%d_s%d.shard", fileID, chunkIndex, shardIdx)
			shardPath := filepath.Join(shardsDir, shardFileName)
			if err := os.WriteFile(shardPath, encryptedShard, 0644); err != nil {
				return fmt.Errorf("writing physical shard to disk: %w", err)
			}
		}
	}

	a.files = append(a.files, FileDTO{
		FileID:    fileID,
		FileName:  fileName,
		Size:      size,
		CreatedAt: time.Now().Unix(),
	})

	return nil
}

// GetComputeWorkloads connects Wails GUI directly to internal/runtime engine.
func (a *App) GetComputeWorkloads() []ComputeWorkloadDTO {
	a.mu.Lock()
	defer a.mu.Unlock()

	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	containers, err := a.runtimeEngine.ListWorkloads(ctx)
	if err != nil || len(containers) == 0 {
		return make([]ComputeWorkloadDTO, 0)
	}

	res := make([]ComputeWorkloadDTO, 0, len(containers))
	for _, c := range containers {
		res = append(res, ComputeWorkloadDTO{
			WorkloadID: c.WorkloadID,
			Name:       c.Name,
			Image:      c.Image,
			State:      string(c.State),
			IPAddress:  c.IPAddress,
		})
	}
	return res
}

func (a *App) GetNetworkMetrics() NetworkMetricsDTO {
	a.mu.Lock()
	defer a.mu.Unlock()

	snapshot := a.metricsCollector.GetSnapshot()

	peers := make([]PeerMetricDTO, 0, len(snapshot.Peers))
	for nodeID, p := range snapshot.Peers {
		peers = append(peers, PeerMetricDTO{
			NodeID:    nodeID,
			OS:        "linux",
			LatencyMs: p.LatencyMs,
			IsHealthy: p.IsHealthy,
		})
	}

	upSpeed := float64(snapshot.BytesSentTotal) / 1024.0
	downSpeed := float64(snapshot.BytesReceivedTotal) / 1024.0

	return NetworkMetricsDTO{
		UploadSpeedKBps:   upSpeed,
		DownloadSpeedKBps: downSpeed,
		Peers:             peers,
	}
}
