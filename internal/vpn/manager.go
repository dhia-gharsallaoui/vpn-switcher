package vpn

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
)

// Manager orchestrates VPN operations across providers.
type Manager struct {
	providers map[ProviderType]Provider
	executor  system.CommandExecutor
}

// NewManager creates a new VPN manager.
func NewManager(exec system.CommandExecutor, providers ...Provider) *Manager {
	m := &Manager{
		providers: make(map[ProviderType]Provider),
		executor:  exec,
	}
	for _, p := range providers {
		m.providers[p.Type()] = p
	}
	return m
}

// StatusAll discovers and returns status of all VPNs across all providers.
func (m *Manager) StatusAll(ctx context.Context) ([]VPN, error) {
	discovery := NewDiscovery(m.providerList()...)
	return discovery.DiscoverAll(ctx)
}

// Connect establishes a VPN connection.
func (m *Manager) Connect(ctx context.Context, target VPN) error {
	provider, ok := m.providers[target.Provider]
	if !ok {
		return fmt.Errorf("%w: %s", ErrProviderNotFound, target.Provider)
	}
	return provider.Connect(ctx, target)
}

// Disconnect tears down a VPN connection.
func (m *Manager) Disconnect(ctx context.Context, target VPN) error {
	provider, ok := m.providers[target.Provider]
	if !ok {
		return fmt.Errorf("%w: %s", ErrProviderNotFound, target.Provider)
	}
	return provider.Disconnect(ctx, target)
}

// Switch safely disconnects active VPNs and connects the target.
// It waits for the old interface to be removed before connecting.
func (m *Manager) Switch(ctx context.Context, target VPN) error {
	// Find and disconnect active VPNs
	active, err := m.ActiveVPNs(ctx)
	if err != nil {
		return fmt.Errorf("switch: check active: %w", err)
	}

	for _, v := range active {
		provider, ok := m.providers[v.Provider]
		if !ok {
			continue
		}
		if err := provider.Disconnect(ctx, v); err != nil {
			return fmt.Errorf("switch: disconnect %s: %w", v.Name, err)
		}

		// Wait for interface removal
		if err := m.waitForInterfaceRemoval(ctx, provider.InterfaceName()); err != nil {
			return fmt.Errorf("switch: wait for %s removal: %w", provider.InterfaceName(), err)
		}
	}

	if err := m.Connect(ctx, target); err != nil {
		return err
	}

	// Wait for the new interface to be ready
	targetProvider, ok := m.providers[target.Provider]
	if ok {
		if err := m.WaitForInterfaceReady(ctx, targetProvider.InterfaceName()); err != nil {
			return fmt.Errorf("switch: wait for %s ready: %w", targetProvider.InterfaceName(), err)
		}
	}

	return nil
}

// ConnectMulti connects a VPN without disconnecting others (multi-VPN mode).
func (m *Manager) ConnectMulti(ctx context.Context, target VPN) error {
	return m.Connect(ctx, target)
}

// ActiveVPNs returns all currently connected VPNs.
func (m *Manager) ActiveVPNs(ctx context.Context) ([]VPN, error) {
	all, err := m.StatusAll(ctx)
	if err != nil {
		return nil, err
	}

	var active []VPN
	for _, v := range all {
		if v.Status == StatusConnected {
			active = append(active, v)
		}
	}
	return active, nil
}

// waitForInterfaceRemoval polls until the named interface disappears.
func (m *Manager) waitForInterfaceRemoval(ctx context.Context, ifaceName string) error {
	timeout := 10 * time.Second
	poll := 500 * time.Millisecond

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		result, err := m.executor.Run(ctx, "ip", "link", "show", ifaceName)
		if err != nil {
			return fmt.Errorf("check interface %s: %w", ifaceName, err)
		}
		// Interface is gone when the command fails (exit code != 0)
		if result.ExitCode != 0 {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(poll):
		}
	}

	return fmt.Errorf("%w: %s still present after %s", ErrInterfaceTimeout, ifaceName, timeout)
}

// WaitForInterfaceReady polls until the named interface is up with an IP.
func (m *Manager) WaitForInterfaceReady(ctx context.Context, ifaceName string) error {
	timeout := 15 * time.Second
	poll := 500 * time.Millisecond

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		result, err := m.executor.Run(ctx, "ip", "addr", "show", ifaceName)
		if err != nil {
			return fmt.Errorf("check interface %s: %w", ifaceName, err)
		}
		// Interface is ready when command succeeds and output contains "inet "
		if result.ExitCode == 0 && strings.Contains(result.Stdout, "inet ") {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(poll):
		}
	}

	return fmt.Errorf("%w: %s not ready after %s", ErrInterfaceTimeout, ifaceName, timeout)
}

// MergeProfiles merges configured VPN profiles with discovered VPNs.
// Profiles can override display names or add VPNs not found by discovery.
func MergeProfiles(discovered []VPN, profiles []VPNProfile) []VPN {
	if len(profiles) == 0 {
		return discovered
	}

	// Index discovered VPNs by config path and provider+name
	byPath := make(map[string]int)
	byKey := make(map[string]int)
	for i, v := range discovered {
		if v.ConfigPath != "" {
			byPath[v.ConfigPath] = i
		}
		byKey[string(v.Provider)+":"+v.Name] = i
	}

	result := make([]VPN, len(discovered))
	copy(result, discovered)

	for _, p := range profiles {
		provider := ProviderType(p.Provider)

		// Try to match existing discovered VPN
		idx := -1
		if p.ConfigPath != "" {
			if i, ok := byPath[p.ConfigPath]; ok {
				idx = i
			}
		}
		if idx < 0 {
			if i, ok := byKey[p.Provider+":"+p.Name]; ok {
				idx = i
			}
		}

		if idx >= 0 {
			// Override display name if profile specifies one
			if p.Name != "" {
				result[idx].Name = p.Name
			}
			if p.Interface != "" {
				result[idx].Interface = p.Interface
			}
			continue
		}

		// Add new VPN from profile
		v := VPN{
			ID:         fmt.Sprintf("%s:%s", provider, p.Name),
			Name:       p.Name,
			Provider:   provider,
			ConfigPath: p.ConfigPath,
			Interface:  p.Interface,
			ExitNode:   p.ExitNode,
			Status:     StatusDisconnected,
		}
		result = append(result, v)
	}

	return result
}

// VPNProfile mirrors config.VPNProfile to avoid circular imports.
type VPNProfile struct {
	Name       string
	ConfigPath string
	Provider   string
	ExitNode   string
	Interface  string
}

func (m *Manager) providerList() []Provider {
	providers := make([]Provider, 0, len(m.providers))
	for _, p := range m.providers {
		providers = append(providers, p)
	}
	return providers
}

