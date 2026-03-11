package vpn

import (
	"context"
	"fmt"
	"strings"

	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
)

// OpenVPNMethod determines how OpenVPN connections are managed.
type OpenVPNMethod string

const (
	MethodAuto    OpenVPNMethod = "auto"
	MethodSystemd OpenVPNMethod = "systemd"
	MethodNmcli   OpenVPNMethod = "nmcli"
)

// OpenVPNProvider manages OpenVPN connections.
type OpenVPNProvider struct {
	executor      system.CommandExecutor
	configDirs    []string
	method        OpenVPNMethod
	interfaceName string
}

// NewOpenVPNProvider creates a new OpenVPN provider.
func NewOpenVPNProvider(exec system.CommandExecutor, configDirs []string, method string) *OpenVPNProvider {
	m := OpenVPNMethod(method)
	if m != MethodSystemd && m != MethodNmcli {
		m = MethodAuto
	}
	return &OpenVPNProvider{
		executor:      exec,
		configDirs:    configDirs,
		method:        m,
		interfaceName: "tun0",
	}
}

// SetInterfaceName overrides the default interface name.
func (p *OpenVPNProvider) SetInterfaceName(name string) {
	if name != "" {
		p.interfaceName = name
	}
}

func (p *OpenVPNProvider) Type() ProviderType    { return ProviderOpenVPN }
func (p *OpenVPNProvider) InterfaceName() string { return p.interfaceName }

// Discover finds OpenVPN configs and checks their status.
func (p *OpenVPNProvider) Discover(ctx context.Context) ([]VPN, error) {
	configs, err := GlobConfigs(p.configDirs)
	if err != nil {
		return nil, fmt.Errorf("openvpn discover: %w", err)
	}

	var vpns []VPN
	for _, cfg := range configs {
		name := ConfigName(cfg)
		v := VPN{
			ID:         fmt.Sprintf("openvpn:%s", name),
			Name:       name,
			Provider:   ProviderOpenVPN,
			ConfigPath: cfg,
			Interface:  p.interfaceName,
		}

		status, err := p.Status(ctx, v)
		if err == nil {
			v.Status = status
		}

		vpns = append(vpns, v)
	}

	return vpns, nil
}

// Status checks whether a specific OpenVPN connection is active.
func (p *OpenVPNProvider) Status(ctx context.Context, v VPN) (ConnectionStatus, error) {
	method := p.resolveMethod(ctx)

	switch method {
	case MethodSystemd:
		return p.statusSystemd(ctx, v)
	case MethodNmcli:
		return p.statusNmcli(ctx, v)
	default:
		return StatusDisconnected, nil
	}
}

// Connect establishes an OpenVPN connection.
func (p *OpenVPNProvider) Connect(ctx context.Context, v VPN) error {
	method := p.resolveMethod(ctx)

	switch method {
	case MethodSystemd:
		return p.connectSystemd(ctx, v)
	case MethodNmcli:
		return p.connectNmcli(ctx, v)
	default:
		return fmt.Errorf("openvpn: no suitable connection method found")
	}
}

// Disconnect tears down an OpenVPN connection.
func (p *OpenVPNProvider) Disconnect(ctx context.Context, v VPN) error {
	method := p.resolveMethod(ctx)

	switch method {
	case MethodSystemd:
		return p.disconnectSystemd(ctx, v)
	case MethodNmcli:
		return p.disconnectNmcli(ctx, v)
	default:
		return fmt.Errorf("openvpn: no suitable disconnection method found")
	}
}

// resolveMethod detects which method to use if set to auto.
func (p *OpenVPNProvider) resolveMethod(ctx context.Context) OpenVPNMethod {
	if p.method != MethodAuto {
		return p.method
	}

	// Try systemd first
	result, err := p.executor.Run(ctx, "systemctl", "list-unit-files", "openvpn-client@.service")
	if err == nil && result.ExitCode == 0 && strings.Contains(result.Stdout, "openvpn-client@") {
		return MethodSystemd
	}

	// Try nmcli
	result, err = p.executor.Run(ctx, "nmcli", "--version")
	if err == nil && result.ExitCode == 0 {
		return MethodNmcli
	}

	return MethodSystemd // fallback
}

