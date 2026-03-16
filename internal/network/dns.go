package network

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
)

// DNSResult represents a single DNS answer record.
type DNSResult struct {
	Domain       string
	Type         string
	Value        string
	TTL          int
	ResponseTime time.Duration
	Server       string
}

// Lookup runs "dig +noall +answer +stats <domain> <recordType>" and parses
// the answer section and query time from the output.
func Lookup(ctx context.Context, executor system.CommandExecutor, domain, recordType string) ([]DNSResult, error) {
	result, err := executor.Run(ctx, "dig", "+noall", "+answer", "+stats", domain, recordType)
	if err != nil {
		return nil, fmt.Errorf("dns lookup %s %s: %w", domain, recordType, err)
	}
	if result.ExitCode != 0 {
		return nil, fmt.Errorf("dns lookup %s %s: %s", domain, recordType, result.Stderr)
	}

	var results []DNSResult
	var responseTime time.Duration
	var server string

	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse query time line: ";; Query time: 12 msec"
		if strings.HasPrefix(line, ";; Query time:") {
			responseTime = parseQueryTime(line)
			continue
		}

		// Parse server line: ";; SERVER: 127.0.0.53#53(127.0.0.53)"
		if strings.HasPrefix(line, ";; SERVER:") {
			server = parseServer(line)
			continue
		}

		// Skip other comment lines
		if strings.HasPrefix(line, ";;") {
			continue
		}

		// Parse answer lines: "example.com.  300  IN  A  93.184.216.34"
		r, ok := parseDigAnswer(line)
		if ok {
			results = append(results, r)
		}
	}

	// Apply query-level metadata to all results
	for i := range results {
		results[i].ResponseTime = responseTime
		results[i].Server = server
	}

	return results, nil
}

// parseDigAnswer parses a dig answer section line.
// Format: <name> <ttl> <class> <type> <value>
func parseDigAnswer(line string) (DNSResult, bool) {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return DNSResult{}, false
	}

	ttl, err := strconv.Atoi(fields[1])
	if err != nil {
		return DNSResult{}, false
	}

	// The value may contain spaces (e.g., MX records: "10 mail.example.com.")
	value := strings.Join(fields[4:], " ")

	return DNSResult{
		Domain: strings.TrimSuffix(fields[0], "."),
		TTL:    ttl,
		Type:   fields[3],
		Value:  value,
	}, true
}

// parseQueryTime extracts the duration from a dig stats line.
// Example: ";; Query time: 12 msec"
func parseQueryTime(line string) time.Duration {
	// Remove the prefix
	s := strings.TrimPrefix(line, ";; Query time:")
	s = strings.TrimSpace(s)

	fields := strings.Fields(s)
	if len(fields) < 2 {
		return 0
	}

	ms, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0
	}
	return time.Duration(ms) * time.Millisecond
}

// parseServer extracts the server address from a dig stats line.
// Example: ";; SERVER: 127.0.0.53#53(127.0.0.53)"
func parseServer(line string) string {
	s := strings.TrimPrefix(line, ";; SERVER:")
	s = strings.TrimSpace(s)
	// Take the part before #
	if idx := strings.Index(s, "#"); idx > 0 {
		return s[:idx]
	}
	return s
}

// NSlookup runs "nslookup <domain>" and parses the output.
// Useful as a fallback when dig is not installed.
func NSlookup(ctx context.Context, executor system.CommandExecutor, domain string) ([]DNSResult, error) {
	result, err := executor.Run(ctx, "nslookup", domain)
	if err != nil {
		return nil, fmt.Errorf("nslookup %s: %w", domain, err)
	}

	var results []DNSResult
	var server string
	inAnswer := false

	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)

		// Parse the server line: "Server:  127.0.0.53"
		if strings.HasPrefix(line, "Server:") {
			server = strings.TrimSpace(strings.TrimPrefix(line, "Server:"))
			continue
		}

		// The "Non-authoritative answer:" line marks start of results
		if strings.Contains(line, "answer:") {
			inAnswer = true
			continue
		}

		if !inAnswer {
			continue
		}

		// Skip "Name:" lines
		if strings.HasPrefix(line, "Name:") {
			continue
		}

		// Parse "Address: 93.184.216.34" lines
		if strings.HasPrefix(line, "Address:") {
			addr := strings.TrimSpace(strings.TrimPrefix(line, "Address:"))
			if addr == "" {
				continue
			}
			recType := "A"
			if strings.Contains(addr, ":") {
				recType = "AAAA"
			}
			results = append(results, DNSResult{
				Domain: domain,
				Type:   recType,
				Value:  addr,
				Server: server,
			})
		}
	}

	if len(results) == 0 && result.ExitCode != 0 {
		return nil, fmt.Errorf("nslookup %s: %s", domain, result.Stderr)
	}

	return results, nil
}

// LookupAuto tries dig first, falls back to nslookup if dig is not available.
func LookupAuto(ctx context.Context, executor system.CommandExecutor, domain, recordType string) ([]DNSResult, error) {
	results, err := Lookup(ctx, executor, domain, recordType)
	if err == nil {
		return results, nil
	}

	// If dig failed (likely not installed), try nslookup
	return NSlookup(ctx, executor, domain)
}
