package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// CategoryConfig defines custom patterns and extensions for a category.
type CategoryConfig struct {
	Patterns   []string `yaml:"patterns"`
	Extensions []string `yaml:"extensions"`
}

// Config holds all configuration fields for differ.
type Config struct {
	Include    []string                  `yaml:"include"`
	Exclude    []string                  `yaml:"exclude"`
	Categories map[string]CategoryConfig `yaml:"categories"`
	Empty      string                    `yaml:"empty"`
	Sort       string                    `yaml:"sort"`
}

// defaults returns the built-in default configuration.
func defaults() Config {
	return Config{
		Empty: "exclude",
		Sort:  "churn",
	}
}

// Load reads configuration from the global config file (~/.config/differ/config.yml)
// and the repo-local config file (.differ.yml in repoRoot), then merges them
// with CLI overrides using the precedence:
// cliOverrides > repo config > global config > built-in defaults.
//
// Missing config files are silently skipped. Malformed YAML returns an error.
func Load(repoRoot string, cliOverrides Config) (Config, error) {
	globalPath := ""
	if p, err := globalConfigPath(); err == nil {
		globalPath = p
	}
	return load(globalPath, repoRoot, cliOverrides)
}

// load is the internal implementation that accepts explicit paths for testability.
func load(globalPath, repoRoot string, cliOverrides Config) (Config, error) {
	cfg := defaults()

	// Load global config.
	if globalPath != "" {
		global, err := loadFile(globalPath)
		if err != nil {
			return Config{}, fmt.Errorf("global config %s: %w", globalPath, err)
		}
		if global != nil {
			cfg = merge(cfg, *global)
		}
	}

	// Load repo config.
	if repoRoot != "" {
		repoPath := filepath.Join(repoRoot, ".differ.yml")
		repo, err := loadFile(repoPath)
		if err != nil {
			return Config{}, fmt.Errorf("repo config %s: %w", repoPath, err)
		}
		if repo != nil {
			cfg = merge(cfg, *repo)
		}
	}

	// Apply CLI overrides.
	cfg = merge(cfg, cliOverrides)

	return cfg, nil
}

// globalConfigPath returns the path to ~/.config/differ/config.yml.
func globalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "differ", "config.yml"), nil
}

// loadFile reads and parses a YAML config file. Returns (nil, nil) if the file
// does not exist. Returns an error if the file exists but is malformed.
func loadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("malformed YAML: %w", err)
	}
	return &cfg, nil
}

// merge returns a new Config where non-zero fields in override replace the
// corresponding fields in base.
func merge(base, override Config) Config {
	result := base

	if len(override.Include) > 0 {
		result.Include = override.Include
	}
	if len(override.Exclude) > 0 {
		result.Exclude = override.Exclude
	}
	if override.Empty != "" {
		result.Empty = override.Empty
	}
	if override.Sort != "" {
		result.Sort = override.Sort
	}
	if len(override.Categories) > 0 {
		result.Categories = make(map[string]CategoryConfig, len(override.Categories))
		// Start with base categories if any.
		for k, v := range base.Categories {
			result.Categories[k] = v
		}
		// Override/add from override.
		for k, v := range override.Categories {
			result.Categories[k] = v
		}
	}

	return result
}
