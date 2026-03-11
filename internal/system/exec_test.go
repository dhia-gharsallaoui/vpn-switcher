package system

import (
	"context"
	"errors"
	"testing"
)

func TestMockExecutor_RecordsCalls(t *testing.T) {
	mock := NewMockExecutor(func(name string, args []string) (ExecResult, error) {
		return ExecResult{Stdout: "ok", ExitCode: 0}, nil
	})

	ctx := context.Background()
	result, err := mock.Run(ctx, "tailscale", "status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stdout != "ok" {
		t.Errorf("got stdout %q, want %q", result.Stdout, "ok")
	}

	calls := mock.Calls()
	if len(calls) != 1 {
		t.Fatalf("got %d calls, want 1", len(calls))
	}
	if calls[0].Name != "tailscale" {
		t.Errorf("got name %q, want %q", calls[0].Name, "tailscale")
	}
	if len(calls[0].Args) != 1 || calls[0].Args[0] != "status" {
		t.Errorf("got args %v, want [status]", calls[0].Args)
	}
}

func TestMockExecutor_HandlerReturnsError(t *testing.T) {
	mock := NewMockExecutor(func(name string, args []string) (ExecResult, error) {
		return ExecResult{}, errors.New("command failed")
	})

	_, err := mock.Run(context.Background(), "ip", "link")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "command failed" {
		t.Errorf("got error %q, want %q", err.Error(), "command failed")
	}
}

func TestMockExecutor_NilHandler(t *testing.T) {
	mock := NewMockExecutor(nil)

	_, err := mock.Run(context.Background(), "ip", "link")
	if err == nil {
		t.Fatal("expected error for nil handler")
	}
}

func TestMockExecutor_Reset(t *testing.T) {
	mock := NewMockExecutor(func(name string, args []string) (ExecResult, error) {
		return ExecResult{}, nil
	})

	mock.Run(context.Background(), "ip", "link")
	mock.Reset()

	if len(mock.Calls()) != 0 {
		t.Errorf("expected 0 calls after reset, got %d", len(mock.Calls()))
	}
}

func TestMockExecutor_RunWithStdin(t *testing.T) {
	mock := NewMockExecutor(func(name string, args []string) (ExecResult, error) {
		return ExecResult{Stdout: "done"}, nil
	})

	result, err := mock.RunWithStdin(context.Background(), "input data", "cat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stdout != "done" {
		t.Errorf("got stdout %q, want %q", result.Stdout, "done")
	}

	calls := mock.Calls()
	if len(calls) != 1 || calls[0].Name != "cat" {
		t.Errorf("unexpected calls: %v", calls)
	}
}

func TestRealExecutor_AllowlistBlocks(t *testing.T) {
	exec := NewRealExecutor()

	_, err := exec.Run(context.Background(), "rm", "-rf", "/")
	if err == nil {
		t.Fatal("expected error for blocked command")
	}
	if !errors.Is(err, ErrCommandNotAllowed) {
		t.Errorf("got error %v, want ErrCommandNotAllowed", err)
	}
}

func TestRealExecutor_AllowlistBlocksStdin(t *testing.T) {
	exec := NewRealExecutor()

	_, err := exec.RunWithStdin(context.Background(), "", "bash", "-c", "echo pwned")
	if err == nil {
		t.Fatal("expected error for blocked command")
	}
	if !errors.Is(err, ErrCommandNotAllowed) {
		t.Errorf("got error %v, want ErrCommandNotAllowed", err)
	}
}

func TestRealExecutor_AllowedCommandRuns(t *testing.T) {
	exec := NewRealExecutor()

	// "ip" is in the allowlist and should be available on Linux
	result, err := exec.Run(context.Background(), "ip", "link", "show", "lo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d (stderr: %s)", result.ExitCode, result.Stderr)
	}
}
