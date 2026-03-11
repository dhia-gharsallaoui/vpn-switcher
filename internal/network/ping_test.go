package network

import (
	"context"
	"testing"
	"time"

	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
)

func TestPingGatewaySuccess(t *testing.T) {
	output := `PING 8.8.8.8 (8.8.8.8) from 10.0.0.2 tun0: 56(84) bytes of data.
64 bytes from 8.8.8.8: icmp_seq=1 ttl=117 time=12.3 ms

--- 8.8.8.8 ping statistics ---
1 packets transmitted, 1 received, 0% packet loss, time 0ms
rtt min/avg/max/mdev = 12.300/12.300/12.300/0.000 ms`

	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		if name != "ping" {
			t.Fatalf("unexpected command: %s", name)
		}
		return system.ExecResult{Stdout: output, ExitCode: 0}, nil
	})

	result, err := PingGateway(context.Background(), mock, "tun0", "8.8.8.8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}
	if result.Interface != "tun0" {
		t.Errorf("expected interface 'tun0', got %q", result.Interface)
	}
	if result.Target != "8.8.8.8" {
		t.Errorf("expected target '8.8.8.8', got %q", result.Target)
	}
	// time=12.3 ms should parse to 12.3ms
	expected := 12300 * time.Microsecond
	if result.Latency != expected {
		t.Errorf("expected latency %v, got %v", expected, result.Latency)
	}
}

func TestPingGatewayTimeout(t *testing.T) {
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		return system.ExecResult{
			Stdout:   "PING 8.8.8.8 ...\n--- 8.8.8.8 ping statistics ---\n1 packets transmitted, 0 received, 100% packet loss",
			ExitCode: 1,
		}, nil
	})

	result, err := PingGateway(context.Background(), mock, "eth0", "8.8.8.8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Success {
		t.Error("expected failure")
	}
	if result.Latency != 0 {
		t.Errorf("expected zero latency, got %v", result.Latency)
	}
}

func TestParseRTT(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected time.Duration
	}{
		{
			name:     "time= format",
			output:   "64 bytes from 8.8.8.8: icmp_seq=1 ttl=117 time=12.3 ms",
			expected: 12300 * time.Microsecond,
		},
		{
			name:     "rtt summary",
			output:   "rtt min/avg/max/mdev = 1.200/1.500/1.800/0.300 ms",
			expected: 1500 * time.Microsecond,
		},
		{
			name:     "no match",
			output:   "nothing here",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := parseRTT(tt.output)
			if d != tt.expected {
				t.Errorf("got %v, want %v", d, tt.expected)
			}
		})
	}
}
