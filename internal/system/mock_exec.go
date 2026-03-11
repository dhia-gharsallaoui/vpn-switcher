package system

import (
	"context"
	"fmt"
	"sync"
)

// MockCall records a single command invocation.
type MockCall struct {
	Name string
	Args []string
}

// MockExecutor implements CommandExecutor for testing.
type MockExecutor struct {
	mu      sync.Mutex
	calls   []MockCall
	handler func(name string, args []string) (ExecResult, error)
}

// NewMockExecutor creates a mock with a handler function.
func NewMockExecutor(handler func(name string, args []string) (ExecResult, error)) *MockExecutor {
	return &MockExecutor{handler: handler}
}

// Run records the call and delegates to the handler.
func (m *MockExecutor) Run(_ context.Context, name string, args ...string) (ExecResult, error) {
	m.mu.Lock()
	m.calls = append(m.calls, MockCall{Name: name, Args: args})
	m.mu.Unlock()

	if m.handler == nil {
		return ExecResult{}, fmt.Errorf("no handler configured for: %s", name)
	}
	return m.handler(name, args)
}

// RunWithStdin records the call and delegates to the handler (stdin is ignored in mock).
func (m *MockExecutor) RunWithStdin(_ context.Context, _ string, name string, args ...string) (ExecResult, error) {
	m.mu.Lock()
	m.calls = append(m.calls, MockCall{Name: name, Args: args})
	m.mu.Unlock()

	if m.handler == nil {
		return ExecResult{}, fmt.Errorf("no handler configured for: %s", name)
	}
	return m.handler(name, args)
}

// Calls returns all recorded invocations (thread-safe).
func (m *MockExecutor) Calls() []MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]MockCall, len(m.calls))
	copy(result, m.calls)
	return result
}

// Reset clears all recorded calls.
func (m *MockExecutor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = nil
}
