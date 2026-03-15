package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type NodeConfig struct {
	Host       string   `yaml:"host"`
	Interfaces []string `yaml:"interfaces"`
	Loopback   string   `yaml:"loopback"`
}

type Config struct {
	SSHUser    string                `yaml:"ssh_user"`
	SSHKey     string                `yaml:"ssh_key"`
	ResultsDir string                `yaml:"results_dir"`
	Nodes      map[string]NodeConfig `yaml:"nodes"`
}

// LoadConfig loads the config from the given explicit path (may be empty),
// falling back to ./chaosmonkey.yaml and ~/.config/chaosmonkey/chaosmonkey.yaml.
func LoadConfig(explicitPath string) (*Config, error) {
	paths := buildSearchPaths(explicitPath)

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read %s: %w", p, err)
		}
		var cfg Config
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse %s: %w", p, err)
		}
		if cfg.ResultsDir == "" {
			cfg.ResultsDir = "./results"
		}
		cfg.SSHKey = expandHome(cfg.SSHKey)
		cfg.ResultsDir = expandHome(cfg.ResultsDir)
		return &cfg, nil
	}

	tried := make([]string, 0, len(paths))
	for _, p := range paths {
		tried = append(tried, "  "+p)
	}
	return nil, fmt.Errorf("config file not found; tried:\n%s", strings.Join(tried, "\n"))
}

func buildSearchPaths(explicit string) []string {
	if explicit != "" {
		return []string{explicit}
	}
	paths := []string{"./chaosmonkey.yaml"}
	home, err := os.UserHomeDir()
	if err == nil {
		paths = append(paths, filepath.Join(home, ".config", "chaosmonkey", "chaosmonkey.yaml"))
	}
	return paths
}

func expandHome(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[2:])
}
