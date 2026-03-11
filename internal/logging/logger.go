package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
)

const (
	defaultLogDir  = ".local/share/vpn-switcher"
	defaultLogFile = "debug.log"
	maxLogSize     = 10 * 1024 * 1024 // 10MB
	maxLogFiles    = 5
)

// Setup initializes structured logging. If debug is true, logs to a file;
// otherwise logging is discarded.
func Setup(debug bool) (*slog.Logger, error) {
	if !debug {
		return slog.New(slog.NewTextHandler(io.Discard, nil)), nil
	}

	logDir, err := logDirectory()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, err
	}

	logPath := filepath.Join(logDir, defaultLogFile)

	// Rotate if current log is too large
	rotateIfNeeded(logDir, logPath)

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	handler := slog.NewJSONHandler(f, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	return slog.New(handler), nil
}

func logDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, defaultLogDir), nil
}

func rotateIfNeeded(logDir, logPath string) {
	info, err := os.Stat(logPath)
	if err != nil || info.Size() < maxLogSize {
		return
	}

	// Shift existing rotated files
	for i := maxLogFiles - 1; i >= 1; i-- {
		old := filepath.Join(logDir, rotatedName(i))
		new := filepath.Join(logDir, rotatedName(i+1))
		os.Rename(old, new)
	}

	// Rotate current file
	os.Rename(logPath, filepath.Join(logDir, rotatedName(1)))
}

func rotatedName(n int) string {
	return defaultLogFile + "." + strconv.Itoa(n)
}
