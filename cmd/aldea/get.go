package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/GioTld/aldea/internal/chunker"
	"github.com/GioTld/aldea/internal/config"
	"github.com/GioTld/aldea/internal/crypto"
	"github.com/GioTld/aldea/internal/erasure"
	"github.com/GioTld/aldea/internal/protocol"
	"github.com/GioTld/aldea/internal/tracker"
)

var cmdGet = &cobra.Command{
	Use:   "get <fileID> <output>",
	Short: "Download a file from the Aldea network",
	Args:  cobra.ExactArgs(2),
	RunE:  runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	fileID, outputPath := args[0], args[1]

	cfg, err := config.LoadClientConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	salt, err := cfg.DecodedSalt()
	if err != nil {
		return err
	}

	key, err := crypto.DeriveKey([]byte(cfg.NetworkKey), salt)
	if err != nil {
		return fmt.Errorf("deriving encryption key: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	file, placements, err := getFilePlacement(client, cfg, fileID)
	if err != nil {
		return fmt.Errorf("fetching file placement: %w", err)
	}

	nodes, err := listNodes(client, cfg)
	if err != nil {
		return fmt.Errorf("fetching nodes: %w", err)
	}

	nodeByID := make(map[string]tracker.NodeMetadata, len(nodes))
	for _, n := range nodes {
		nodeByID[n.NodeID] = n
	}

	byChunk := make(map[int][]tracker.ShardPlacement)
	for _, p := range placements {
		byChunk[p.ChunkIndex] = append(byChunk[p.ChunkIndex], p)
	}

	chunkIndices := make([]int, 0, len(byChunk))
	for k := range byChunk {
		chunkIndices = append(chunkIndices, k)
	}
	sort.Ints(chunkIndices)

	enc, err := erasure.NewEncoder(erasure.DefaultDataShards, erasure.DefaultParityShards)
	if err != nil {
		return fmt.Errorf("creating erasure encoder: %w", err)
	}

	totalShards := erasure.DefaultDataShards + erasure.DefaultParityShards
	var chunks []chunker.Chunk

	for _, ci := range chunkIndices {
		group := byChunk[ci]
		rawShards := make([]erasure.Shard, totalShards)
		for i := range rawShards {
			rawShards[i] = erasure.Shard{Index: i}
		}

		var originalChunkSize int64
		for _, p := range group {
			if p.ShardIndex >= totalShards {
				continue
			}
			originalChunkSize = p.OriginalChunkSize

			node, ok := nodeByID[p.NodeID]
			if !ok {
				continue
			}

			data, err := getShard(client, cfg, node, p.ShardID)
			if err != nil || len(data) == 0 {
				continue
			}

			plaintext, err := crypto.Decrypt(data, key)
			if err != nil {
				continue
			}

			rawShards[p.ShardIndex].Data = plaintext
		}

		reconstructed, err := enc.Reconstruct(rawShards, int(originalChunkSize))
		if err != nil {
			return fmt.Errorf("reconstructing chunk %d: %w", ci, err)
		}

		hash := sha256.Sum256(reconstructed)
		chunks = append(chunks, chunker.Chunk{
			Index: ci,
			Data:  reconstructed,
			Hash:  hash,
		})

		fmt.Printf("chunk %d/%d reconstructed\n", ci+1, file.ChunkCount)
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer out.Close()

	if err := chunker.Reassemble(chunks, out); err != nil {
		return fmt.Errorf("reassembling file: %w", err)
	}

	fmt.Printf("downloaded: %s → %s\n", fileID, outputPath)
	return nil
}

func getFilePlacement(client *http.Client, cfg *config.ClientConfig, fileID string) (*tracker.FileMetadata, []tracker.ShardPlacement, error) {
	req, _ := http.NewRequest("GET", cfg.TrackerAddr+"/files/"+fileID, nil)
	req.Header.Set("X-Aldea-Key", cfg.NetworkKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("tracker returned %d: %s", resp.StatusCode, b)
	}

	var result struct {
		File       tracker.FileMetadata     `json:"file"`
		Placements []tracker.ShardPlacement `json:"placements"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, nil, err
	}

	return &result.File, result.Placements, nil
}

func listNodes(client *http.Client, cfg *config.ClientConfig) ([]tracker.NodeMetadata, error) {
	req, _ := http.NewRequest("GET", cfg.TrackerAddr+"/nodes", nil)
	req.Header.Set("X-Aldea-Key", cfg.NetworkKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var nodes []tracker.NodeMetadata
	return nodes, json.NewDecoder(resp.Body).Decode(&nodes)
}

func getShard(client *http.Client, cfg *config.ClientConfig, node tracker.NodeMetadata, shardID string) ([]byte, error) {
	encoded, err := protocol.Encode(protocol.TypeGetShardRequest, "aldea-cli", protocol.GetShardRequest{
		ShardID: shardID,
	}, []byte(cfg.NetworkKey))
	if err != nil {
		return nil, err
	}

	resp, err := client.Post("http://"+node.Address+"/shards/get", "application/octet-stream", bytes.NewReader(encoded))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	env, err := protocol.Decode(body, []byte(cfg.NetworkKey))
	if err != nil {
		return nil, err
	}

	var shardResp protocol.GetShardResponse
	if err := env.UnmarshalBody(&shardResp); err != nil {
		return nil, err
	}

	if !shardResp.Found {
		return nil, nil
	}

	return shardResp.Data, nil
}
