package ui

import (
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/network"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/vpn"
)

// VPNListMsg carries discovered VPN list.
type VPNListMsg struct {
	VPNs []vpn.VPN
	Err  error
}

// ConnectResultMsg carries the result of a connect/switch operation.
type ConnectResultMsg struct {
	VPN vpn.VPN
	Err error
}

// DisconnectResultMsg carries the result of a disconnect operation.
type DisconnectResultMsg struct {
	VPN vpn.VPN
	Err error
}

// PublicIPMsg carries the fetched public IP.
type PublicIPMsg struct {
	IP  string
	Err error
}

// InterfaceListMsg carries interface status.
type InterfaceListMsg struct {
	Interfaces []network.InterfaceInfo
	Err        error
}

// RoutingRuleAppliedMsg carries the result of applying a routing rule.
type RoutingRuleAppliedMsg struct {
	Rule network.RoutingRule
	Err  error
}

// RoutingRuleRemovedMsg carries the result of removing a routing rule.
type RoutingRuleRemovedMsg struct {
	Rule network.RoutingRule
	Err  error
}

// TickMsg triggers periodic status refresh.
type TickMsg struct{}

// RouteListMsg carries route table listing results.
type RouteListMsg struct {
	Routes []network.Route
	Table  string
	Err    error
}

// PingResultMsg carries a single interface ping result.
type PingResultMsg struct {
	Result network.PingResult
	Err    error
}

// DNSResultMsg carries DNS lookup results.
type DNSResultMsg struct {
	Results []network.DNSResult
	Err     error
}

// TracerouteResultMsg carries traceroute results.
type TracerouteResultMsg struct {
	Hops []network.Hop
	Err  error
}

// StatusMsg is a general status message for the status bar.
type StatusMsg struct {
	Text    string
	IsError bool
}
