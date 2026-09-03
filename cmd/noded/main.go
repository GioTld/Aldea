package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GioTld/aldea/internal/config"
	"github.com/GioTld/aldea/internal/tracker"
)

func main() {
	cfgPath := flag.String("config", "node.yaml", "path to node config file")
	flag.Parse()

	cfg, err := config.LoadNodeConfig(*cfgPath)
	if err != nil {
		slog.Error("loading config", "err", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		slog.Error("creating data directory", "err", err)
		os.Exit(1)
	}

	srv, err := newShardServer(cfg)
	if err != nil {
		slog.Error("initializing shard server", "err", err)
		os.Exit(1)
	}
	defer srv.close()

	if cfg.TrackerAddr != "" {
		registerWithTracker(cfg)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	httpSrv := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: srv.routes(),
	}

	go func() {
		slog.Info("noded starting", "addr", cfg.ListenAddr, "node_id", cfg.NodeID)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server stopped unexpectedly", "err", err)
		}
	}()

	<-stop
	slog.Info("shutting down", "node_id", cfg.NodeID)
}

func registerWithTracker(cfg *config.NodeConfig) {
	node := tracker.NodeMetadata{
		NodeID:           cfg.NodeID,
		Address:          cfg.ListenAddr,
		StorageAllocated: cfg.StorageAllocated,
		StorageUsed:      0,
		LastSeen:         time.Now().Unix(),
		IsHealthy:        true,
	}

	data, err := json.Marshal(node)
	if err != nil {
		slog.Warn("could not serialize node registration", "err", err)
		return
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/nodes", cfg.TrackerAddr), bytes.NewReader(data))
	if err != nil {
		slog.Warn("could not create registration request", "err", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Aldea-Key", cfg.NetworkKey)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		slog.Warn("could not reach tracker for registration", "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("tracker registration returned unexpected status", "status", resp.StatusCode)
		return
	}

	slog.Info("registered with tracker", "tracker", cfg.TrackerAddr)
}
