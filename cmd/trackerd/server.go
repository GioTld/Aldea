package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/GioTld/aldea/internal/config"
	"github.com/GioTld/aldea/internal/tracker"
)

type trackerServer struct {
	cfg   *config.TrackerConfig
	store *tracker.Store
}

func newTrackerServer(cfg *config.TrackerConfig) (*trackerServer, error) {
	store, err := tracker.NewStore(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("opening tracker store: %w", err)
	}
	return &trackerServer{cfg: cfg, store: store}, nil
}

func (s *trackerServer) close() error {
	return s.store.Close()
}

func (s *trackerServer) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /nodes", s.handleRegisterNode)
	mux.HandleFunc("GET /nodes", s.handleListNodes)
	mux.HandleFunc("POST /placement", s.handleSelectPlacement)
	mux.HandleFunc("POST /files", s.handleSaveFilePlacement)
	mux.HandleFunc("GET /files/{id}", s.handleGetFilePlacement)
	return mux
}

func (s *trackerServer) authenticated(r *http.Request) bool {
	return r.Header.Get("X-Aldea-Key") == s.cfg.NetworkKey
}

func (s *trackerServer) handleRegisterNode(w http.ResponseWriter, r *http.Request) {
	if !s.authenticated(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	var node tracker.NodeMetadata
	if err := json.Unmarshal(body, &node); err != nil {
		http.Error(w, "malformed node metadata", http.StatusBadRequest)
		return
	}

	if err := s.store.RegisterNode(node); err != nil {
		http.Error(w, "failed to register node", http.StatusInternalServerError)
		return
	}

	slog.Info("node registered", "node_id", node.NodeID, "addr", node.Address)
	w.WriteHeader(http.StatusOK)
}

func (s *trackerServer) handleListNodes(w http.ResponseWriter, r *http.Request) {
	if !s.authenticated(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	nodes, err := s.store.ListNodes()
	if err != nil {
		http.Error(w, "failed to list nodes", http.StatusInternalServerError)
		return
	}

	writeJSON(w, nodes)
}

func (s *trackerServer) handleSelectPlacement(w http.ResponseWriter, r *http.Request) {
	if !s.authenticated(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Count         int   `json:"count"`
		RequiredSpace int64 `json:"required_space"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "malformed request", http.StatusBadRequest)
		return
	}

	nodes, err := s.store.SelectPlacementNodes(req.Count, req.RequiredSpace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	writeJSON(w, nodes)
}

func (s *trackerServer) handleSaveFilePlacement(w http.ResponseWriter, r *http.Request) {
	if !s.authenticated(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		File       tracker.FileMetadata    `json:"file"`
		Placements []tracker.ShardPlacement `json:"placements"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "malformed request", http.StatusBadRequest)
		return
	}

	if err := s.store.SaveFilePlacement(req.File, req.Placements); err != nil {
		http.Error(w, "failed to save placement", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *trackerServer) handleGetFilePlacement(w http.ResponseWriter, r *http.Request) {
	if !s.authenticated(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	fileID := r.PathValue("id")
	file, placements, err := s.store.GetFilePlacement(fileID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, map[string]any{
		"file":       file,
		"placements": placements,
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
