package network

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
)

// PingResult holds the outcome of a single ping attempt through an interface.
type PingResult struct {
	Interface string
	Target    string
	Latency   time.Duration
	Success   bool
}

// PingGateway sends a single ICMP ping to target via the specified interface.
// It runs: ping -c 1 -W 2 -I <ifaceName> <target>
func PingGateway(ctx context.Context, executor system.CommandExecutor, ifaceName, target string) (PingResult, error) {
	result, err := executor.Run(ctx, "ping", "-c", "1", "-W", "2", "-I", ifaceName, target)
	if err != nil {
		return PingResult{Interface: ifaceName, Target: target}, fmt.Errorf("ping %s via %s: %w", target, ifaceName, err)
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

// parseRTT extracts the round-trip time from ping output.
// It looks for the "time=X.Y ms" pattern in the response line, or
// parses the rtt summary line "rtt min/avg/max/mdev = 1.2/1.2/1.2/0.0 ms".
func parseRTT(output string) time.Duration {
	for _, line := range strings.Split(output, "\n") {
		// Try the per-packet "time=X.Y ms" format
		if idx := strings.Index(line, "time="); idx >= 0 {
			s := line[idx+5:]
			// Find the end of the numeric value
			end := strings.IndexAny(s, " \t")
			if end > 0 {
				s = s[:end]
			}
			// Remove "ms" suffix if present
			s = strings.TrimSuffix(s, "ms")
			s = strings.TrimSpace(s)
			if ms, err := strconv.ParseFloat(s, 64); err == nil {
				return time.Duration(ms*1000) * time.Microsecond
			}
		}

		// Try the summary line: "rtt min/avg/max/mdev = 1.234/1.234/1.234/0.000 ms"
		if strings.HasPrefix(strings.TrimSpace(line), "rtt") || strings.HasPrefix(strings.TrimSpace(line), "round-trip") {
			if idx := strings.Index(line, "="); idx >= 0 {
				s := strings.TrimSpace(line[idx+1:])
				s = strings.TrimSuffix(s, "ms")
				s = strings.TrimSpace(s)
				parts := strings.Split(s, "/")
				if len(parts) >= 2 {
					// Use the avg (second value)
					if ms, err := strconv.ParseFloat(parts[1], 64); err == nil {
						return time.Duration(ms*1000) * time.Microsecond
					}
				}
			}
		}
	}
	return 0
}
