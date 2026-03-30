package vpn

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
)

// tailscaleStatus represents the JSON output of `tailscale status --json`.
type tailscaleStatus struct {
	BackendState string                   `json:"BackendState"`
	Self         tailscaleSelf            `json:"Self"`
	Peer         map[string]tailscalePeer `json:"Peer"`
}

type tailscaleSelf struct {
	TailscaleIPs []string `json:"TailscaleIPs"`
	HostName     string   `json:"HostName"`
}

type tailscalePeer struct {
	HostName       string   `json:"HostName"`
	TailscaleIPs   []string `json:"TailscaleIPs"`
	ExitNode       bool     `json:"ExitNode"`
	ExitNodeOption bool     `json:"ExitNodeOption"`
	Online         bool     `json:"Online"`
}

// TailscaleProvider manages Tailscale connections.
type TailscaleProvider struct {
	executor      system.CommandExecutor
	interfaceName string
}

// NewTailscaleProvider creates a new Tailscale provider.
func NewTailscaleProvider(exec system.CommandExecutor) *TailscaleProvider {
	return &TailscaleProvider{
		executor:      exec,
		interfaceName: "tailscale0",
	}
}

// SetInterfaceName overrides the default interface name.
func (p *TailscaleProvider) SetInterfaceName(name string) {
	if name != "" {
		p.interfaceName = name
	}
}

func (p *TailscaleProvider) Type() ProviderType    { return ProviderTailscale }
func (p *TailscaleProvider) InterfaceName() string { return p.interfaceName }

// Discover checks if Tailscale is installed and returns its status.
func (p *TailscaleProvider) Discover(ctx context.Context) ([]VPN, error) {
	status, err := p.getStatus(ctx)
	if err != nil {
		return nil, nil // Tailscale not installed or not running
	}

	v := VPN{
		ID:        "tailscale:default",
		Name:      "Tailscale",
		Provider:  ProviderTailscale,
		Interface: p.interfaceName,
	}

	switch status.BackendState {
	case "Running":
		v.Status = StatusConnected
		if len(status.Self.TailscaleIPs) > 0 {
			v.IP = status.Self.TailscaleIPs[0]
		}
	case "Stopped":
		v.Status = StatusDisconnected
	case "Starting":
		v.Status = StatusConnecting
	default:
		v.Status = StatusDisconnected
	}

	vpns := []VPN{v}

	// Discover exit nodes
	for _, peer := range status.Peer {
		if peer.ExitNodeOption && peer.Online {
			exitVPN := VPN{
				ID:        fmt.Sprintf("tailscale:exit:%s", peer.HostName),
				Name:      fmt.Sprintf("Tailscale Exit (%s)", peer.HostName),
				Provider:  ProviderTailscale,
				Interface: p.interfaceName,
				ExitNode:  peer.HostName,
				Status:    StatusDisconnected,
			}
			if peer.ExitNode {
				exitVPN.Status = StatusConnected
			}
			vpns = append(vpns, exitVPN)
		}
	}

	return vpns, nil
}

// Status checks the current Tailscale connection status.
func (p *TailscaleProvider) Status(ctx context.Context, v VPN) (ConnectionStatus, error) {
	status, err := p.getStatus(ctx)
	if err != nil {
		return StatusError, fmt.Errorf("tailscale status: %w", err)
	}

	switch status.BackendState {
	case "Running":
		// If this is an exit node VPN, check if it's the active exit node
		if v.ExitNode != "" {
			for _, peer := range status.Peer {
				if peer.HostName == v.ExitNode && peer.ExitNode {
					return StatusConnected, nil
				}
			}
			return StatusDisconnected, nil
		}
		return StatusConnected, nil
	case "Stopped":
		return StatusDisconnected, nil
	case "Starting":
		return StatusConnecting, nil
	default:
		return StatusDisconnected, nil
	}
}

// Connect starts Tailscale.
func (p *TailscaleProvider) Connect(ctx context.Context, v VPN) error {
	args := []string{"up"}
	if v.ExitNode != "" {
		args = append(args, fmt.Sprintf("--exit-node=%s", v.ExitNode))
	}

	result, err := p.executor.Run(ctx, "tailscale", args...)
	if err != nil {
		return fmt.Errorf("tailscale connect: %w", err)
	}
	if result.ExitCode != 0 {
		if strings.Contains(result.Stderr, "permission denied") || strings.Contains(result.Stderr, "Access denied") {
			return fmt.Errorf("tailscale connect: %w", ErrPermissionDenied)
		}
		return fmt.Errorf("tailscale connect: %s", result.Stderr)
	}
	return nil
}

// Disconnect stops Tailscale.
func (p *TailscaleProvider) Disconnect(ctx context.Context, v VPN) error {
	// If disconnecting an exit node, just unset the exit node
	if v.ExitNode != "" {
		result, err := p.executor.Run(ctx, "tailscale", "up", "--exit-node=")
		if err != nil {
			return fmt.Errorf("tailscale unset exit node: %w", err)
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("tailscale unset exit node: %s", result.Stderr)
		}
		return nil
	}

	result, err := p.executor.Run(ctx, "tailscale", "down")
	if err != nil {
		return fmt.Errorf("tailscale disconnect: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("tailscale disconnect: %s", result.Stderr)
	}
	return nil
}

func (p *TailscaleProvider) getStatus(ctx context.Context) (*tailscaleStatus, error) {
	result, err := p.executor.Run(ctx, "tailscale", "status", "--json")
	if err != nil {
		return nil, err
	}
	if result.ExitCode != 0 {
		return nil, fmt.Errorf("tailscale status: exit code %d", result.ExitCode)
	}

	var status tailscaleStatus
	if err := json.Unmarshal([]byte(result.Stdout), &status); err != nil {
		return nil, fmt.Errorf("parse tailscale status: %w", err)
	}

	return &status, nil
}
