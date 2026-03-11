package vpn

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
)

func TestManager_StatusAll(t *testing.T) {
	tsStatus := makeTailscaleStatusJSON("Running", []string{"100.64.1.4"}, nil)

	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		if name == "tailscale" {
			return system.ExecResult{Stdout: tsStatus, ExitCode: 0}, nil
		}
		return system.ExecResult{ExitCode: 0}, nil
	})

	ts := NewTailscaleProvider(mock)
	m := NewManager(mock, ts)

	vpns, err := m.StatusAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vpns) == 0 {
		t.Fatal("expected at least 1 VPN")
	}
}

func TestManager_ConnectDisconnect(t *testing.T) {
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		return system.ExecResult{ExitCode: 0}, nil
	})

	ts := NewTailscaleProvider(mock)
	m := NewManager(mock, ts)

	v := VPN{Provider: ProviderTailscale}
	if err := m.Connect(context.Background(), v); err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := m.Disconnect(context.Background(), v); err != nil {
		t.Fatalf("disconnect: %v", err)
	}
}

func TestManager_ConnectUnknownProvider(t *testing.T) {
	mock := system.NewMockExecutor(nil)
	m := NewManager(mock)

	err := m.Connect(context.Background(), VPN{Provider: "wireguard"})
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestManager_ActiveVPNs(t *testing.T) {
	tsStatus := tailscaleStatus{
		BackendState: "Running",
		Self:         tailscaleSelf{TailscaleIPs: []string{"100.64.1.4"}},
	}
	tsJSON, _ := json.Marshal(tsStatus)

	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		if name == "tailscale" {
			return system.ExecResult{Stdout: string(tsJSON), ExitCode: 0}, nil
		}
		return system.ExecResult{ExitCode: 0}, nil
	})

	ts := NewTailscaleProvider(mock)
	m := NewManager(mock, ts)

	active, err := m.ActiveVPNs(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("got %d active VPNs, want 1", len(active))
	}
	if active[0].Status != StatusConnected {
		t.Errorf("active VPN should be connected")
	}
}

func TestManager_Switch(t *testing.T) {
	callCount := 0
	interfaceExists := true

	tsStatus := makeTailscaleStatusJSON("Running", []string{"100.64.1.4"}, nil)

	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		callCount++

		// Tailscale status
		if name == "tailscale" && len(args) > 0 && args[0] == "status" {
			return system.ExecResult{Stdout: tsStatus, ExitCode: 0}, nil
		}

		// Tailscale down
		if name == "tailscale" && len(args) > 0 && args[0] == "down" {
			interfaceExists = false
			// Update status to stopped for subsequent calls
			tsStatus = makeTailscaleStatusJSON("Stopped", nil, nil)
			return system.ExecResult{ExitCode: 0}, nil
		}

		// Tailscale up
		if name == "tailscale" && len(args) > 0 && args[0] == "up" {
			return system.ExecResult{ExitCode: 0}, nil
		}

		// ip link show (check interface)
		if name == "ip" && len(args) >= 3 && args[0] == "link" && args[1] == "show" {
			if interfaceExists {
				return system.ExecResult{Stdout: fmt.Sprintf("2: %s: <POINTOPOINT>", args[2]), ExitCode: 0}, nil
			}
			return system.ExecResult{Stderr: "Device does not exist", ExitCode: 1}, nil
		}

		// systemctl start (OpenVPN connect)
		if name == "systemctl" && len(args) >= 1 && args[0] == "start" {
			return system.ExecResult{ExitCode: 0}, nil
		}

		// systemctl is-active (OpenVPN status)
		if name == "systemctl" && len(args) >= 1 && args[0] == "is-active" {
			return system.ExecResult{Stdout: "inactive", ExitCode: 3}, nil
		}

		// systemctl list-unit-files (auto-detect)
		if name == "systemctl" && len(args) >= 1 && args[0] == "list-unit-files" {
			return system.ExecResult{
				Stdout:   "openvpn-client@.service enabled",
				ExitCode: 0,
			}, nil
		}

		return system.ExecResult{ExitCode: 0}, nil
	})

	ts := NewTailscaleProvider(mock)
	ovpn := NewOpenVPNProvider(mock, nil, "systemd")
	m := NewManager(mock, ts, ovpn)

	target := VPN{
		Provider:   ProviderOpenVPN,
		ConfigPath: "/etc/openvpn/client/work.conf",
	}

	err := m.Switch(context.Background(), target)
	if err != nil {
		t.Fatalf("switch: %v", err)
	}

	// Verify that tailscale down was called and systemctl start was called
	calls := mock.Calls()
	var downCalled, startCalled bool
	for _, c := range calls {
		if c.Name == "tailscale" && len(c.Args) > 0 && c.Args[0] == "down" {
			downCalled = true
		}
		if c.Name == "systemctl" && len(c.Args) > 0 && c.Args[0] == "start" && strings.Contains(c.Args[1], "work") {
			startCalled = true
		}
	}
	if !downCalled {
		t.Error("expected tailscale down to be called")
	}
	if !startCalled {
		t.Error("expected systemctl start to be called")
	}
}
