package network

import (
	"context"
	"testing"
	"time"

	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
)

func TestLookup(t *testing.T) {
	output := `example.com.		300	IN	A	93.184.216.34
;; Query time: 12 msec
;; SERVER: 127.0.0.53#53(127.0.0.53)
;; WHEN: Tue Mar 11 10:00:00 UTC 2026
;; MSG SIZE  rcvd: 56`

	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		if name != "dig" {
			t.Fatalf("unexpected command: %s", name)
		}
		return system.ExecResult{Stdout: output}, nil
	})

	results, err := Lookup(context.Background(), mock, "example.com", "A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.Domain != "example.com" {
		t.Errorf("expected domain 'example.com', got %q", r.Domain)
	}
	if r.Type != "A" {
		t.Errorf("expected type 'A', got %q", r.Type)
	}
	if r.Value != "93.184.216.34" {
		t.Errorf("expected value '93.184.216.34', got %q", r.Value)
	}
	if r.TTL != 300 {
		t.Errorf("expected TTL 300, got %d", r.TTL)
	}
	if r.ResponseTime != 12*time.Millisecond {
		t.Errorf("expected response time 12ms, got %v", r.ResponseTime)
	}
	if r.Server != "127.0.0.53" {
		t.Errorf("expected server '127.0.0.53', got %q", r.Server)
	}
}

func TestLookupMultipleRecords(t *testing.T) {
	output := `example.com.		300	IN	MX	10 mail.example.com.
example.com.		300	IN	MX	20 mail2.example.com.
;; Query time: 5 msec
;; SERVER: 8.8.8.8#53(8.8.8.8)`

	mock := system.NewMockExecutor(func(name string, args []string) (system.ExecResult, error) {
		return system.ExecResult{Stdout: output}, nil
	})

	results, err := Lookup(context.Background(), mock, "example.com", "MX")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].Value != "10 mail.example.com." {
		t.Errorf("expected '10 mail.example.com.', got %q", results[0].Value)
	}
}

func TestParseDigAnswer(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantOK  bool
		domain  string
		recType string
		value   string
		ttl     int
	}{
		{
			name:    "A record",
			line:    "example.com.		300	IN	A	93.184.216.34",
			wantOK:  true,
			domain:  "example.com",
			recType: "A",
			value:   "93.184.216.34",
			ttl:     300,
		},
		{
			name:    "AAAA record",
			line:    "example.com.		300	IN	AAAA	2606:2800:220:1:248:1893:25c8:1946",
			wantOK:  true,
			domain:  "example.com",
			recType: "AAAA",
			value:   "2606:2800:220:1:248:1893:25c8:1946",
			ttl:     300,
		},
		{
			name:   "too few fields",
			line:   "example.com. 300",
			wantOK: false,
		},
		{
			name:   "invalid TTL",
			line:   "example.com. abc IN A 1.2.3.4",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, ok := parseDigAnswer(tt.line)
			if ok != tt.wantOK {
				t.Fatalf("ok: got %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if r.Domain != tt.domain {
				t.Errorf("domain: got %q, want %q", r.Domain, tt.domain)
			}
			if r.Type != tt.recType {
				t.Errorf("type: got %q, want %q", r.Type, tt.recType)
			}
			if r.Value != tt.value {
				t.Errorf("value: got %q, want %q", r.Value, tt.value)
			}
			if r.TTL != tt.ttl {
				t.Errorf("ttl: got %d, want %d", r.TTL, tt.ttl)
			}
		})
	}
}

func TestParseQueryTime(t *testing.T) {
	tests := []struct {
		line     string
		expected time.Duration
	}{
		{";; Query time: 12 msec", 12 * time.Millisecond},
		{";; Query time: 0 msec", 0},
		{";; Query time: invalid", 0},
	}

	for _, tt := range tests {
		d := parseQueryTime(tt.line)
		if d != tt.expected {
			t.Errorf("parseQueryTime(%q): got %v, want %v", tt.line, d, tt.expected)
		}
	}
}

func TestParseServer(t *testing.T) {
	tests := []struct {
		line     string
		expected string
	}{
		{";; SERVER: 127.0.0.53#53(127.0.0.53)", "127.0.0.53"},
		{";; SERVER: 8.8.8.8#53(8.8.8.8)", "8.8.8.8"},
	}

	for _, tt := range tests {
		s := parseServer(tt.line)
		if s != tt.expected {
			t.Errorf("parseServer(%q): got %q, want %q", tt.line, s, tt.expected)
		}
	}
}
