package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the LABRAT configuration
type Config struct {
	Hub      HubConfig `yaml:"hub"`
	Defaults Defaults  `yaml:"defaults"`
	Verbose  bool      `yaml:"verbose"`
}

// HubConfig contains configuration for the ACM Hub cluster
type HubConfig struct {
	Kubeconfig string `yaml:"kubeconfig"`
	Context    string `yaml:"context"`
	Namespace  string `yaml:"namespace"`
}

// Defaults contains default configurations for resources
type Defaults struct {
	Spoke SpokeDefaults `yaml:"spoke"`
}

// SpokeDefaults contains default configuration for spoke clusters
type SpokeDefaults struct {
	Provider string `yaml:"provider"`
	Region   string `yaml:"region"`
}

// Load reads and parses the configuration file from the given path
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Expand paths after unmarshaling
	cfg.expandPaths()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Hub.Kubeconfig == "" {
		return fmt.Errorf("validation failed: hub kubeconfig is required")
	}

	if c.Hub.Namespace == "" {
		return fmt.Errorf("validation failed: hub namespace is required")
	}

	return nil
}

// GetHubKubeconfig returns the path to the hub kubeconfig
func (c *Config) GetHubKubeconfig() string {
	return c.Hub.Kubeconfig
}

// NewDefaultConfig creates a new configuration with default values
func NewDefaultConfig() *Config {
	return &Config{
		Hub: HubConfig{
			Namespace: "open-cluster-management",
		},
		Verbose: false,
	}
}

// expandPaths expands environment variables and ~ in path fields
func (c *Config) expandPaths() {
	c.Hub.Kubeconfig = ExpandPath(c.Hub.Kubeconfig)
}

// ExpandPath expands environment variables and ~ in a single path
// This is exported so it can be used for config file paths as well
func ExpandPath(path string) string {
	if path == "" {
		return path
	}

	// Expand environment variables
	path = os.ExpandEnv(path)

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	} else if path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			path = home
		}
	}

	return path
}