// systemd methods

func (p *OpenVPNProvider) statusSystemd(ctx context.Context, v VPN) (ConnectionStatus, error) {
	name := ConfigName(v.ConfigPath)
	result, err := p.executor.Run(ctx, "systemctl", "is-active", fmt.Sprintf("openvpn-client@%s", name))
	if err != nil {
		return StatusError, fmt.Errorf("openvpn status %s: %w", name, err)
	}

	if strings.TrimSpace(result.Stdout) == "active" {
		return StatusConnected, nil
	}
	return StatusDisconnected, nil
}

func (p *OpenVPNProvider) connectSystemd(ctx context.Context, v VPN) error {
	name := ConfigName(v.ConfigPath)
	result, err := p.executor.Run(ctx, "systemctl", "start", fmt.Sprintf("openvpn-client@%s", name))
	if err != nil {
		return fmt.Errorf("openvpn connect %s: %w", name, err)
	}
	if result.ExitCode != 0 {
		if strings.Contains(result.Stderr, "Access denied") || strings.Contains(result.Stderr, "Permission denied") {
			return fmt.Errorf("openvpn connect %s: %w", name, ErrPermissionDenied)
		}
		return fmt.Errorf("openvpn connect %s: %s", name, result.Stderr)
	}
	return nil
}

func (p *OpenVPNProvider) disconnectSystemd(ctx context.Context, v VPN) error {
	name := ConfigName(v.ConfigPath)
	result, err := p.executor.Run(ctx, "systemctl", "stop", fmt.Sprintf("openvpn-client@%s", name))
	if err != nil {
		return fmt.Errorf("openvpn disconnect %s: %w", name, err)
	}
	if result.ExitCode != 0 {
		if strings.Contains(result.Stderr, "Access denied") || strings.Contains(result.Stderr, "Permission denied") {
			return fmt.Errorf("openvpn disconnect %s: %w", name, ErrPermissionDenied)
		}
		return fmt.Errorf("openvpn disconnect %s: %s", name, result.Stderr)
	}
	return nil
}

// nmcli methods

func (p *OpenVPNProvider) statusNmcli(ctx context.Context, v VPN) (ConnectionStatus, error) {
	name := ConfigName(v.ConfigPath)
	result, err := p.executor.Run(ctx, "nmcli", "-t", "-f", "NAME,TYPE,DEVICE", "connection", "show", "--active")
	if err != nil {
		return StatusError, fmt.Errorf("openvpn nmcli status: %w", err)
	}

	for _, line := range strings.Split(result.Stdout, "\n") {
		fields := strings.Split(line, ":")
		if len(fields) >= 2 && fields[0] == name && fields[1] == "vpn" {
			return StatusConnected, nil
		}
	}
	return StatusDisconnected, nil
}

func (p *OpenVPNProvider) connectNmcli(ctx context.Context, v VPN) error {
	name := ConfigName(v.ConfigPath)
	result, err := p.executor.Run(ctx, "nmcli", "connection", "up", name)
	if err != nil {
		return fmt.Errorf("openvpn nmcli connect %s: %w", name, err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("openvpn nmcli connect %s: %s", name, result.Stderr)
	}
	return nil
}

func (p *OpenVPNProvider) disconnectNmcli(ctx context.Context, v VPN) error {
	name := ConfigName(v.ConfigPath)
	result, err := p.executor.Run(ctx, "nmcli", "connection", "down", name)
	if err != nil {
		return fmt.Errorf("openvpn nmcli disconnect %s: %w", name, err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("openvpn nmcli disconnect %s: %s", name, result.Stderr)
	}
	return nil
}
