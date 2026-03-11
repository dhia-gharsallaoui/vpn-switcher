package network

import (
	"context"
	"testing"

	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
)

func TestInterfaceMonitor_List(t *testing.T) {
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		if name == "ip" && args[0] == "-o" {
			return system.ExecResult{
				Stdout: `1: lo    inet 127.0.0.1/8 scope host lo
2: eth0    inet 192.168.1.100/24 brd 192.168.1.255 scope global eth0
3: tun0    inet 10.8.0.2/24 scope global tun0 UP
4: tailscale0    inet 100.64.1.4/32 scope global tailscale0 UP`,
				ExitCode: 0,
			}, nil
		}
		return system.ExecResult{}, nil
	})

	mon := NewInterfaceMonitor(mock)
	ifaces, err := mon.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ifaces) != 2 {
		t.Fatalf("got %d interfaces, want 2 (only VPN interfaces)", len(ifaces))
	}

	names := make(map[string]bool)
	for _, iface := range ifaces {
		names[iface.Name] = true
	}
	if !names["tun0"] {
		t.Error("tun0 not found")
	}
	if !names["tailscale0"] {
		t.Error("tailscale0 not found")
	}
}

func TestInterfaceMonitor_ListEmpty(t *testing.T) {
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		return system.ExecResult{
			Stdout:   "1: lo    inet 127.0.0.1/8 scope host lo",
			ExitCode: 0,
		}, nil
	})

	mon := NewInterfaceMonitor(mock)
	ifaces, err := mon.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ifaces) != 0 {
		t.Errorf("got %d interfaces, want 0", len(ifaces))
	}
}
