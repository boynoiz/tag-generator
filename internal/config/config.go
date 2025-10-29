package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	ConfigDir  = ".release"
	ConfigFile = "config.yaml"
)

// Config represents the release tool configuration
type Config struct {
	UsePrefix     bool   `yaml:"use_prefix"`      // Use prefix for release tags
	Prefix        string `yaml:"prefix"`          // Prefix for release branch tags (e.g., "v")
	DevPrefix     string `yaml:"dev_prefix"`      // Prefix for non-release branch tags (e.g., "dev-")
	ReleaseBranch string `yaml:"release_branch"`  // Branch for CalVer tags, others get hash tags
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		UsePrefix:     true,
		Prefix:        "v",
		DevPrefix:     "dev-",
		ReleaseBranch: "main",
	}
}

// Load reads config from .release/config.yaml
// If config doesn't exist, returns default config
func Load() (*Config, error) {
	configPath := filepath.Join(ConfigDir, ConfigFile)

	// If config doesn't exist, return default
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return Default(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.UsePrefix && c.Prefix == c.DevPrefix {
		return fmt.Errorf("prefix and dev_prefix must be different (both are '%s'), otherwise you can't distinguish release and dev tags in registries like Harbor", c.Prefix)
	}
	return nil
}

// Init creates .release/config.yaml with default settings
func Init() error {
	// Create .release directory if it doesn't exist
	if err := os.MkdirAll(ConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(ConfigDir, ConfigFile)

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	// Generate default config
	cfg := Default()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write config file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
