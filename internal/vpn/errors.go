package vpn

import "errors"

var (
	// ErrAlreadyConnected is returned when attempting to connect a VPN that is already active.
	ErrAlreadyConnected = errors.New("vpn already connected")

	// ErrNotConnected is returned when attempting to disconnect a VPN that is not active.
	ErrNotConnected = errors.New("vpn not connected")

	// ErrInterfaceTimeout is returned when waiting for an interface to appear/disappear times out.
	ErrInterfaceTimeout = errors.New("timed out waiting for interface")

	// ErrProviderNotFound is returned when the requested VPN provider is not registered.
	ErrProviderNotFound = errors.New("vpn provider not found")

	// ErrPermissionDenied is returned when an operation requires elevated privileges.
	ErrPermissionDenied = errors.New("permission denied: operation requires root or sudo")
)
