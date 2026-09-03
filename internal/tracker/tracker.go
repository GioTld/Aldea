package tracker

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"go.etcd.io/bbolt"
)

var (
	bucketNodes      = []byte("nodes")
	bucketFiles      = []byte("files")
	bucketPlacements = []byte("placements")
)

var (
	ErrInsufficientNodes = errors.New("insufficient healthy storage nodes available")
	ErrFileNotFound      = errors.New("file placement metadata not found")
)

type NodeMetadata struct {
	NodeID           string `json:"node_id"`
	Address          string `json:"address"`
	StorageAllocated int64  `json:"storage_allocated"`
	StorageUsed      int64  `json:"storage_used"`
	LastSeen         int64  `json:"last_seen"`
	IsHealthy        bool   `json:"is_healthy"`
}

type FileMetadata struct {
	FileID     string `json:"file_id"`
	FileName   string `json:"file_name"`
	Size       int64  `json:"size"`
	ChunkCount int    `json:"chunk_count"`
	CreatedAt  int64  `json:"created_at"`
}

type ShardPlacement struct {
	ShardID    string `json:"shard_id"`
	FileID     string `json:"file_id"`
	ChunkIndex int    `json:"chunk_index"`
	ShardIndex int    `json:"shard_index"`
	NodeID     string `json:"node_id"`
}

type Store struct {
	db *bbolt.DB
}

func NewStore(dbPath string) (*Store, error) {
	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("opening bbolt database: %w", err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		for _, b := range [][]byte{bucketNodes, bucketFiles, bucketPlacements} {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("initializing buckets: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) RegisterNode(node NodeMetadata) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketNodes)
		data, err := json.Marshal(node)
		if err != nil {
			return err
		}
		return b.Put([]byte(node.NodeID), data)
	})
}

func (s *Store) ListNodes() ([]NodeMetadata, error) {
	var nodes []NodeMetadata
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketNodes)
		return b.ForEach(func(k, v []byte) error {
			var node NodeMetadata
			if err := json.Unmarshal(v, &node); err != nil {
				return err
			}
			nodes = append(nodes, node)
			return nil
		})
	})
	return nodes, err
}

func (s *Store) SelectPlacementNodes(count int, requiredSpace int64) ([]string, error) {
	nodes, err := s.ListNodes()
	if err != nil {
		return nil, err
	}

	var eligible []NodeMetadata
	for _, n := range nodes {
		if n.IsHealthy && (n.StorageAllocated-n.StorageUsed) >= requiredSpace {
			eligible = append(eligible, n)
		}
	}

	if len(eligible) < count {
		return nil, ErrInsufficientNodes
	}

	sort.Slice(eligible, func(i, j int) bool {
		availableI := eligible[i].StorageAllocated - eligible[i].StorageUsed
		availableJ := eligible[j].StorageAllocated - eligible[j].StorageUsed
		return availableI > availableJ
	})

	selected := make([]string, count)
	for i := 0; i < count; i++ {
		selected[i] = eligible[i].NodeID
	}

	return selected, nil
}

func (s *Store) SaveFilePlacement(file FileMetadata, placements []ShardPlacement) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		fb := tx.Bucket(bucketFiles)
		fileData, err := json.Marshal(file)
		if err != nil {
			return err
		}
		if err := fb.Put([]byte(file.FileID), fileData); err != nil {
			return err
		}

		pb := tx.Bucket(bucketPlacements)
		placementData, err := json.Marshal(placements)
		if err != nil {
			return err
		}
		return pb.Put([]byte(file.FileID), placementData)
	})
}

func (s *Store) GetFilePlacement(fileID string) (*FileMetadata, []ShardPlacement, error) {
	var file FileMetadata
	var placements []ShardPlacement

	err := s.db.View(func(tx *bbolt.Tx) error {
		fb := tx.Bucket(bucketFiles)
		fileBytes := fb.Get([]byte(fileID))
		if fileBytes == nil {
			return ErrFileNotFound
		}
		if err := json.Unmarshal(fileBytes, &file); err != nil {
			return err
		}

		pb := tx.Bucket(bucketPlacements)
		placementBytes := pb.Get([]byte(fileID))
		if placementBytes != nil {
			if err := json.Unmarshal(placementBytes, &placements); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return &file, placements, nil
}
