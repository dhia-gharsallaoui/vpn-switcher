package network

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIPFetcher_FetchPublicIP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("185.1.2.3"))
	}))
	defer server.Close()

	fetcher := NewIPFetcher(server.URL)
	ip, err := fetcher.FetchPublicIP(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "185.1.2.3" {
		t.Errorf("got IP %q, want %q", ip, "185.1.2.3")
	}
}

func TestIPFetcher_TrimsWhitespace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("  10.0.0.1\n  "))
	}))
	defer server.Close()

	fetcher := NewIPFetcher(server.URL)
	ip, err := fetcher.FetchPublicIP(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "10.0.0.1" {
		t.Errorf("got IP %q, want %q", ip, "10.0.0.1")
	}
}

func TestIPFetcher_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	fetcher := NewIPFetcher(server.URL)
	_, err := fetcher.FetchPublicIP(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestIPFetcher_DefaultURL(t *testing.T) {
	fetcher := NewIPFetcher("")
	if fetcher.url != "https://api.ipify.org" {
		t.Errorf("got URL %q, want default", fetcher.url)
	}
}
