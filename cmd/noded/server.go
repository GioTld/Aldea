package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/GioTld/aldea/internal/config"
	"github.com/GioTld/aldea/internal/protocol"
)

var bucketShards = []byte("shards")

type shardServer struct {
	cfg *config.NodeConfig
	db  *bolt.DB
}

func newShardServer(cfg *config.NodeConfig) (*shardServer, error) {
	db, err := bolt.Open(cfg.DataDir+"/shards.db", 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, fmt.Errorf("opening shard database: %w", err)
	}

	if err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketShards)
		return err
	}); err != nil {
		db.Close()
		return nil, fmt.Errorf("initializing shard bucket: %w", err)
	}

	return &shardServer{cfg: cfg, db: db}, nil
}

func (s *shardServer) close() error {
	return s.db.Close()
}

func (s *shardServer) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /shards", s.handlePut)
	mux.HandleFunc("POST /shards/get", s.handleGet)
	mux.HandleFunc("GET /health", s.handleHealth)
	return mux
}

func (s *shardServer) handlePut(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	env, err := protocol.Decode(body, []byte(s.cfg.NetworkKey))
	if err != nil {
		slog.Debug("shard put authentication failed", "err", err)
		http.Error(w, "authentication failed", http.StatusUnauthorized)
		return
	}

	var req protocol.PutShardRequest
	if err := env.UnmarshalBody(&req); err != nil {
		http.Error(w, "malformed request", http.StatusBadRequest)
		return
	}

	if s.storageUsed()+int64(len(req.Data)) > s.cfg.StorageAllocated {
		http.Error(w, "storage quota exceeded", http.StatusInsufficientStorage)
		return
	}

	if err := s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketShards).Put([]byte(req.ShardID), req.Data)
	}); err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}

	s.writeResponse(w, protocol.TypePutShardResponse, env.NodeID, protocol.PutShardResponse{
		ShardID: req.ShardID,
		Success: true,
	})
}

func (s *shardServer) handleGet(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	env, err := protocol.Decode(body, []byte(s.cfg.NetworkKey))
	if err != nil {
		slog.Debug("shard get authentication failed", "err", err)
		http.Error(w, "authentication failed", http.StatusUnauthorized)
		return
	}

	var req protocol.GetShardRequest
	if err := env.UnmarshalBody(&req); err != nil {
		http.Error(w, "malformed request", http.StatusBadRequest)
		return
	}

	var data []byte
	_ = s.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketShards).Get([]byte(req.ShardID))
		if v != nil {
			data = make([]byte, len(v))
			copy(data, v)
		}
		return nil
	})

	s.writeResponse(w, protocol.TypeGetShardResponse, s.cfg.NodeID, protocol.GetShardResponse{
		ShardID: req.ShardID,
		Data:    data,
		Found:   data != nil,
	})
}

func (s *shardServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := protocol.HealthReportResponse{
		NodeID:           s.cfg.NodeID,
		StorageUsed:      s.storageUsed(),
		StorageAllocated: s.cfg.StorageAllocated,
		IsHealthy:        true,
	}
	data, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s *shardServer) storageUsed() int64 {
	var used int64
	_ = s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketShards).ForEach(func(_, v []byte) error {
			used += int64(len(v))
			return nil
		})
	})
	return used
}

func (s *shardServer) writeResponse(w http.ResponseWriter, msgType protocol.MessageType, nodeID string, body any) {
	encoded, err := protocol.Encode(msgType, nodeID, body, []byte(s.cfg.NetworkKey))
	if err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(encoded)
}
