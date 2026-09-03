package e2e_test

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndToEndStorageFlow(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := filepath.Join(tmpDir, "bin")
	require.NoError(t, os.MkdirAll(binDir, 0755))

	aldeaBin := filepath.Join(binDir, "aldea")
	nodedBin := filepath.Join(binDir, "noded")
	trackerdBin := filepath.Join(binDir, "trackerd")

	cmd := exec.Command("go", "build", "-o", aldeaBin, "../../cmd/aldea")
	require.NoError(t, cmd.Run())
	cmd = exec.Command("go", "build", "-o", nodedBin, "../../cmd/noded")
	require.NoError(t, cmd.Run())
	cmd = exec.Command("go", "build", "-o", trackerdBin, "../../cmd/trackerd")
	require.NoError(t, cmd.Run())

	netKey := "test-network-secret-key-32bytes!"

	trackerYaml := fmt.Sprintf(`
listen_addr: "127.0.0.1:18080"
db_path: "%s/tracker.db"
network_key: "%s"
`, tmpDir, netKey)
	trackerCfgPath := filepath.Join(tmpDir, "tracker.yaml")
	require.NoError(t, os.WriteFile(trackerCfgPath, []byte(trackerYaml), 0644))

	trackerdCmd := exec.Command(trackerdBin, "-config", trackerCfgPath)
	require.NoError(t, trackerdCmd.Start())
	defer func() {
		_ = trackerdCmd.Process.Kill()
	}()

	time.Sleep(200 * time.Millisecond)

	var nodeCmds []*exec.Cmd
	for i := 1; i <= 8; i++ {
		nodeDir := filepath.Join(tmpDir, fmt.Sprintf("node%d", i))
		require.NoError(t, os.MkdirAll(nodeDir, 0755))

		nodeYaml := fmt.Sprintf(`
node_id: "node-%d"
listen_addr: "127.0.0.1:%d"
data_dir: "%s"
storage_allocated: 104857600
network_key: "%s"
tracker_addr: "http://127.0.0.1:18080"
`, i, 19000+i, nodeDir, netKey)
		nodeCfgPath := filepath.Join(nodeDir, "node.yaml")
		require.NoError(t, os.WriteFile(nodeCfgPath, []byte(nodeYaml), 0644))

		nCmd := exec.Command(nodedBin, "-config", nodeCfgPath)
		require.NoError(t, nCmd.Start())
		nodeCmds = append(nodeCmds, nCmd)
	}

	defer func() {
		for _, nc := range nodeCmds {
			_ = nc.Process.Kill()
		}
	}()

	time.Sleep(500 * time.Millisecond)

	clientCfgPath := filepath.Join(tmpDir, "client.yaml")
	initCmd := exec.Command(aldeaBin, "--config", clientCfgPath, "init", "--tracker", "http://127.0.0.1:18080", "--key", netKey)
	require.NoError(t, initCmd.Run())

	sampleData := bytes.Repeat([]byte("Aldea P2P Distributed Storage Network Test Content! "), 2000)
	sampleFilePath := filepath.Join(tmpDir, "sample.txt")
	require.NoError(t, os.WriteFile(sampleFilePath, sampleData, 0644))

	putCmd := exec.Command(aldeaBin, "--config", clientCfgPath, "put", sampleFilePath)
	putOut, err := putCmd.CombinedOutput()
	require.NoError(t, err, string(putOut))

	var fileID string
	lines := strings.Split(string(putOut), "\n")
	for _, l := range lines {
		if strings.HasPrefix(l, "uploaded:") {
			parts := strings.Split(l, " → ")
			if len(parts) == 2 {
				fileID = strings.TrimSpace(parts[1])
			}
		}
	}
	require.NotEmpty(t, fileID, "fileID should not be empty")

	statusCmd := exec.Command(aldeaBin, "--config", clientCfgPath, "status")
	statusOut, err := statusCmd.CombinedOutput()
	require.NoError(t, err, string(statusOut))
	assert.Contains(t, string(statusOut), "HEALTHY")

	downloadPath := filepath.Join(tmpDir, "downloaded.txt")
	getCmd := exec.Command(aldeaBin, "--config", clientCfgPath, "get", fileID, downloadPath)
	getOut, err := getCmd.CombinedOutput()
	require.NoError(t, err, string(getOut))

	downloadedData, err := os.ReadFile(downloadPath)
	require.NoError(t, err)
	assert.Equal(t, sampleData, downloadedData)
}

func waitForHTTP(t *testing.T, url string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s", url)
}
