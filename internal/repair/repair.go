package repair

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/GioTld/aldea/internal/crypto"
	"github.com/GioTld/aldea/internal/erasure"
	"github.com/GioTld/aldea/internal/protocol"
	"github.com/GioTld/aldea/internal/tracker"
)

type Repairer struct {
	store      *tracker.Store
	client     *http.Client
	networkKey []byte
	cryptoKey  []byte
}

func NewRepairer(store *tracker.Store, client *http.Client, networkKey, cryptoKey []byte) *Repairer {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &Repairer{
		store:      store,
		client:     client,
		networkKey: networkKey,
		cryptoKey:  cryptoKey,
	}
}

func (r *Repairer) InspectAndRepair(timeoutSeconds int64) (int, error) {
	nodes, err := r.store.ListNodes()
	if err != nil {
		return 0, fmt.Errorf("listing nodes for repair: %w", err)
	}

	now := time.Now().Unix()
	unhealthyNodes := make(map[string]bool)
	healthyNodes := make(map[string]tracker.NodeMetadata)

	for _, n := range nodes {
		if !n.IsHealthy || (now-n.LastSeen) > timeoutSeconds {
			unhealthyNodes[n.NodeID] = true
			n.IsHealthy = false
			_ = r.store.RegisterNode(n)
		} else {
			healthyNodes[n.NodeID] = n
		}
	}

	if len(unhealthyNodes) == 0 {
		return 0, nil
	}

	enc, err := erasure.NewEncoder(erasure.DefaultDataShards, erasure.DefaultParityShards)
	if err != nil {
		return 0, fmt.Errorf("creating erasure encoder: %w", err)
	}
	totalShards := erasure.DefaultDataShards + erasure.DefaultParityShards

	repairedCount := 0

	err = r.iterateFiles(func(file tracker.FileMetadata, placements []tracker.ShardPlacement) error {
		byChunk := make(map[int][]tracker.ShardPlacement)
		for _, p := range placements {
			byChunk[p.ChunkIndex] = append(byChunk[p.ChunkIndex], p)
		}

		updatedPlacements := make([]tracker.ShardPlacement, len(placements))
		copy(updatedPlacements, placements)
		fileNeedsUpdate := false

		for ci, chunkPlacements := range byChunk {
			missingInChunk := false
			for _, p := range chunkPlacements {
				if unhealthyNodes[p.NodeID] {
					missingInChunk = true
					break
				}
			}
			if !missingInChunk {
				continue
			}

			rawShards := make([]erasure.Shard, totalShards)
			for i := range rawShards {
				rawShards[i] = erasure.Shard{Index: i}
			}

			var originalChunkSize int64
			for _, p := range chunkPlacements {
				originalChunkSize = p.OriginalChunkSize
				if unhealthyNodes[p.NodeID] {
					continue
				}
				node, ok := healthyNodes[p.NodeID]
				if !ok {
					continue
				}

				ctData, err := r.fetchShard(node, p.ShardID)
				if err != nil || len(ctData) == 0 {
					continue
				}

				plain, err := crypto.Decrypt(ctData, r.cryptoKey)
				if err != nil {
					continue
				}
				rawShards[p.ShardIndex].Data = plain
			}

			reconstructedChunk, err := enc.Reconstruct(rawShards, int(originalChunkSize))
			if err != nil {
				return fmt.Errorf("reconstructing chunk %d of file %s: %w", ci, file.FileID, err)
			}

			freshShards, err := enc.Encode(reconstructedChunk)
			if err != nil {
				return fmt.Errorf("re-encoding chunk %d: %w", ci, err)
			}

			for idx, p := range updatedPlacements {
				if p.ChunkIndex != ci || !unhealthyNodes[p.NodeID] {
					continue
				}

				newNodes, err := r.store.SelectPlacementNodes(1, 0)
				if err != nil {
					return fmt.Errorf("selecting new node for repair: %w", err)
				}

				newNode := newNodes[0]
				reconstructedShardData := freshShards[p.ShardIndex].Data

				encryptedShard, err := crypto.Encrypt(reconstructedShardData, r.cryptoKey)
				if err != nil {
					return fmt.Errorf("encrypting reconstructed shard: %w", err)
				}

				if err := r.uploadShard(newNode, p.ShardID, encryptedShard); err != nil {
					return fmt.Errorf("uploading repaired shard %s to node %s: %w", p.ShardID, newNode.NodeID, err)
				}

				updatedPlacements[idx].NodeID = newNode.NodeID
				fileNeedsUpdate = true
				repairedCount++
			}
		}

		if fileNeedsUpdate {
			if err := r.store.SaveFilePlacement(file, updatedPlacements); err != nil {
				return fmt.Errorf("saving repaired placement for file %s: %w", file.FileID, err)
			}
		}

		return nil
	})

	return repairedCount, err
}

func (r *Repairer) iterateFiles(fn func(file tracker.FileMetadata, placements []tracker.ShardPlacement) error) error {
	files, err := r.store.ListFiles()
	if err != nil {
		return err
	}

	for _, f := range files {
		file, placements, err := r.store.GetFilePlacement(f.FileID)
		if err != nil {
			continue
		}
		if err := fn(*file, placements); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repairer) fetchShard(node tracker.NodeMetadata, shardID string) ([]byte, error) {
	encoded, err := protocol.Encode(protocol.TypeGetShardRequest, "repairer", protocol.GetShardRequest{
		ShardID: shardID,
	}, r.networkKey)
	if err != nil {
		return nil, err
	}

	url := "http://" + node.Address + "/shards/get"
	resp, err := r.client.Post(url, "application/octet-stream", bytes.NewReader(encoded))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("node %s returned status %d", node.NodeID, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	env, err := protocol.Decode(body, r.networkKey)
	if err != nil {
		return nil, err
	}

	var shardResp protocol.GetShardResponse
	if err := env.UnmarshalBody(&shardResp); err != nil {
		return nil, err
	}

	if !shardResp.Found {
		return nil, fmt.Errorf("shard %s not found on node %s", shardID, node.NodeID)
	}

	return shardResp.Data, nil
}

func (r *Repairer) uploadShard(node tracker.NodeMetadata, shardID string, data []byte) error {
	encoded, err := protocol.Encode(protocol.TypePutShardRequest, "repairer", protocol.PutShardRequest{
		ShardID: shardID,
		Data:    data,
	}, r.networkKey)
	if err != nil {
		return err
	}

	url := "http://" + node.Address + "/shards"
	resp, err := r.client.Post(url, "application/octet-stream", bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("node %s returned status %d: %s", node.NodeID, resp.StatusCode, body)
	}

	return nil
}
