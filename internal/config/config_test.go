package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GioTld/aldea/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadNodeConfig(t *testing.T) {
	t.Run("valid config file", func(t *testing.T) {
		content := `
node_id: "node-1"
listen_addr: "127.0.0.1:9090"
data_dir: "/tmp/aldea-data"
storage_allocated: 10737418240
network_key: "secret-key-32-bytes-length-ok!"
`
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "node.yaml")
		require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

		cfg, err := config.LoadNodeConfig(cfgPath)
		require.NoError(t, err)
		assert.Equal(t, "node-1", cfg.NodeID)
		assert.Equal(t, "127.0.0.1:9090", cfg.ListenAddr)
		assert.Equal(t, int64(10737418240), cfg.StorageAllocated)
	})

	t.Run("missing file returns error", func(t *testing.T) {
		_, err := config.LoadNodeConfig("/path/does/not/exist.yaml")
		assert.Error(t, err)
	})

	t.Run("invalid config missing node_id fails validation", func(t *testing.T) {
		content := `
listen_addr: "127.0.0.1:9090"
data_dir: "/tmp/aldea-data"
network_key: "secret-key-32-bytes-length-ok!"
`
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "invalid.yaml")
		require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

		_, err := config.LoadNodeConfig(cfgPath)
		assert.ErrorIs(t, err, config.ErrInvalidConfig)
	})
}

func TestLoadTrackerConfig(t *testing.T) {
	t.Run("valid tracker config", func(t *testing.T) {
		content := `
listen_addr: "127.0.0.1:8080"
db_path: "/tmp/aldea-tracker.db"
network_key: "secret-key-32-bytes-length-ok!"
`
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "tracker.yaml")
		require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

		cfg, err := config.LoadTrackerConfig(cfgPath)
		require.NoError(t, err)
		assert.Equal(t, "127.0.0.1:8080", cfg.ListenAddr)
		assert.Equal(t, "/tmp/aldea-tracker.db", cfg.DBPath)
	})
}
