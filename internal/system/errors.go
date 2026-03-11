package system

import "errors"

var (
	// ErrCommandNotAllowed is returned when a command is not in the allowlist.
	ErrCommandNotAllowed = errors.New("command not in allowlist")

	// ErrCommandTimeout is returned when a command exceeds its context deadline.
	ErrCommandTimeout = errors.New("command timed out")
)
