package vpn

import "context"

// Provider defines operations for a specific VPN technology.
type Provider interface {
	// Type returns the provider type.
	Type() ProviderType

	// Discover finds available VPN configurations.
	Discover(ctx context.Context) ([]VPN, error)

	// Status checks the current connection status of a specific VPN.
	Status(ctx context.Context, v VPN) (ConnectionStatus, error)

	// Connect establishes the VPN connection.
	Connect(ctx context.Context, v VPN) error

	// Disconnect tears down the VPN connection.
	Disconnect(ctx context.Context, v VPN) error

	// InterfaceName returns the network interface this provider uses.
	InterfaceName() string
}
