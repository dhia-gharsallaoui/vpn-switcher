package network

import (
	"context"
	"testing"

	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
)

func TestListRoutes(t *testing.T) {
	output := `default via 192.168.1.1 dev eth0 proto dhcp metric 100
10.0.0.0/8 dev tun0 scope link
192.168.1.0/24 dev eth0 proto kernel scope link src 192.168.1.50 metric 100`

	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		if name != "ip" {
			t.Fatalf("unexpected command: %s", name)
		}
		return system.ExecResult{Stdout: output}, nil
	})

	routes, err := ListRoutes(context.Background(), mock, "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(routes) != 3 {
		t.Fatalf("expected 3 routes, got %d", len(routes))
	}

	// Check default route
	r := routes[0]
	if r.Dest != "default" {
		t.Errorf("expected dest 'default', got %q", r.Dest)
	}
	if r.Via != "192.168.1.1" {
		t.Errorf("expected via '192.168.1.1', got %q", r.Via)
	}
	if r.Dev != "eth0" {
		t.Errorf("expected dev 'eth0', got %q", r.Dev)
	}
	if r.Proto != "dhcp" {
		t.Errorf("expected proto 'dhcp', got %q", r.Proto)
	}
	if r.Metric != "100" {
		t.Errorf("expected metric '100', got %q", r.Metric)
	}

	// Check tun0 route
	r = routes[1]
	if r.Dest != "10.0.0.0/8" {
		t.Errorf("expected dest '10.0.0.0/8', got %q", r.Dest)
	}
	if r.Dev != "tun0" {
		t.Errorf("expected dev 'tun0', got %q", r.Dev)
	}
	if r.Scope != "link" {
		t.Errorf("expected scope 'link', got %q", r.Scope)
	}
	if r.Via != "" {
		t.Errorf("expected empty via, got %q", r.Via)
	}
}

func TestListRoutesEmpty(t *testing.T) {
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		return system.ExecResult{Stdout: ""}, nil
	})

	routes, err := ListRoutes(context.Background(), mock, "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(routes) != 0 {
		t.Fatalf("expected 0 routes, got %d", len(routes))
	}
}

func TestListRoutesError(t *testing.T) {
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		return system.ExecResult{ExitCode: 1, Stderr: "table does not exist"}, nil
	})

	_, err := ListRoutes(context.Background(), mock, "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseRouteLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected Route
	}{
		{
			name: "default route",
			line: "default via 192.168.1.1 dev eth0 proto dhcp metric 100",
			expected: Route{
				Dest:   "default",
				Via:    "192.168.1.1",
				Dev:    "eth0",
				Proto:  "dhcp",
				Metric: "100",
			},
		},
		{
			name: "subnet with scope",
			line: "10.0.0.0/8 dev tun0 scope link",
			expected: Route{
				Dest:  "10.0.0.0/8",
				Dev:   "tun0",
				Scope: "link",
			},
		},
		{
			name:     "empty line",
			line:     "",
			expected: Route{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := parseRouteLine(tt.line)
			if r.Dest != tt.expected.Dest {
				t.Errorf("dest: got %q, want %q", r.Dest, tt.expected.Dest)
			}
			if r.Via != tt.expected.Via {
				t.Errorf("via: got %q, want %q", r.Via, tt.expected.Via)
			}
			if r.Dev != tt.expected.Dev {
				t.Errorf("dev: got %q, want %q", r.Dev, tt.expected.Dev)
			}
			if r.Scope != tt.expected.Scope {
				t.Errorf("scope: got %q, want %q", r.Scope, tt.expected.Scope)
			}
			if r.Metric != tt.expected.Metric {
				t.Errorf("metric: got %q, want %q", r.Metric, tt.expected.Metric)
			}
			if r.Proto != tt.expected.Proto {
				t.Errorf("proto: got %q, want %q", r.Proto, tt.expected.Proto)
			}
		})
	}
}
