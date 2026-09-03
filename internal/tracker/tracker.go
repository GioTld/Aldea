package tracker

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	bolt "go.etcd.io/bbolt"
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
	ShardID           string `json:"shard_id"`
	FileID            string `json:"file_id"`
	ChunkIndex        int    `json:"chunk_index"`
	ShardIndex        int    `json:"shard_index"`
	NodeID            string `json:"node_id"`
	OriginalChunkSize int64  `json:"original_chunk_size"`
}

type Store struct {
	db *bolt.DB
}

func NewStore(dbPath string) (*Store, error) {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, fmt.Errorf("opening bbolt database: %w", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
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
	return s.db.Update(func(tx *bolt.Tx) error {
		data, err := json.Marshal(node)
		if err != nil {
			return err
		}
		return tx.Bucket(bucketNodes).Put([]byte(node.NodeID), data)
	})
}

func (s *Store) ListNodes() ([]NodeMetadata, error) {
	var nodes []NodeMetadata
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketNodes).ForEach(func(_, v []byte) error {
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

func (s *Store) SelectPlacementNodes(count int, requiredSpace int64) ([]NodeMetadata, error) {
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
		return (eligible[i].StorageAllocated - eligible[i].StorageUsed) >
			(eligible[j].StorageAllocated - eligible[j].StorageUsed)
	})

	return eligible[:count], nil
}

func (s *Store) SaveFilePlacement(file FileMetadata, placements []ShardPlacement) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		fileData, err := json.Marshal(file)
		if err != nil {
			return err
		}
		if err := tx.Bucket(bucketFiles).Put([]byte(file.FileID), fileData); err != nil {
			return err
		}

		placementData, err := json.Marshal(placements)
		if err != nil {
			return err
		}
		return tx.Bucket(bucketPlacements).Put([]byte(file.FileID), placementData)
	})
}

func (s *Store) ListFiles() ([]FileMetadata, error) {
	var files []FileMetadata
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketFiles).ForEach(func(_, v []byte) error {
			var file FileMetadata
			if err := json.Unmarshal(v, &file); err != nil {
				return err
			}
			files = append(files, file)
			return nil
		})
	})
	return files, err
}

func (s *Store) GetFilePlacement(fileID string) (*FileMetadata, []ShardPlacement, error) {
	var file FileMetadata
	var placements []ShardPlacement

	err := s.db.View(func(tx *bolt.Tx) error {
		fileBytes := tx.Bucket(bucketFiles).Get([]byte(fileID))
		if fileBytes == nil {
			return ErrFileNotFound
		}
		if err := json.Unmarshal(fileBytes, &file); err != nil {
			return err
		}

		placementBytes := tx.Bucket(bucketPlacements).Get([]byte(fileID))
		if placementBytes != nil {
			return json.Unmarshal(placementBytes, &placements)
		}
		return nil
	})

	if err != nil {
		return nil, nil, err
	}
	return &file, placements, nil
}
