package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.General.IPCheckURL != "https://api.ipify.org" {
		t.Errorf("got IPCheckURL %q, want %q", cfg.General.IPCheckURL, "https://api.ipify.org")
	}
	if cfg.General.IPCheckInterval != 30 {
		t.Errorf("got IPCheckInterval %d, want 30", cfg.General.IPCheckInterval)
	}
	if cfg.General.OpenVPNMethod != "auto" {
		t.Errorf("got OpenVPNMethod %q, want %q", cfg.General.OpenVPNMethod, "auto")
	}
	if cfg.General.AllowMultiVPN {
		t.Error("AllowMultiVPN should default to false")
	}
	if len(cfg.General.OpenVPNConfigDirs) != 2 {
		t.Errorf("got %d config dirs, want 2", len(cfg.General.OpenVPNConfigDirs))
	}
}

func TestLoad_MissingFileReturnsDefault(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.General.IPCheckURL != "https://api.ipify.org" {
		t.Errorf("missing file should return default config")
	}
}

func TestLoad_ParsesYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	data := []byte(`
general:
  ip_check_url: "https://ifconfig.me"
  ip_check_interval: 60
  openvpn_method: "systemd"
  allow_multi_vpn: true
  openvpn_config_dirs:
    - "/custom/path"

vpn_profiles:
  - name: "Test VPN"
    provider: "openvpn"
    config_path: "/etc/openvpn/client/test.conf"

routing_rules:
  - id: "test-rule"
    description: "Test routing rule"
    enabled: true
    type: "subnet"
    dest_cidr: "10.0.0.0/8"
    vpn_interface: "tun0"
    table: "100"
`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write test config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.General.IPCheckURL != "https://ifconfig.me" {
		t.Errorf("got IPCheckURL %q, want %q", cfg.General.IPCheckURL, "https://ifconfig.me")
	}
	if cfg.General.IPCheckInterval != 60 {
		t.Errorf("got IPCheckInterval %d, want 60", cfg.General.IPCheckInterval)
	}
	if cfg.General.OpenVPNMethod != "systemd" {
		t.Errorf("got OpenVPNMethod %q, want %q", cfg.General.OpenVPNMethod, "systemd")
	}
	if !cfg.General.AllowMultiVPN {
		t.Error("AllowMultiVPN should be true")
	}
	if len(cfg.General.OpenVPNConfigDirs) != 1 || cfg.General.OpenVPNConfigDirs[0] != "/custom/path" {
		t.Errorf("got config dirs %v", cfg.General.OpenVPNConfigDirs)
	}
	if len(cfg.VPNProfiles) != 1 {
		t.Fatalf("got %d profiles, want 1", len(cfg.VPNProfiles))
	}
	if cfg.VPNProfiles[0].Name != "Test VPN" {
		t.Errorf("got profile name %q, want %q", cfg.VPNProfiles[0].Name, "Test VPN")
	}
	if len(cfg.RoutingRules) != 1 {
		t.Fatalf("got %d routing rules, want 1", len(cfg.RoutingRules))
	}
	if cfg.RoutingRules[0].DestCIDR != "10.0.0.0/8" {
		t.Errorf("got dest CIDR %q, want %q", cfg.RoutingRules[0].DestCIDR, "10.0.0.0/8")
	}
}

func TestLoad_AppliesDefaultsForMissingFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	data := []byte(`
general:
  allow_multi_vpn: true
`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write test config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.General.IPCheckURL != "https://api.ipify.org" {
		t.Errorf("missing IPCheckURL should get default, got %q", cfg.General.IPCheckURL)
	}
	if cfg.General.IPCheckInterval != 30 {
		t.Errorf("missing IPCheckInterval should get default, got %d", cfg.General.IPCheckInterval)
	}
	if !cfg.General.AllowMultiVPN {
		t.Error("explicit AllowMultiVPN should be preserved")
	}
}

func TestSave_CreatesDirectoryAndFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.yaml")

	cfg := Default()
	cfg.General.IPCheckURL = "https://example.com/ip"

	if err := Save(path, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error loading saved config: %v", err)
	}

	if loaded.General.IPCheckURL != "https://example.com/ip" {
		t.Errorf("got IPCheckURL %q, want %q", loaded.General.IPCheckURL, "https://example.com/ip")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(path, []byte("{{invalid yaml"), 0o644); err != nil {
		t.Fatalf("write test config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestExpandConfigDirs(t *testing.T) {
	dirs := ExpandConfigDirs([]string{"/etc/openvpn", "~/openvpn"})

	if dirs[0] != "/etc/openvpn" {
		t.Errorf("absolute path should not change: %q", dirs[0])
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "openvpn")
	if dirs[1] != expected {
		t.Errorf("got %q, want %q", dirs[1], expected)
	}
}
