package network

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
)

func TestRoutingManager_Apply(t *testing.T) {
	var commands []string
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		cmd := name + " " + strings.Join(args, " ")
		commands = append(commands, cmd)

		if name == "cat" {
			return system.ExecResult{
				Stdout: "255\tlocal\n254\tmain\n253\tdefault\n100\tcorp-network",
			}, nil
		}
		return system.ExecResult{ExitCode: 0}, nil
	})

	rm := NewRoutingManager(mock)

	rule := RoutingRule{
		ID:           "corp-network",
		Enabled:      true,
		Type:         RuleTypeSubnet,
		DestCIDR:     "10.0.0.0/8",
		VPNInterface: "tun0",
		Table:        "100",
	}

	err := rm.Apply(context.Background(), rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have: cat, ip route add, ip rule add
	var hasRouteAdd, hasRuleAdd bool
	for _, cmd := range commands {
		if strings.Contains(cmd, "route add") {
			hasRouteAdd = true
		}
		if strings.Contains(cmd, "rule add") {
			hasRuleAdd = true
		}
	}
	if !hasRouteAdd {
		t.Error("expected ip route add command")
	}
	if !hasRuleAdd {
		t.Error("expected ip rule add command")
	}
}

func TestRoutingManager_ApplyDisabled(t *testing.T) {
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		t.Fatal("no commands should be run for disabled rules")
		return system.ExecResult{}, nil
	})

	rm := NewRoutingManager(mock)
	rule := RoutingRule{Enabled: false, DestCIDR: "10.0.0.0/8"}

	err := rm.Apply(context.Background(), rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRoutingManager_ApplyInvalidCIDR(t *testing.T) {
	mock := system.NewMockExecutor(nil)
	rm := NewRoutingManager(mock)

	rule := RoutingRule{Enabled: true, DestCIDR: "not-a-cidr"}
	err := rm.Apply(context.Background(), rule)
	if err == nil {
		t.Fatal("expected error for invalid CIDR")
	}
	if !errors.Is(err, ErrInvalidCIDR) {
		t.Errorf("expected ErrInvalidCIDR, got %v", err)
	}
}

func TestRoutingManager_ApplyEmptyCIDR(t *testing.T) {
	mock := system.NewMockExecutor(nil)
	rm := NewRoutingManager(mock)

	rule := RoutingRule{Enabled: true, DestCIDR: ""}
	err := rm.Apply(context.Background(), rule)
	if !errors.Is(err, ErrInvalidCIDR) {
		t.Errorf("expected ErrInvalidCIDR for empty CIDR, got %v", err)
	}
}

func TestRoutingManager_Remove(t *testing.T) {
	var commands []string
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		commands = append(commands, name+" "+strings.Join(args, " "))
		return system.ExecResult{ExitCode: 0}, nil
	})

	rm := NewRoutingManager(mock)
	rule := RoutingRule{
		DestCIDR:     "10.0.0.0/8",
		VPNInterface: "tun0",
		Table:        "100",
	}

	err := rm.Remove(context.Background(), rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(commands) != 2 {
		t.Fatalf("got %d commands, want 2", len(commands))
	}

	// Rule should be deleted before route
	if !strings.Contains(commands[0], "rule del") {
		t.Errorf("first command should be rule del: %s", commands[0])
	}
	if !strings.Contains(commands[1], "route del") {
		t.Errorf("second command should be route del: %s", commands[1])
	}
}

func TestRoutingManager_ListActive(t *testing.T) {
	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		return system.ExecResult{
			Stdout: "0:\tfrom all lookup local\n100:\tfrom all to 10.0.0.0/8 lookup 100",
		}, nil
	})

	rm := NewRoutingManager(mock)
	rules, err := rm.ListActive(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 2 {
		t.Errorf("got %d rules, want 2", len(rules))
	}
}

func TestValidateCIDR(t *testing.T) {
	tests := []struct {
		cidr    string
		wantErr bool
	}{
		{"10.0.0.0/8", false},
		{"192.168.1.0/24", false},
		{"100.64.0.0/10", false},
		{"::1/128", false},
		{"not-cidr", true},
		{"10.0.0.0", true},
		{"", true},
	}

	for _, tt := range tests {
		err := validateCIDR(tt.cidr)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateCIDR(%q) error = %v, wantErr = %v", tt.cidr, err, tt.wantErr)
		}
	}
}
