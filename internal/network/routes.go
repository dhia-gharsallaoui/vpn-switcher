package network

import (
	"context"
	"fmt"
	"strings"

	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
)

// Route represents a single entry from the kernel routing table.
type Route struct {
	Dest   string
	Via    string
	Dev    string
	Scope  string
	Metric string
	Proto  string
}

// ListRoutes runs "ip route show table <tableName>" and parses the output
// into a slice of Route structs.
func ListRoutes(ctx context.Context, executor system.CommandExecutor, tableName string) ([]Route, error) {
	result, err := executor.Run(ctx, "ip", "route", "show", "table", tableName)
	if err != nil {
		return nil, fmt.Errorf("list routes table %s: %w", tableName, err)
	}
	if result.ExitCode != 0 {
		return nil, fmt.Errorf("list routes table %s: %s", tableName, result.Stderr)
	}

	var routes []Route
	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		r := parseRouteLine(line)
		routes = append(routes, r)
	}
	return routes, nil
}

// parseRouteLine parses a single line of "ip route show" output.
// Examples:
//
//	default via 192.168.1.1 dev eth0 proto dhcp metric 100
//	10.0.0.0/8 dev tun0 scope link
//	192.168.1.0/24 dev eth0 proto kernel scope link src 192.168.1.50 metric 100
func parseRouteLine(line string) Route {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return Route{}
	}

	r := Route{Dest: fields[0]}

	for i := 1; i < len(fields)-1; i++ {
		switch fields[i] {
		case "via":
			r.Via = fields[i+1]
			i++
		case "dev":
			r.Dev = fields[i+1]
			i++
		case "scope":
			r.Scope = fields[i+1]
			i++
		case "metric":
			r.Metric = fields[i+1]
			i++
		case "proto":
			r.Proto = fields[i+1]
			i++
		}
	}

	return r
}
