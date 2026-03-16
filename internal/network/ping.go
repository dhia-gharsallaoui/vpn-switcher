package network

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
)

// PingResult holds the outcome of a single ping attempt.
type PingResult struct {
	Interface string
	Target    string
	Latency   time.Duration
	Success   bool
}

// PingTarget sends a single ICMP ping to target, optionally bound to an interface.
// If ifaceName is empty, no interface binding is used.
func PingTarget(ctx context.Context, executor system.CommandExecutor, target, ifaceName string) (PingResult, error) {
	args := []string{"-c", "1", "-W", "2"}
	if ifaceName != "" {
		args = append(args, "-I", ifaceName)
	}
	args = append(args, target)

	result, err := executor.Run(ctx, "ping", args...)
	if err != nil {
		return PingResult{Interface: ifaceName, Target: target}, fmt.Errorf("ping %s: %w", target, err)
	}

	pr := PingResult{
		Interface: ifaceName,
		Target:    target,
		Success:   result.ExitCode == 0,
	}

	if pr.Success {
		pr.Latency = parseRTT(result.Stdout)
	}

	return pr, nil
}

// PingGateway sends a single ICMP ping to target via the specified interface.
// Kept for backward compatibility with tests.
func PingGateway(ctx context.Context, executor system.CommandExecutor, ifaceName, target string) (PingResult, error) {
	return PingTarget(ctx, executor, target, ifaceName)
}

// DetectGateway returns the default gateway for a given interface.
// It runs: ip route show dev <iface> | grep default
func DetectGateway(ctx context.Context, executor system.CommandExecutor, ifaceName string) (string, error) {
	result, err := executor.Run(ctx, "ip", "route", "show", "dev", ifaceName)
	if err != nil {
		return "", fmt.Errorf("detect gateway for %s: %w", ifaceName, err)
	}

	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "default via ") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				return fields[2], nil
			}
		}
	}

	// No default gateway — try the first route's network as a ping target
	for _, line := range strings.Split(result.Stdout, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) >= 1 && fields[0] != "" {
			// Return the subnet (e.g., "10.8.0.0/24") — caller should extract usable IP
			return fields[0], nil
		}
	}

	return "", fmt.Errorf("no gateway found for %s", ifaceName)
}

// parseRTT extracts the round-trip time from ping output.
func parseRTT(output string) time.Duration {
	for _, line := range strings.Split(output, "\n") {
		// Try the per-packet "time=X.Y ms" format
		if idx := strings.Index(line, "time="); idx >= 0 {
			s := line[idx+5:]
			end := strings.IndexAny(s, " \t")
			if end > 0 {
				s = s[:end]
			}
			s = strings.TrimSuffix(s, "ms")
			s = strings.TrimSpace(s)
			if ms, err := strconv.ParseFloat(s, 64); err == nil {
				return time.Duration(ms*1000) * time.Microsecond
			}
		}

		// Try the summary line: "rtt min/avg/max/mdev = 1.234/..."
		if strings.HasPrefix(strings.TrimSpace(line), "rtt") || strings.HasPrefix(strings.TrimSpace(line), "round-trip") {
			if idx := strings.Index(line, "="); idx >= 0 {
				s := strings.TrimSpace(line[idx+1:])
				s = strings.TrimSuffix(s, "ms")
				s = strings.TrimSpace(s)
				parts := strings.Split(s, "/")
				if len(parts) >= 2 {
					if ms, err := strconv.ParseFloat(parts[1], 64); err == nil {
						return time.Duration(ms*1000) * time.Microsecond
					}
				}
			}
		}
	}
	return 0
}
