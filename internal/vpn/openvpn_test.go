package vpn

import (
	"context"
	"testing"

	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
)

func TestOpenVPNProvider_StatusSystemd(t *testing.T) {
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		if name == "systemctl" && len(args) >= 2 && args[0] == "is-active" {
			if args[1] == "openvpn-client@work" {
				return system.ExecResult{Stdout: "active", ExitCode: 0}, nil
			}
			return system.ExecResult{Stdout: "inactive", ExitCode: 3}, nil
		}
		return system.ExecResult{}, nil
	})

	p := NewOpenVPNProvider(mock, nil, "systemd")

	v := VPN{ConfigPath: "/etc/openvpn/client/work.conf"}
	status, err := p.Status(context.Background(), v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != StatusConnected {
		t.Errorf("got status %v, want StatusConnected", status)
	}

	v2 := VPN{ConfigPath: "/etc/openvpn/client/home.conf"}
	status2, err := p.Status(context.Background(), v2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status2 != StatusDisconnected {
		t.Errorf("got status %v, want StatusDisconnected", status2)
	}
}

func TestOpenVPNProvider_ConnectSystemd(t *testing.T) {
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		if name == "systemctl" && args[0] == "start" {
			return system.ExecResult{ExitCode: 0}, nil
		}
		return system.ExecResult{}, nil
	})

	p := NewOpenVPNProvider(mock, nil, "systemd")
	v := VPN{ConfigPath: "/etc/openvpn/client/work.conf"}

	err := p.Connect(context.Background(), v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	calls := mock.Calls()
	if len(calls) != 1 {
		t.Fatalf("got %d calls, want 1", len(calls))
	}
	if calls[0].Name != "systemctl" || calls[0].Args[0] != "start" || calls[0].Args[1] != "openvpn-client@work" {
		t.Errorf("unexpected call: %v", calls[0])
	}
}

func TestOpenVPNProvider_DisconnectSystemd(t *testing.T) {
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		return system.ExecResult{ExitCode: 0}, nil
	})

	p := NewOpenVPNProvider(mock, nil, "systemd")
	v := VPN{ConfigPath: "/etc/openvpn/client/work.conf"}

	err := p.Disconnect(context.Background(), v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	calls := mock.Calls()
	if calls[0].Args[0] != "stop" || calls[0].Args[1] != "openvpn-client@work" {
		t.Errorf("unexpected call: %v", calls[0])
	}
}

func TestOpenVPNProvider_ConnectPermissionDenied(t *testing.T) {
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		return system.ExecResult{ExitCode: 4, Stderr: "Access denied"}, nil
	})

	p := NewOpenVPNProvider(mock, nil, "systemd")
	v := VPN{ConfigPath: "/etc/openvpn/client/work.conf"}

	err := p.Connect(context.Background(), v)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestOpenVPNProvider_StatusNmcli(t *testing.T) {
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		if name == "nmcli" {
			return system.ExecResult{
				Stdout:   "work:vpn:tun0\neth0:ethernet:eth0",
				ExitCode: 0,
			}, nil
		}
		return system.ExecResult{}, nil
	})

	p := NewOpenVPNProvider(mock, nil, "nmcli")
	v := VPN{ConfigPath: "/etc/openvpn/client/work.conf"}

	status, err := p.Status(context.Background(), v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != StatusConnected {
		t.Errorf("got status %v, want StatusConnected", status)
	}
}

func TestOpenVPNProvider_AutoDetect(t *testing.T) {
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		if name == "systemctl" && args[0] == "list-unit-files" {
			return system.ExecResult{
				Stdout:   "openvpn-client@.service enabled",
				ExitCode: 0,
			}, nil
		}
		if name == "systemctl" && args[0] == "is-active" {
			return system.ExecResult{Stdout: "active", ExitCode: 0}, nil
		}
		return system.ExecResult{ExitCode: 0}, nil
	})

	p := NewOpenVPNProvider(mock, nil, "auto")
	v := VPN{ConfigPath: "/etc/openvpn/client/work.conf"}

	status, err := p.Status(context.Background(), v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != StatusConnected {
		t.Errorf("auto-detect should use systemd, got status %v", status)
	}
}
