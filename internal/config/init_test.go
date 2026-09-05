package config_test

import (
	"path/filepath"
	"testing"

	"github.com/GioTld/aldea/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigWizardInit(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("InitNodeConfig generates valid YAML file", func(t *testing.T) {
		cfgPath := filepath.Join(tmpDir, "node.yaml")
		cfg, err := config.InitNodeConfig(cfgPath, "node-1", "0.0.0.0:9001", filepath.Join(tmpDir, "data"), "127.0.0.1:9090", 10*1024*1024*1024, "")
		require.NoError(t, err)
		assert.NotEmpty(t, cfg.NetworkKey)
		assert.Equal(t, int64(10*1024*1024*1024), cfg.StorageAllocated)

		// Verify it can be loaded back
		loaded, err := config.LoadNodeConfig(cfgPath)
		require.NoError(t, err)
		assert.Equal(t, cfg.NodeID, loaded.NodeID)
		assert.Equal(t, cfg.NetworkKey, loaded.NetworkKey)
	})

	t.Run("InitTrackerConfig generates valid YAML file", func(t *testing.T) {
		cfgPath := filepath.Join(tmpDir, "tracker.yaml")
		dbPath := filepath.Join(tmpDir, "tracker.db")
		cfg, err := config.InitTrackerConfig(cfgPath, "0.0.0.0:9090", dbPath, "")
		require.NoError(t, err)
		assert.NotEmpty(t, cfg.NetworkKey)

		loaded, err := config.LoadTrackerConfig(cfgPath)
		require.NoError(t, err)
		assert.Equal(t, cfg.ListenAddr, loaded.ListenAddr)
	})

	t.Run("InitClientConfig generates valid YAML file", func(t *testing.T) {
		cfgPath := filepath.Join(tmpDir, "client.yaml")
		cfg, err := config.InitClientConfig(cfgPath, "127.0.0.1:9090", "testkey", "")
		require.NoError(t, err)
		assert.NotEmpty(t, cfg.Salt)

		loaded, err := config.LoadClientConfig(cfgPath)
		require.NoError(t, err)
		assert.Equal(t, cfg.TrackerAddr, loaded.TrackerAddr)
	})
}
