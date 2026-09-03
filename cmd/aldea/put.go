package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/GioTld/aldea/internal/chunker"
	"github.com/GioTld/aldea/internal/config"
	"github.com/GioTld/aldea/internal/crypto"
	"github.com/GioTld/aldea/internal/erasure"
	"github.com/GioTld/aldea/internal/protocol"
	"github.com/GioTld/aldea/internal/tracker"
)

var cmdPut = &cobra.Command{
	Use:   "put <file>",
	Short: "Upload a file to the Aldea network",
	Args:  cobra.ExactArgs(1),
	RunE:  runPut,
}

func runPut(cmd *cobra.Command, args []string) error {
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

	f, err := os.Open(args[0])
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	info, _ := f.Stat()
	chunks, err := chunker.Split(f, 0)
	if err != nil {
		return fmt.Errorf("splitting file: %w", err)
	}

	enc, err := erasure.NewEncoder(erasure.DefaultDataShards, erasure.DefaultParityShards)
	if err != nil {
		return fmt.Errorf("creating erasure encoder: %w", err)
	}

	fileID := randomHex(16)
	client := &http.Client{Timeout: 30 * time.Second}
	var allPlacements []tracker.ShardPlacement

	for _, chunk := range chunks {
		shards, err := enc.Encode(chunk.Data)
		if err != nil {
			return fmt.Errorf("erasure encoding chunk %d: %w", chunk.Index, err)
		}

		encryptedShards := make([][]byte, len(shards))
		for i, s := range shards {
			ct, err := crypto.Encrypt(s.Data, key)
			if err != nil {
				return fmt.Errorf("encrypting shard %d/%d: %w", chunk.Index, i, err)
			}
			encryptedShards[i] = ct
		}

		nodes, err := selectNodes(client, cfg, len(shards))
		if err != nil {
			return fmt.Errorf("selecting placement nodes for chunk %d: %w", chunk.Index, err)
		}

		for i, node := range nodes {
			shardID := fmt.Sprintf("%s-c%d-s%d", fileID, chunk.Index, i)
			if err := putShard(client, cfg, node, shardID, encryptedShards[i]); err != nil {
				return fmt.Errorf("uploading shard %s: %w", shardID, err)
			}
			allPlacements = append(allPlacements, tracker.ShardPlacement{
				ShardID:           shardID,
				FileID:            fileID,
				ChunkIndex:        chunk.Index,
				ShardIndex:        i,
				NodeID:            node.NodeID,
				OriginalChunkSize: int64(len(chunk.Data)),
			})
		}

		fmt.Printf("chunk %d/%d uploaded\n", chunk.Index+1, len(chunks))
	}

	fileMeta := tracker.FileMetadata{
		FileID:     fileID,
		FileName:   info.Name(),
		Size:       info.Size(),
		ChunkCount: len(chunks),
		CreatedAt:  time.Now().Unix(),
	}

	if err := saveFilePlacement(client, cfg, fileMeta, allPlacements); err != nil {
		return fmt.Errorf("saving file placement: %w", err)
	}

	fmt.Printf("uploaded: %s → %s\n", args[0], fileID)
	return nil
}

func selectNodes(client *http.Client, cfg *config.ClientConfig, count int) ([]tracker.NodeMetadata, error) {
	reqBody, _ := json.Marshal(map[string]any{"count": count, "required_space": int64(0)})
	req, _ := http.NewRequest("POST", cfg.TrackerAddr+"/placement", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Aldea-Key", cfg.NetworkKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tracker returned %d: %s", resp.StatusCode, body)
	}

	var nodes []tracker.NodeMetadata
	return nodes, json.NewDecoder(resp.Body).Decode(&nodes)
}

func putShard(client *http.Client, cfg *config.ClientConfig, node tracker.NodeMetadata, shardID string, data []byte) error {
	encoded, err := protocol.Encode(protocol.TypePutShardRequest, "aldea-cli", protocol.PutShardRequest{
		ShardID: shardID,
		Data:    data,
	}, []byte(cfg.NetworkKey))
	if err != nil {
		return err
	}

	resp, err := client.Post("http://"+node.Address+"/shards", "application/octet-stream", bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("node %s returned %d: %s", node.NodeID, resp.StatusCode, body)
	}
	return nil
}

func saveFilePlacement(client *http.Client, cfg *config.ClientConfig, file tracker.FileMetadata, placements []tracker.ShardPlacement) error {
	body, _ := json.Marshal(map[string]any{"file": file, "placements": placements})
	req, _ := http.NewRequest("POST", cfg.TrackerAddr+"/files", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Aldea-Key", cfg.NetworkKey)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tracker returned %d: %s", resp.StatusCode, b)
	}
	return nil
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

var _ = sha256.New // keep import if needed by other commands
