package network

import "errors"

// RuleType identifies the type of routing rule.
type RuleType string

const (
	RuleTypeSubnet RuleType = "subnet"
	RuleTypeDomain RuleType = "domain"
)

// RoutingRule represents a policy routing rule.
type RoutingRule struct {
	ID           string   `yaml:"id"`
	Description  string   `yaml:"description"`
	Enabled      bool     `yaml:"enabled"`
	Type         RuleType `yaml:"type"`
	DestCIDR     string   `yaml:"dest_cidr,omitempty"`
	Domains      []string `yaml:"domains,omitempty"`
	VPNInterface string   `yaml:"vpn_interface"`
	Table        string   `yaml:"table"`
}

// InterfaceInfo represents a network interface's current state.
type InterfaceInfo struct {
	Name  string
	Up    bool
	Addrs []string
}

var (
	ErrInvalidCIDR      = errors.New("invalid CIDR notation")
	ErrTableConflict    = errors.New("routing table conflict")
	ErrPermissionDenied = errors.New("permission denied: operation requires root or sudo")
)
