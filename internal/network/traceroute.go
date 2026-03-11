package network

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
)

// Hop represents a single hop in a traceroute.
type Hop struct {
	Number  int
	Host    string
	IP      string
	Latency time.Duration
}

// Traceroute runs "traceroute -i <iface> -n -q 1 -w 2 <target>" and parses
// the output into a slice of Hop structs.
func Traceroute(ctx context.Context, executor system.CommandExecutor, target, iface string) ([]Hop, error) {
	args := []string{"-n", "-q", "1", "-w", "2"}
	if iface != "" {
		args = append([]string{"-i", iface}, args...)
	}
	args = append(args, target)

	result, err := executor.Run(ctx, "traceroute", args...)
	if err != nil {
		return nil, fmt.Errorf("traceroute %s: %w", target, err)
	}
	// traceroute may exit non-zero even on partial success; we still parse output
	if result.Stdout == "" && result.ExitCode != 0 {
		return nil, fmt.Errorf("traceroute %s: %s", target, result.Stderr)
	}

	var hops []Hop
	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip the header line: "traceroute to <target> ..."
		if strings.HasPrefix(line, "traceroute to") {
			continue
		}
		hop, ok := parseHopLine(line)
		if ok {
			hops = append(hops, hop)
		}
	}
	return hops, nil
}

// parseHopLine parses a single traceroute output line.
// With -n -q 1, lines look like:
//
//	1  192.168.1.1  1.234 ms
//	2  * (timeout)
//	3  10.0.0.1  5.678 ms
func parseHopLine(line string) (Hop, bool) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return Hop{}, false
	}

	num, err := strconv.Atoi(fields[0])
	if err != nil {
		return Hop{}, false
	}

	hop := Hop{Number: num}

	if fields[1] == "*" {
		// Timeout hop
		hop.Host = "*"
		hop.IP = "*"
		return hop, true
	}

	hop.IP = fields[1]
	hop.Host = fields[1] // -n mode, so host == IP

	// Parse latency: look for a numeric value followed by "ms"
	if len(fields) >= 3 {
		latStr := fields[2]
		if ms, parseErr := strconv.ParseFloat(latStr, 64); parseErr == nil {
			hop.Latency = time.Duration(ms*1000) * time.Microsecond
		}
	}

	return hop, true
}
