package vpn

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
)

func makeTailscaleStatusJSON(state string, ips []string, peers map[string]tailscalePeer) string {
	status := tailscaleStatus{
		BackendState: state,
		Self: tailscaleSelf{
			TailscaleIPs: ips,
			HostName:     "myhost",
		},
		Peer: peers,
	}
	data, _ := json.Marshal(status)
	return string(data)
}

func TestTailscaleProvider_DiscoverRunning(t *testing.T) {
	statusJSON := makeTailscaleStatusJSON("Running", []string{"100.64.1.4"}, nil)

	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		if name == "tailscale" && len(args) > 0 && args[0] == "status" {
			return system.ExecResult{Stdout: statusJSON, ExitCode: 0}, nil
		}
		return system.ExecResult{}, nil
	})

	p := NewTailscaleProvider(mock)
	vpns, err := p.Discover(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(vpns) < 1 {
		t.Fatal("expected at least 1 VPN")
	}

	if vpns[0].Status != StatusConnected {
		t.Errorf("got status %v, want StatusConnected", vpns[0].Status)
	}
	if vpns[0].IP != "100.64.1.4" {
		t.Errorf("got IP %q, want %q", vpns[0].IP, "100.64.1.4")
	}
}

func TestTailscaleProvider_DiscoverStopped(t *testing.T) {
	statusJSON := makeTailscaleStatusJSON("Stopped", nil, nil)

	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		return system.ExecResult{Stdout: statusJSON, ExitCode: 0}, nil
	})

	p := NewTailscaleProvider(mock)
	vpns, err := p.Discover(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if vpns[0].Status != StatusDisconnected {
		t.Errorf("got status %v, want StatusDisconnected", vpns[0].Status)
	}
}

func TestTailscaleProvider_DiscoverWithExitNodes(t *testing.T) {
	peers := map[string]tailscalePeer{
		"node1": {HostName: "us-east", ExitNodeOption: true, Online: true, ExitNode: false},
		"node2": {HostName: "eu-west", ExitNodeOption: true, Online: true, ExitNode: true},
		"node3": {HostName: "internal", ExitNodeOption: false, Online: true},
	}
	statusJSON := makeTailscaleStatusJSON("Running", []string{"100.64.1.4"}, peers)

	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		return system.ExecResult{Stdout: statusJSON, ExitCode: 0}, nil
	})

	p := NewTailscaleProvider(mock)
	vpns, err := p.Discover(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 1 default + 2 exit nodes (node3 is not an exit node option)
	if len(vpns) != 3 {
		t.Fatalf("got %d VPNs, want 3", len(vpns))
	}

	// Check that active exit node is marked connected
	var activeExit *VPN
	for i := range vpns {
		if vpns[i].ExitNode == "eu-west" {
			activeExit = &vpns[i]
			break
		}
	}
	if activeExit == nil {
		t.Fatal("eu-west exit node not found")
	}
	if activeExit.Status != StatusConnected {
		t.Errorf("active exit node should be connected, got %v", activeExit.Status)
	}
}

func TestTailscaleProvider_DiscoverNotInstalled(t *testing.T) {
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		return system.ExecResult{}, errors.New("tailscale not found")
	})

	p := NewTailscaleProvider(mock)
	vpns, err := p.Discover(context.Background())
	if err != nil {
		t.Fatalf("not installed should not error: %v", err)
	}
	if vpns != nil {
		t.Errorf("not installed should return nil, got %v", vpns)
	}
}

func TestTailscaleProvider_Connect(t *testing.T) {
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		return system.ExecResult{ExitCode: 0}, nil
	})

	p := NewTailscaleProvider(mock)
	err := p.Connect(context.Background(), VPN{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	calls := mock.Calls()
	if calls[0].Args[0] != "up" {
		t.Errorf("expected 'up', got %v", calls[0].Args)
	}
}

func TestTailscaleProvider_ConnectWithExitNode(t *testing.T) {
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		return system.ExecResult{ExitCode: 0}, nil
	})

	p := NewTailscaleProvider(mock)
	err := p.Connect(context.Background(), VPN{ExitNode: "us-east"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	calls := mock.Calls()
	if len(calls[0].Args) != 2 || calls[0].Args[1] != "--exit-node=us-east" {
		t.Errorf("expected exit node arg, got %v", calls[0].Args)
	}
}

func TestTailscaleProvider_Disconnect(t *testing.T) {
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		return system.ExecResult{ExitCode: 0}, nil
	})

	p := NewTailscaleProvider(mock)
	err := p.Disconnect(context.Background(), VPN{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	calls := mock.Calls()
	if calls[0].Args[0] != "down" {
		t.Errorf("expected 'down', got %v", calls[0].Args)
	}
}

func TestTailscaleProvider_DisconnectExitNode(t *testing.T) {
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		return system.ExecResult{ExitCode: 0}, nil
	})

	p := NewTailscaleProvider(mock)
	err := p.Disconnect(context.Background(), VPN{ExitNode: "us-east"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	calls := mock.Calls()
	// Should use "up --exit-node=" to unset, not "down"
	if calls[0].Args[0] != "up" || calls[0].Args[1] != "--exit-node=" {
		t.Errorf("expected 'up --exit-node=', got %v", calls[0].Args)
	}
}
