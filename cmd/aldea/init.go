package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var initTrackerAddr string
var initNetworkKey string

var cmdInit = &cobra.Command{
	Use:   "init",
	Short: "Initialize a client config file with network credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		salt := make([]byte, 32)
		if _, err := rand.Read(salt); err != nil {
			return fmt.Errorf("generating salt: %w", err)
		}

		cfg := map[string]string{
			"tracker_addr": initTrackerAddr,
			"network_key":  initNetworkKey,
			"salt":         base64.StdEncoding.EncodeToString(salt),
		}

		data, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("marshaling config: %w", err)
		}

		if err := os.WriteFile(cfgPath, data, 0600); err != nil {
			return fmt.Errorf("writing config file: %w", err)
		}

		fmt.Printf("config written to %s\n", cfgPath)
		return nil
	},
}

func init() {
	cmdInit.Flags().StringVar(&initTrackerAddr, "tracker", "http://localhost:8080", "tracker address")
	cmdInit.Flags().StringVar(&initNetworkKey, "key", "", "network key (shared secret)")
	_ = cmdInit.MarkFlagRequired("key")
}
