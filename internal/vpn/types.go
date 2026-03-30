package vpn

// ProviderType identifies the VPN technology.
type ProviderType string

const (
	ProviderOpenVPN   ProviderType = "openvpn"
	ProviderTailscale ProviderType = "tailscale"
)

// ConnectionStatus represents the state of a VPN connection.
type ConnectionStatus int

const (
	StatusDisconnected ConnectionStatus = iota
	StatusConnecting
	StatusConnected
	StatusDisconnecting
	StatusError
)

// String returns a human-readable status.
func (s ConnectionStatus) String() string {
	switch s {
	case StatusDisconnected:
		return "disconnected"
	case StatusConnecting:
		return "connecting"
	case StatusConnected:
		return "connected"
	case StatusDisconnecting:
		return "disconnecting"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

// VPN represents a discovered or configured VPN endpoint.
type VPN struct {
	ID         string
	Name       string
	Provider   ProviderType
	ConfigPath string // Path to .conf/.ovpn file (OpenVPN only)
	Interface  string // Expected interface name (tun0, tailscale0)
	Status     ConnectionStatus
	IP         string // IP address assigned by VPN
	ExitNode   string // Tailscale exit node (Tailscale only)
}
