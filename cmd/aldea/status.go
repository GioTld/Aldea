package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/GioTld/aldea/internal/config"
)

var cmdStatus = &cobra.Command{
	Use:   "status",
	Short: "Show status of network nodes and storage pool",
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadClientConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	nodes, err := listNodes(client, cfg)
	if err != nil {
		return fmt.Errorf("fetching network nodes: %w", err)
	}

	fmt.Printf("Aldea Network Status (%d nodes registered):\n\n", len(nodes))
	var totalAllocated, totalUsed int64

	for _, n := range nodes {
		statusStr := "HEALTHY"
		if !n.IsHealthy {
			statusStr = "UNHEALTHY"
		}
		fmt.Printf("  • Node: %s [%s]\n", n.NodeID, statusStr)
		fmt.Printf("    Address: %s\n", n.Address)
		fmt.Printf("    Storage: %d / %d bytes used\n\n", n.StorageUsed, n.StorageAllocated)

		totalAllocated += n.StorageAllocated
		totalUsed += n.StorageUsed
	}

	fmt.Printf("Total Storage Pool: %d / %d bytes used\n", totalUsed, totalAllocated)
	return nil
}
