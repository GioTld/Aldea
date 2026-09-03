package repair_test

import (
	"bytes"
	"crypto/sha256"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/GioTld/aldea/internal/crypto"
	"github.com/GioTld/aldea/internal/erasure"
	"github.com/GioTld/aldea/internal/protocol"
	"github.com/GioTld/aldea/internal/repair"
	"github.com/GioTld/aldea/internal/tracker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepairManager(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tracker.db")

	store, err := tracker.NewStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	netKey := []byte("secret-network-key-32bytes-long!")
	key, err := crypto.DeriveKey(netKey, []byte("salt-1234567890123456"))
	require.NoError(t, err)

	originalData := []byte("test payload for repair manager verification")
	enc, err := erasure.NewEncoder(erasure.DefaultDataShards, erasure.DefaultParityShards)
	require.NoError(t, err)

	shards, err := enc.Encode(originalData)
	require.NoError(t, err)
	totalShards := len(shards)

	nodesStore := make(map[string]map[string][]byte)
	nodeServers := make(map[string]*httptest.Server)

	for i := 0; i < totalShards; i++ {
		nodeID := fmtNodeID(i)
		nodesStore[nodeID] = make(map[string][]byte)

		srvNodeID := nodeID
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/shards/get" {
				body, _ := io.ReadAll(r.Body)
				env, _ := protocol.Decode(body, netKey)
				var req protocol.GetShardRequest
				_ = env.UnmarshalBody(&req)

				data := nodesStore[srvNodeID][req.ShardID]
				resp := protocol.GetShardResponse{
					ShardID: req.ShardID,
					Data:    data,
					Found:   data != nil,
				}
				encoded, _ := protocol.Encode(protocol.TypeGetShardResponse, srvNodeID, resp, netKey)
				w.Write(encoded)
				return
			}
			if r.URL.Path == "/shards" {
				body, _ := io.ReadAll(r.Body)
				env, _ := protocol.Decode(body, netKey)
				var req protocol.PutShardRequest
				_ = env.UnmarshalBody(&req)

				nodesStore[srvNodeID][req.ShardID] = req.Data
				resp := protocol.PutShardResponse{
					ShardID: req.ShardID,
					Success: true,
				}
				encoded, _ := protocol.Encode(protocol.TypePutShardResponse, srvNodeID, resp, netKey)
				w.Write(encoded)
				return
			}
		}))
		defer ts.Close()
		nodeServers[nodeID] = ts

		isHealthy := i != 2
		lastSeen := time.Now().Unix()
		if !isHealthy {
			lastSeen = time.Now().Unix() - 100
		}

		require.NoError(t, store.RegisterNode(tracker.NodeMetadata{
			NodeID:           nodeID,
			Address:          ts.Listener.Addr().String(),
			StorageAllocated: 10000000,
			StorageUsed:      100,
			LastSeen:         lastSeen,
			IsHealthy:        isHealthy,
		}))
	}

	spareNodeID := "node-spare"
	nodesStore[spareNodeID] = make(map[string][]byte)
	spareTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/shards" {
			body, _ := io.ReadAll(r.Body)
			env, _ := protocol.Decode(body, netKey)
			var req protocol.PutShardRequest
			_ = env.UnmarshalBody(&req)

			nodesStore[spareNodeID][req.ShardID] = req.Data
			resp := protocol.PutShardResponse{
				ShardID: req.ShardID,
				Success: true,
			}
			encoded, _ := protocol.Encode(protocol.TypePutShardResponse, spareNodeID, resp, netKey)
			w.Write(encoded)
		}
	}))
	defer spareTs.Close()

	require.NoError(t, store.RegisterNode(tracker.NodeMetadata{
		NodeID:           spareNodeID,
		Address:          spareTs.Listener.Addr().String(),
		StorageAllocated: 10000000,
		StorageUsed:      0,
		LastSeen:         time.Now().Unix(),
		IsHealthy:        true,
	}))

	var placements []tracker.ShardPlacement
	for i, s := range shards {
		nodeID := fmtNodeID(i)
		shardID := "file-1-c0-s" + string(rune('0'+i))
		ct, err := crypto.Encrypt(s.Data, key)
		require.NoError(t, err)

		nodesStore[nodeID][shardID] = ct
		placements = append(placements, tracker.ShardPlacement{
			ShardID:           shardID,
			FileID:            "file-1",
			ChunkIndex:        0,
			ShardIndex:        i,
			NodeID:            nodeID,
			OriginalChunkSize: int64(len(originalData)),
		})
	}

	fileMeta := tracker.FileMetadata{
		FileID:     "file-1",
		FileName:   "test.txt",
		Size:       int64(len(originalData)),
		ChunkCount: 1,
		CreatedAt:  time.Now().Unix(),
	}
	require.NoError(t, store.SaveFilePlacement(fileMeta, placements))

	rep := repair.NewRepairer(store, http.DefaultClient, netKey, key)
	repairedCount, err := rep.InspectAndRepair(30)
	require.NoError(t, err)
	assert.Equal(t, 1, repairedCount)

	_, newPlacements, err := store.GetFilePlacement("file-1")
	require.NoError(t, err)
	assert.Equal(t, spareNodeID, newPlacements[2].NodeID)

	repairedShardData := nodesStore[spareNodeID][newPlacements[2].ShardID]
	require.NotEmpty(t, repairedShardData)

	plain, err := crypto.Decrypt(repairedShardData, key)
	require.NoError(t, err)
	assert.Equal(t, shards[2].Data, plain)
}

func fmtNodeID(i int) string {
	return "node-" + string(rune('0'+i))
}

var _ = sha256.New
var _ = bytes.NewReader
