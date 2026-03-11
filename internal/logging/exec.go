package logging

import (
	"context"
	"log/slog"
	"strings"

	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
)

// LoggingExecutor wraps a CommandExecutor and logs all commands.
type LoggingExecutor struct {
	inner  system.CommandExecutor
	logger *slog.Logger
}

// NewLoggingExecutor wraps an executor with debug logging.
func NewLoggingExecutor(inner system.CommandExecutor, logger *slog.Logger) *LoggingExecutor {
	return &LoggingExecutor{inner: inner, logger: logger}
}

func (e *LoggingExecutor) Run(ctx context.Context, name string, args ...string) (system.ExecResult, error) {
	e.logger.Debug("exec", "cmd", name, "args", strings.Join(args, " "))
	result, err := e.inner.Run(ctx, name, args...)
	e.logger.Debug("exec result",
		"cmd", name,
		"exit", result.ExitCode,
		"stdout_len", len(result.Stdout),
		"stderr_len", len(result.Stderr),
		"err", err,
	)
	return result, err
}

func (e *LoggingExecutor) RunWithStdin(ctx context.Context, stdin string, name string, args ...string) (system.ExecResult, error) {
	e.logger.Debug("exec+stdin", "cmd", name, "args", strings.Join(args, " "), "stdin_len", len(stdin))
	result, err := e.inner.RunWithStdin(ctx, stdin, name, args...)
	e.logger.Debug("exec+stdin result",
		"cmd", name,
		"exit", result.ExitCode,
		"err", err,
	)
	return result, err
}
