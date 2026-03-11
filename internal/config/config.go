package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the top-level application configuration.
type Config struct {
	General      GeneralConfig `yaml:"general"`
	VPNProfiles  []VPNProfile  `yaml:"vpn_profiles"`
	RoutingRules []RoutingRule `yaml:"routing_rules"`
}

// GeneralConfig holds general application settings.
type GeneralConfig struct {
	IPCheckURL       string   `yaml:"ip_check_url"`
	IPCheckInterval  int      `yaml:"ip_check_interval"`
	OpenVPNConfigDirs []string `yaml:"openvpn_config_dirs"`
	OpenVPNMethod    string   `yaml:"openvpn_method"`
	AllowMultiVPN    bool     `yaml:"allow_multi_vpn"`
}

// VPNProfile defines a VPN connection profile.
type VPNProfile struct {
	Name       string `yaml:"name"`
	ConfigPath string `yaml:"config_path,omitempty"`
	Provider   string `yaml:"provider"`
	ExitNode   string `yaml:"exit_node,omitempty"`
}

// RoutingRule defines a policy routing rule.
type RoutingRule struct {
	ID           string   `yaml:"id"`
	Description  string   `yaml:"description"`
	Enabled      bool     `yaml:"enabled"`
	Type         string   `yaml:"type"`
	DestCIDR     string   `yaml:"dest_cidr,omitempty"`
	Domains      []string `yaml:"domains,omitempty"`
	VPNInterface string   `yaml:"vpn_interface"`
	Table        string   `yaml:"table"`
}

// DefaultConfigPath returns the default config file path.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "config.yaml")
	}
	return filepath.Join(home, ".config", "vpn-switcher", "config.yaml")
}

// Default returns a Config with sensible defaults.
func Default() *Config {
	return &Config{
		General: GeneralConfig{
			IPCheckURL:        "https://api.ipify.org",
			IPCheckInterval:   30,
			OpenVPNConfigDirs: []string{"/etc/openvpn/client", "~/.config/openvpn"},
			OpenVPNMethod:     "auto",
			AllowMultiVPN:     false,
		},
	}
}

// Load reads config from disk, applying defaults for missing values.
// If the file does not exist, a default config is returned.
func Load(path string) (*Config, error) {
	path = expandHome(path)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}

	applyDefaults(cfg)
	return cfg, nil
}

// Save writes config to disk, creating parent directories if needed.
func Save(path string, cfg *Config) error {
	path = expandHome(path)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir %s: %w", dir, err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}

	return nil
}

// ExpandConfigDirs expands ~ in all config directory paths.
func ExpandConfigDirs(dirs []string) []string {
	result := make([]string, len(dirs))
	for i, d := range dirs {
		result[i] = expandHome(d)
	}
	return result
}

func applyDefaults(cfg *Config) {
	if cfg.General.IPCheckURL == "" {
		cfg.General.IPCheckURL = "https://api.ipify.org"
	}
	if cfg.General.IPCheckInterval <= 0 {
		cfg.General.IPCheckInterval = 30
	}
	if len(cfg.General.OpenVPNConfigDirs) == 0 {
		cfg.General.OpenVPNConfigDirs = []string{"/etc/openvpn/client", "~/.config/openvpn"}
	}
	if cfg.General.OpenVPNMethod == "" {
		cfg.General.OpenVPNMethod = "auto"
	}
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
