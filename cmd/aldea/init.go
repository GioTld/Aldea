package main

import (
	"fmt"

	"github.com/GioTld/aldea/internal/config"
	"github.com/spf13/cobra"
)

var (
	initRole        string
	initTrackerAddr string
	initNetworkKey  string
	initListenAddr  string
	initDataDir     string
	initAllocGB     int64
	initOutFile     string
)

var cmdInit = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration files for node, tracker, or client",
	Long:  "Generates validated YAML configuration files with secure cryptographic keys for Aldea network roles.",
	RunE: func(cmd *cobra.Command, args []string) error {
		allocBytes := initAllocGB * 1024 * 1024 * 1024

		switch initRole {
		case "node":
			out := initOutFile
			if out == "" {
				out = "node.yaml"
			}
			cfg, err := config.InitNodeConfig(out, "", initListenAddr, initDataDir, initTrackerAddr, allocBytes, initNetworkKey)
			if err != nil {
				return err
			}
			fmt.Printf("[✓] Node configuration written to %s (Node ID: %s)\n", out, cfg.NodeID)

		case "tracker":
			out := initOutFile
			if out == "" {
				out = "tracker.yaml"
			}
			dbPath := "./tracker.db"
			cfg, err := config.InitTrackerConfig(out, initListenAddr, dbPath, initNetworkKey)
			if err != nil {
				return err
			}
			fmt.Printf("[✓] Tracker configuration written to %s (Listen: %s)\n", out, cfg.ListenAddr)

		case "client":
			out := initOutFile
			if out == "" {
				out = cfgPath
			}
			cfg, err := config.InitClientConfig(out, initTrackerAddr, initNetworkKey, "")
			if err != nil {
				return err
			}
			fmt.Printf("[✓] Client configuration written to %s (Tracker: %s)\n", out, cfg.TrackerAddr)

		default:
			return fmt.Errorf("invalid role %q: choose between 'node', 'tracker', or 'client'", initRole)
		}

		return nil
	},
}

func init() {
	cmdInit.Flags().StringVar(&initRole, "role", "client", "role configuration to generate ('node', 'tracker', 'client')")
	cmdInit.Flags().StringVar(&initTrackerAddr, "tracker", "127.0.0.1:9090", "tracker address")
	cmdInit.Flags().StringVar(&initNetworkKey, "key", "", "network shared key (randomly generated if empty)")
	cmdInit.Flags().StringVar(&initListenAddr, "listen", "0.0.0.0:9001", "listen address for node/tracker")
	cmdInit.Flags().StringVar(&initDataDir, "data-dir", "./data", "data directory for node shard storage")
	cmdInit.Flags().StringVar(&initOutFile, "out", "", "output config file path")
	cmdInit.Flags().Int64Var(&initAllocGB, "alloc-gb", 10, "storage allocation limit in GB")
}
