package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeCmd_DeployAndList(t *testing.T) {
	tempDir := t.TempDir()
	manifestPath := filepath.Join(tempDir, "test_manifest.yaml")
	manifestContent := `
workload_id: wl-cli-test-01
name: sample-app
image: nginx:alpine
type: stateless
cpu_cores: 1.0
memory_mb: 256
`
	err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
	require.NoError(t, err)

	cmd := NewComputeCmd()

	t.Run("deploy subcommand with --yes flag executes successfully", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"deploy", manifestPath, "--yes"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Workload deployed successfully")
		assert.Contains(t, buf.String(), "wl-cli-test-01")
	})

	t.Run("list subcommand shows running workloads", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"list"})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "WORKLOAD ID")
	})
}
