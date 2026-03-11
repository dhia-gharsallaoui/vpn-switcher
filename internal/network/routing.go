package network

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
)

// RoutingManager manages policy-based routing rules.
type RoutingManager struct {
	executor system.CommandExecutor
}

// NewRoutingManager creates a new routing manager.
func NewRoutingManager(exec system.CommandExecutor) *RoutingManager {
	return &RoutingManager{executor: exec}
}

// Apply installs a routing rule (ip route add + ip rule add).
func (m *RoutingManager) Apply(ctx context.Context, rule RoutingRule) error {
	if !rule.Enabled {
		return nil
	}

	if err := validateCIDR(rule.DestCIDR); err != nil {
		return err
	}

	tableNum, err := strconv.Atoi(rule.Table)
	if err != nil {
		return fmt.Errorf("invalid table number %q: %w", rule.Table, err)
	}

	// Ensure routing table exists
	if err := m.EnsureTable(ctx, tableNum, rule.ID); err != nil {
		return fmt.Errorf("ensure table %s: %w", rule.Table, err)
	}

	// Add route: ip route add <dest> dev <iface> table <table>
	result, err := m.executor.Run(ctx, "ip", "route", "add", rule.DestCIDR, "dev", rule.VPNInterface, "table", rule.Table)
	if err != nil {
		return fmt.Errorf("add route: %w", err)
	}
	if result.ExitCode != 0 {
		// Ignore "File exists" (route already present)
		if !strings.Contains(result.Stderr, "File exists") {
			if strings.Contains(result.Stderr, "Operation not permitted") {
				return fmt.Errorf("add route: %w", ErrPermissionDenied)
			}
			return fmt.Errorf("add route: %s", result.Stderr)
		}
	}

	// Add rule: ip rule add to <dest> lookup <table>
	result, err = m.executor.Run(ctx, "ip", "rule", "add", "to", rule.DestCIDR, "lookup", rule.Table)
	if err != nil {
		return fmt.Errorf("add rule: %w", err)
	}
	if result.ExitCode != 0 {
		if !strings.Contains(result.Stderr, "File exists") {
			if strings.Contains(result.Stderr, "Operation not permitted") {
				return fmt.Errorf("add rule: %w", ErrPermissionDenied)
			}
			return fmt.Errorf("add rule: %s", result.Stderr)
		}
	}

	return nil
}

// Remove deletes a routing rule.
func (m *RoutingManager) Remove(ctx context.Context, rule RoutingRule) error {
	if err := validateCIDR(rule.DestCIDR); err != nil {
		return err
	}

	// Remove rule first
	result, err := m.executor.Run(ctx, "ip", "rule", "del", "to", rule.DestCIDR, "lookup", rule.Table)
	if err != nil {
		return fmt.Errorf("del rule: %w", err)
	}
	if result.ExitCode != 0 && !strings.Contains(result.Stderr, "No such file") {
		return fmt.Errorf("del rule: %s", result.Stderr)
	}

	// Remove route
	result, err = m.executor.Run(ctx, "ip", "route", "del", rule.DestCIDR, "dev", rule.VPNInterface, "table", rule.Table)
	if err != nil {
		return fmt.Errorf("del route: %w", err)
	}
	if result.ExitCode != 0 && !strings.Contains(result.Stderr, "No such") {
		return fmt.Errorf("del route: %s", result.Stderr)
	}

	return nil
}

// ListActive returns currently active ip rules from the kernel.
func (m *RoutingManager) ListActive(ctx context.Context) ([]string, error) {
	result, err := m.executor.Run(ctx, "ip", "rule", "list")
	if err != nil {
		return nil, fmt.Errorf("list rules: %w", err)
	}

	var rules []string
	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			rules = append(rules, line)
		}
	}
	return rules, nil
}

// EnsureTable ensures a routing table entry exists in /etc/iproute2/rt_tables.
func (m *RoutingManager) EnsureTable(ctx context.Context, tableNum int, tableName string) error {
	rtTablesPath := "/etc/iproute2/rt_tables"

	result, err := m.executor.Run(ctx, "cat", rtTablesPath)
	if err != nil {
		return fmt.Errorf("read rt_tables: %w", err)
	}

	tableStr := strconv.Itoa(tableNum)

	// Check if table already exists
	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == tableStr {
			if fields[1] == tableName {
				return nil // Already exists with correct name
			}
			return fmt.Errorf("%w: table %d already defined as %q, not %q", ErrTableConflict, tableNum, fields[1], tableName)
		}
	}

	// Append new table entry via tee (respects executor security model)
	entry := fmt.Sprintf("%d\t%s\n", tableNum, tableName)
	content := result.Stdout + "\n" + entry

	writeResult, err := m.executor.RunWithStdin(ctx, content, "tee", rtTablesPath)
	if err != nil {
		return fmt.Errorf("write rt_tables: %w", err)
	}
	if writeResult.ExitCode != 0 {
		if strings.Contains(writeResult.Stderr, "Permission denied") {
			return fmt.Errorf("write rt_tables: %w", ErrPermissionDenied)
		}
		return fmt.Errorf("write rt_tables: %s", writeResult.Stderr)
	}

	return nil
}

func validateCIDR(cidr string) error {
	if cidr == "" {
		return fmt.Errorf("%w: empty CIDR", ErrInvalidCIDR)
	}
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidCIDR, cidr)
	}
	return nil
}
