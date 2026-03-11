package network

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
)

// InterfaceMonitor checks network interface status.
type InterfaceMonitor struct {
	executor system.CommandExecutor
}

// NewInterfaceMonitor creates a new interface monitor.
func NewInterfaceMonitor(exec system.CommandExecutor) *InterfaceMonitor {
	return &InterfaceMonitor{executor: exec}
}

// List returns info about VPN-related interfaces (tun*, tailscale*).
func (m *InterfaceMonitor) List(ctx context.Context) ([]InterfaceInfo, error) {
	result, err := m.executor.Run(ctx, "ip", "-o", "addr", "show")
	if err != nil {
		return nil, fmt.Errorf("list interfaces: %w", err)
	}

	var interfaces []InterfaceInfo
	seen := make(map[string]bool)

	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		name := strings.TrimSuffix(fields[1], ":")
		if !isVPNInterface(name) {
			continue
		}

		if seen[name] {
			// Add additional address to existing interface
			for i := range interfaces {
				if interfaces[i].Name == name {
					addr := extractAddr(fields)
					if addr != "" {
						interfaces[i].Addrs = append(interfaces[i].Addrs, addr)
					}
				}
			}
			continue
		}

		seen[name] = true
		iface := InterfaceInfo{
			Name: name,
			Up:   strings.Contains(line, "UP"),
		}

		addr := extractAddr(fields)
		if addr != "" {
			iface.Addrs = append(iface.Addrs, addr)
		}

		interfaces = append(interfaces, iface)
	}

	return interfaces, nil
}

// WaitForRemoval blocks until the named interface disappears.
func (m *InterfaceMonitor) WaitForRemoval(ctx context.Context, ifaceName string) error {
	return m.waitFor(ctx, ifaceName, false, 10*time.Second)
}

// WaitForReady blocks until the named interface is up with an IP.
func (m *InterfaceMonitor) WaitForReady(ctx context.Context, ifaceName string) error {
	return m.waitFor(ctx, ifaceName, true, 15*time.Second)
}

func (m *InterfaceMonitor) waitFor(ctx context.Context, ifaceName string, wantUp bool, timeout time.Duration) error {
	poll := 500 * time.Millisecond
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		result, err := m.executor.Run(ctx, "ip", "addr", "show", ifaceName)
		if err != nil {
			return fmt.Errorf("check interface %s: %w", ifaceName, err)
		}

		exists := result.ExitCode == 0
		hasIP := strings.Contains(result.Stdout, "inet ")

		if wantUp && exists && hasIP {
			return nil
		}
		if !wantUp && !exists {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(poll):
		}
	}

	action := "removal"
	if wantUp {
		action = "ready"
	}
	return fmt.Errorf("timed out waiting for %s %s after %s", ifaceName, action, timeout)
}

func isVPNInterface(name string) bool {
	return strings.HasPrefix(name, "tun") || strings.HasPrefix(name, "tailscale")
}

func extractAddr(fields []string) string {
	for i, f := range fields {
		if f == "inet" || f == "inet6" {
			if i+1 < len(fields) {
				return fields[i+1]
			}
		}
	}
	return ""
}
