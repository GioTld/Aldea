package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var ErrInvalidConfig = errors.New("invalid configuration: missing required fields")

type NodeConfig struct {
	NodeID           string `yaml:"node_id"`
	ListenAddr       string `yaml:"listen_addr"`
	DataDir          string `yaml:"data_dir"`
	StorageAllocated int64  `yaml:"storage_allocated"`
	NetworkKey       string `yaml:"network_key"`
	TrackerAddr      string `yaml:"tracker_addr"`
}

func (c *NodeConfig) Validate() error {
	if c.NodeID == "" || c.ListenAddr == "" || c.DataDir == "" || c.NetworkKey == "" {
		return ErrInvalidConfig
	}
	return nil
}

type TrackerConfig struct {
	ListenAddr string `yaml:"listen_addr"`
	DBPath     string `yaml:"db_path"`
	NetworkKey string `yaml:"network_key"`
}

func (c *TrackerConfig) Validate() error {
	if c.ListenAddr == "" || c.DBPath == "" || c.NetworkKey == "" {
		return ErrInvalidConfig
	}
	return nil
}

type ClientConfig struct {
	TrackerAddr string `yaml:"tracker_addr"`
	NetworkKey  string `yaml:"network_key"`
	Salt        string `yaml:"salt"`
}

func (c *ClientConfig) Validate() error {
	if c.TrackerAddr == "" || c.NetworkKey == "" || c.Salt == "" {
		return ErrInvalidConfig
	}
	return nil
}

func (c *ClientConfig) DecodedSalt() ([]byte, error) {
	b, err := base64.StdEncoding.DecodeString(c.Salt)
	if err != nil {
		return nil, fmt.Errorf("decoding salt: %w", err)
	}
	return b, nil
}

func LoadNodeConfig(path string) (*NodeConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading node config file: %w", err)
	}
	var cfg NodeConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing node config yaml: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func LoadTrackerConfig(path string) (*TrackerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading tracker config file: %w", err)
	}
	var cfg TrackerConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing tracker config yaml: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func LoadClientConfig(path string) (*ClientConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading client config file: %w", err)
	}
	var cfg ClientConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing client config yaml: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}
