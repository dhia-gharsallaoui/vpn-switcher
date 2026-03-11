package ui

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/config"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/network"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/vpn"
)

func timeout(seconds int) time.Duration {
	return time.Duration(seconds) * time.Second
}

func discoverVPNsCmd(mgr *vpn.Manager, profiles []vpn.VPNProfile, t config.TimeoutConfig) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout(t.Discovery))
		defer cancel()

		vpns, err := mgr.StatusAll(ctx)
		if err == nil && len(profiles) > 0 {
			vpns = vpn.MergeProfiles(vpns, profiles)
		}
		return VPNListMsg{VPNs: vpns, Err: err}
	}
}

func connectVPNCmd(mgr *vpn.Manager, target vpn.VPN, t config.TimeoutConfig) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout(t.Connect))
		defer cancel()

		err := mgr.Connect(ctx, target)
		return ConnectResultMsg{VPN: target, Err: err}
	}
}

func switchVPNCmd(mgr *vpn.Manager, target vpn.VPN, t config.TimeoutConfig) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout(t.Switch))
		defer cancel()

		err := mgr.Switch(ctx, target)
		return ConnectResultMsg{VPN: target, Err: err}
	}
}

func connectMultiVPNCmd(mgr *vpn.Manager, target vpn.VPN, t config.TimeoutConfig) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout(t.Connect))
		defer cancel()

		err := mgr.ConnectMulti(ctx, target)
		return ConnectResultMsg{VPN: target, Err: err}
	}
}

func disconnectVPNCmd(mgr *vpn.Manager, target vpn.VPN, t config.TimeoutConfig) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout(t.Connect))
		defer cancel()

		err := mgr.Disconnect(ctx, target)
		return DisconnectResultMsg{VPN: target, Err: err}
	}
}

func fetchPublicIPCmd(fetcher *network.IPFetcher, t config.TimeoutConfig) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout(t.IPFetch))
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

func listRoutesCmd(executor system.CommandExecutor, tableName string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		routes, err := network.ListRoutes(ctx, executor, tableName)
		return RouteListMsg{Routes: routes, Table: tableName, Err: err}
	}
}

func pingInterfaceCmd(executor system.CommandExecutor, ifaceName, target string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := network.PingGateway(ctx, executor, ifaceName, target)
		return PingResultMsg{Result: result, Err: err}
	}
}

func dnsLookupCmd(executor system.CommandExecutor, domain, recordType string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		results, err := network.Lookup(ctx, executor, domain, recordType)
		return DNSResultMsg{Results: results, Err: err}
	}
}

func tracerouteCmd(executor system.CommandExecutor, target, iface string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		hops, err := network.Traceroute(ctx, executor, target, iface)
		return TracerouteResultMsg{Hops: hops, Err: err}
	}
}
