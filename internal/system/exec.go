package system

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// ExecResult holds the output of a command execution.
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// CommandExecutor abstracts all system command execution.
// This is the ONLY way the application runs external commands.
type CommandExecutor interface {
	Run(ctx context.Context, name string, args ...string) (ExecResult, error)
	RunWithStdin(ctx context.Context, stdin string, name string, args ...string) (ExecResult, error)
}

// allowedCommands is the set of binaries the executor is permitted to run.
var allowedCommands = map[string]bool{
	"systemctl":  true,
	"tailscale":  true,
	"ip":         true,
	"openvpn":    true,
	"nmcli":      true,
	"curl":       true,
	"cat":        true,
	"tee":        true,
	"ping":       true,
	"dig":        true,
	"traceroute": true,
	"resolvectl": true,
}

// RealExecutor implements CommandExecutor using os/exec.
type RealExecutor struct{}

// NewRealExecutor creates a new RealExecutor.
func NewRealExecutor() *RealExecutor {
	return &RealExecutor{}
}

// Run executes a command and waits for completion.
func (e *RealExecutor) Run(ctx context.Context, name string, args ...string) (ExecResult, error) {
	if !allowedCommands[name] {
		return ExecResult{}, fmt.Errorf("%w: %s", ErrCommandNotAllowed, name)
	}

	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := ExecResult{
		Stdout: strings.TrimRight(stdout.String(), "\n"),
		Stderr: strings.TrimRight(stderr.String(), "\n"),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			return result, nil
		}
		if ctx.Err() != nil {
			return result, fmt.Errorf("%w: %s %v", ErrCommandTimeout, name, args)
		}
		return result, fmt.Errorf("exec %s: %w", name, err)
	}

	return result, nil
}

// RunWithStdin executes a command with stdin input.
func (e *RealExecutor) RunWithStdin(ctx context.Context, stdin string, name string, args ...string) (ExecResult, error) {
	if !allowedCommands[name] {
		return ExecResult{}, fmt.Errorf("%w: %s", ErrCommandNotAllowed, name)
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = strings.NewReader(stdin)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := ExecResult{
		Stdout: strings.TrimRight(stdout.String(), "\n"),
		Stderr: strings.TrimRight(stderr.String(), "\n"),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			return result, nil
		}
		if ctx.Err() != nil {
			return result, fmt.Errorf("%w: %s %v", ErrCommandTimeout, name, args)
		}
		return result, fmt.Errorf("exec %s: %w", name, err)
	}

	return result, nil
}
