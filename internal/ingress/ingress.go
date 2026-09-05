package ingress

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrTunnelNotFound = errors.New("ingress tunnel not found")
	ErrInvalidPort    = errors.New("invalid target port: must be between 1 and 65535")
)

type IngressTunnel struct {
	TunnelID   string    `json:"tunnel_id"`
	WorkloadID string    `json:"workload_id"`
	PublicAddr string    `json:"public_addr"`
	TargetPort int       `json:"target_port"`
	RelayAddr  string    `json:"relay_addr"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

type Manager struct {
	mu      sync.RWMutex
	tunnels map[string]*IngressTunnel
}

func NewManager() *Manager {
	return &Manager{
		tunnels: make(map[string]*IngressTunnel),
	}
}

// CreateTunnel exposes a compute workload running behind NAT through a public relay/IP node (RF-31).
func (m *Manager) CreateTunnel(ctx context.Context, workloadID string, targetPort int, relayAddr string) (*IngressTunnel, error) {
	if targetPort <= 0 || targetPort > 65535 {
		return nil, ErrInvalidPort
	}

	tunnelID := fmt.Sprintf("tun-%s-%d", workloadID, time.Now().Unix())
	publicAddr := fmt.Sprintf("https://%s.tunnel.aldea.net", workloadID)

	tunnel := &IngressTunnel{
		TunnelID:   tunnelID,
		WorkloadID: workloadID,
		PublicAddr: publicAddr,
		TargetPort: targetPort,
		RelayAddr:  relayAddr,
		IsActive:   true,
		CreatedAt:  time.Now(),
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.tunnels[tunnelID] = tunnel

	return tunnel, nil
}

func (m *Manager) CloseTunnel(ctx context.Context, tunnelID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tunnel, ok := m.tunnels[tunnelID]
	if !ok {
		return ErrTunnelNotFound
	}

	tunnel.IsActive = false
	delete(m.tunnels, tunnelID)
	return nil
}

func (m *Manager) ListTunnels(ctx context.Context) ([]IngressTunnel, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]IngressTunnel, 0, len(m.tunnels))
	for _, t := range m.tunnels {
		result = append(result, *t)
	}
	return result, nil
}
