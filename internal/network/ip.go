package network

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// IPFetcher fetches the public IP address.
type IPFetcher struct {
	url    string
	client *http.Client
}

// NewIPFetcher creates a new IP fetcher with the given URL.
func NewIPFetcher(url string) *IPFetcher {
	if url == "" {
		url = "https://api.ipify.org"
	}
	return &IPFetcher{
		url: url,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchPublicIP returns the current public IP address.
func (f *IPFetcher) FetchPublicIP(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch public IP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch public IP: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64))
	if err != nil {
		return "", fmt.Errorf("read public IP: %w", err)
	}

	return strings.TrimSpace(string(body)), nil
}
