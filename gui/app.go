package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/GioTld/aldea/internal/invite"
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

type App struct {
	ctx          context.Context
	dataDir      string
	mu           sync.Mutex
	configured   bool
	paused       bool
	nodeID       string
	trackerAddr  string
	networkKey   []byte
	signingKey   []byte
	storageAlloc int64
	storageUsed  int64
	files        []FileDTO
}

func NewApp(dataDir string) *App {
	signingKey := make([]byte, 32)
	rand.Read(signingKey)

	return &App{
		dataDir:      dataDir,
		signingKey:   signingKey,
		storageAlloc: 5 * 1024 * 1024 * 1024, // Default 5 GB
		files:        make([]FileDTO, 0),
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

	return NodeStatusDTO{
		Configured:       a.configured,
		NodeID:           a.nodeID,
		TrackerAddr:      a.trackerAddr,
		StorageAllocated: a.storageAlloc,
		StorageUsed:      a.storageUsed,
		PeerCount:        4,
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

	fileID := fmt.Sprintf("file-%d", time.Now().UnixNano())
	a.files = append(a.files, FileDTO{
		FileID:    fileID,
		FileName:  fileName,
		Size:      size,
		CreatedAt: time.Now().Unix(),
	})
	a.storageUsed += size / 4 // mock local shard footprint
	return nil
}
