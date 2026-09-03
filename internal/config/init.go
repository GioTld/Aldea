package config

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func generateRandomKey(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateRandomSalt(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func InitNodeConfig(targetPath, nodeID, listenAddr, dataDir, trackerAddr string, allocBytes int64, networkKey string) (*NodeConfig, error) {
	if networkKey == "" {
		networkKey = generateRandomKey(16) // 32 hex chars
	}
	if nodeID == "" {
		nodeID = "node-" + generateRandomKey(4)
	}
	if dataDir == "" {
		dataDir = "./data"
	}
	if listenAddr == "" {
		listenAddr = "0.0.0.0:9001"
	}
	if allocBytes <= 0 {
		allocBytes = 10 * 1024 * 1024 * 1024 // 10 GB
	}

	cfg := NodeConfig{
		NodeID:           nodeID,
		ListenAddr:       listenAddr,
		DataDir:          dataDir,
		StorageAllocated: allocBytes,
		NetworkKey:       networkKey,
		TrackerAddr:      trackerAddr,
	}

	if err := saveYAML(targetPath, cfg); err != nil {
		return nil, fmt.Errorf("saving node config: %w", err)
	}

	return &cfg, nil
}

func InitTrackerConfig(targetPath, listenAddr, dbPath, networkKey string) (*TrackerConfig, error) {
	if networkKey == "" {
		networkKey = generateRandomKey(16)
	}
	if listenAddr == "" {
		listenAddr = "0.0.0.0:9090"
	}
	if dbPath == "" {
		dbPath = "./tracker.db"
	}

	cfg := TrackerConfig{
		ListenAddr: listenAddr,
		DBPath:     dbPath,
		NetworkKey: networkKey,
	}

	if err := saveYAML(targetPath, cfg); err != nil {
		return nil, fmt.Errorf("saving tracker config: %w", err)
	}

	return &cfg, nil
}

func InitClientConfig(targetPath, trackerAddr, networkKey, salt string) (*ClientConfig, error) {
	if trackerAddr == "" {
		trackerAddr = "127.0.0.1:9090"
	}
	if networkKey == "" {
		networkKey = generateRandomKey(16)
	}
	if salt == "" {
		salt = generateRandomSalt(16)
	}

	cfg := ClientConfig{
		TrackerAddr: trackerAddr,
		NetworkKey:  networkKey,
		Salt:        salt,
	}

	if err := saveYAML(targetPath, cfg); err != nil {
		return nil, fmt.Errorf("saving client config: %w", err)
	}

	return &cfg, nil
}

func saveYAML(path string, v any) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(v)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
