package ui

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/network"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/vpn"
)

func discoverVPNsCmd(mgr *vpn.Manager) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		vpns, err := mgr.StatusAll(ctx)
		return VPNListMsg{VPNs: vpns, Err: err}
	}
}

func connectVPNCmd(mgr *vpn.Manager, target vpn.VPN) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := mgr.Connect(ctx, target)
		return ConnectResultMsg{VPN: target, Err: err}
	}
}

func switchVPNCmd(mgr *vpn.Manager, target vpn.VPN) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		err := mgr.Switch(ctx, target)
		return ConnectResultMsg{VPN: target, Err: err}
	}
}

func connectMultiVPNCmd(mgr *vpn.Manager, target vpn.VPN) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := mgr.ConnectMulti(ctx, target)
		return ConnectResultMsg{VPN: target, Err: err}
	}
}

func disconnectVPNCmd(mgr *vpn.Manager, target vpn.VPN) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := mgr.Disconnect(ctx, target)
		return DisconnectResultMsg{VPN: target, Err: err}
	}
}

func fetchPublicIPCmd(fetcher *network.IPFetcher) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ip, err := fetcher.FetchPublicIP(ctx)
		return PublicIPMsg{IP: ip, Err: err}
	}
}

func listInterfacesCmd(mon *network.InterfaceMonitor) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		ifaces, err := mon.List(ctx)
		return InterfaceListMsg{Interfaces: ifaces, Err: err}
	}
}

func applyRoutingRuleCmd(rm *network.RoutingManager, rule network.RoutingRule) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := rm.Apply(ctx, rule)
		return RoutingRuleAppliedMsg{Rule: rule, Err: err}
	}
}

func removeRoutingRuleCmd(rm *network.RoutingManager, rule network.RoutingRule) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := rm.Remove(ctx, rule)
		return RoutingRuleRemovedMsg{Rule: rule, Err: err}
	}
}

func tickCmd(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}
