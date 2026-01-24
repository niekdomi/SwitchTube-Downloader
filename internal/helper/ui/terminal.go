package ui

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/term"
)

var (
	errFailedToSetRawMode           = errors.New("failed to set raw mode")
	errFailedToRestoreTerminalState = errors.New("failed to restore terminal state")
)

// TerminalState stores the original terminal state for restoration.
type TerminalState struct {
	fd    int
	state *term.State
}

// EnableRawMode switches the terminal to raw mode for interactive input.
// Returns the original state that should be restored later.
func EnableRawMode() (*TerminalState, error) {
	fd := int(os.Stdin.Fd())

	// Save original state
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToSetRawMode, err)
	}

	return &TerminalState{
		fd:    fd,
		state: oldState,
	}, nil
}

// Restore returns the terminal to its original state.
func (ts *TerminalState) Restore() error {
	if ts.state == nil {
		return nil
	}

	if err := term.Restore(ts.fd, ts.state); err != nil {
		return fmt.Errorf("%w: %w", errFailedToRestoreTerminalState, err)
	}

	return nil
}

// IsTerminal checks if stdin is a terminal (not piped input).
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
