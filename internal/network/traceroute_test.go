package network

import (
	"context"
	"testing"
	"time"

	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
)

func TestTraceroute(t *testing.T) {
	output := `traceroute to 8.8.8.8 (8.8.8.8), 30 hops max, 60 byte packets
 1  192.168.1.1  1.234 ms
 2  10.0.0.1  5.678 ms
 3  *
 4  8.8.8.8  12.345 ms`

	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		if name != "traceroute" {
			t.Fatalf("unexpected command: %s", name)
		}
		return system.ExecResult{Stdout: output, ExitCode: 0}, nil
	})

	hops, err := Traceroute(context.Background(), mock, "8.8.8.8", "eth0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hops) != 4 {
		t.Fatalf("expected 4 hops, got %d", len(hops))
	}

	// Hop 1
	if hops[0].Number != 1 {
		t.Errorf("hop 0 number: got %d, want 1", hops[0].Number)
	}
	if hops[0].IP != "192.168.1.1" {
		t.Errorf("hop 0 IP: got %q, want 192.168.1.1", hops[0].IP)
	}
	expected := 1234 * time.Microsecond
	if hops[0].Latency != expected {
		t.Errorf("hop 0 latency: got %v, want %v", hops[0].Latency, expected)
	}

	// Hop 3 (timeout)
	if hops[2].IP != "*" {
		t.Errorf("hop 2 IP: got %q, want *", hops[2].IP)
	}
	if hops[2].Latency != 0 {
		t.Errorf("hop 2 latency: got %v, want 0", hops[2].Latency)
	}

	// Hop 4
	if hops[3].IP != "8.8.8.8" {
		t.Errorf("hop 3 IP: got %q, want 8.8.8.8", hops[3].IP)
	}
}

func TestTracerouteNoInterface(t *testing.T) {
	output := `traceroute to 8.8.8.8 (8.8.8.8), 30 hops max, 60 byte packets
 1  192.168.1.1  1.234 ms`

	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		// When iface is empty, -i should NOT be present
		for _, a := range args {
			if a == "-i" {
				t.Error("unexpected -i flag when iface is empty")
			}
		}
		return system.ExecResult{Stdout: output, ExitCode: 0}, nil
	})

	hops, err := Traceroute(context.Background(), mock, "8.8.8.8", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hops) != 1 {
		t.Fatalf("expected 1 hop, got %d", len(hops))
	}
}

func TestParseHopLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantOK  bool
		number  int
		ip      string
		latency time.Duration
	}{
		{
			name:    "normal hop",
			line:    " 1  192.168.1.1  1.234 ms",
			wantOK:  true,
			number:  1,
			ip:      "192.168.1.1",
			latency: 1234 * time.Microsecond,
		},
		{
			name:   "timeout hop",
			line:   " 3  *",
			wantOK: true,
			number: 3,
			ip:     "*",
		},
		{
			name:   "empty line",
			line:   "",
			wantOK: false,
		},
		{
			name:   "non-numeric start",
			line:   "traceroute to 8.8.8.8",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hop, ok := parseHopLine(tt.line)
			if ok != tt.wantOK {
				t.Fatalf("ok: got %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if hop.Number != tt.number {
				t.Errorf("number: got %d, want %d", hop.Number, tt.number)
			}
			if hop.IP != tt.ip {
				t.Errorf("ip: got %q, want %q", hop.IP, tt.ip)
			}
			if hop.Latency != tt.latency {
				t.Errorf("latency: got %v, want %v", hop.Latency, tt.latency)
			}
		})
	}
}
