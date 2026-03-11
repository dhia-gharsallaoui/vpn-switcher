package vpn

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Discovery finds available VPN configurations across all providers.
type Discovery struct {
	providers []Provider
}

// NewDiscovery creates a new Discovery with the given providers.
func NewDiscovery(providers ...Provider) *Discovery {
	return &Discovery{providers: providers}
}

// DiscoverAll scans all providers and returns all discovered VPNs.
func (d *Discovery) DiscoverAll(ctx context.Context) ([]VPN, error) {
	var all []VPN
	var errs []string

	for _, p := range d.providers {
		vpns, err := p.Discover(ctx)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", p.Type(), err))
			continue
		}
		all = append(all, vpns...)
	}

	if len(errs) > 0 && len(all) == 0 {
		return nil, fmt.Errorf("discovery failed: %s", strings.Join(errs, "; "))
	}

	return all, nil
}

// GlobConfigs scans directories for OpenVPN config files (.conf, .ovpn).
func GlobConfigs(dirs []string) ([]string, error) {
	var configs []string
	seen := make(map[string]bool)

	for _, dir := range dirs {
		dir = expandHome(dir)

		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}

		for _, pattern := range []string{"*.conf", "*.ovpn"} {
			matches, err := filepath.Glob(filepath.Join(dir, pattern))
			if err != nil {
				return nil, fmt.Errorf("glob %s/%s: %w", dir, pattern, err)
			}
			for _, m := range matches {
				if !seen[m] {
					seen[m] = true
					configs = append(configs, m)
				}
			}
		}
	}

	return configs, nil
}

// ConfigName extracts the VPN name from a config file path.
// e.g., "/etc/openvpn/client/work.conf" -> "work"
func ConfigName(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
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
