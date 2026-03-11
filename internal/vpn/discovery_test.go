package vpn

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigName(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/etc/openvpn/client/work.conf", "work"},
		{"/home/user/.config/openvpn/home.ovpn", "home"},
		{"lab.conf", "lab"},
	}

	for _, tt := range tests {
		got := ConfigName(tt.path)
		if got != tt.want {
			t.Errorf("ConfigName(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestGlobConfigs(t *testing.T) {
	dir := t.TempDir()

	// Create test config files
	for _, name := range []string{"work.conf", "home.ovpn", "lab.conf", "readme.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("test"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	configs, err := GlobConfigs([]string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(configs) != 3 {
		t.Fatalf("got %d configs, want 3 (txt should be excluded): %v", len(configs), configs)
	}

	// Verify no duplicates
	seen := make(map[string]bool)
	for _, c := range configs {
		if seen[c] {
			t.Errorf("duplicate config: %s", c)
		}
		seen[c] = true
	}
}

func TestGlobConfigs_NonexistentDir(t *testing.T) {
	configs, err := GlobConfigs([]string{"/nonexistent/dir"})
	if err != nil {
		t.Fatalf("nonexistent dir should not error: %v", err)
	}
	if len(configs) != 0 {
		t.Errorf("got %d configs, want 0", len(configs))
	}
}

func TestGlobConfigs_MultipleDirs(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	os.WriteFile(filepath.Join(dir1, "a.conf"), []byte("a"), 0o644)
	os.WriteFile(filepath.Join(dir2, "b.ovpn"), []byte("b"), 0o644)

	configs, err := GlobConfigs([]string{dir1, dir2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 2 {
		t.Fatalf("got %d configs, want 2", len(configs))
	}
}
