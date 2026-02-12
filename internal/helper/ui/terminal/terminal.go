// Package terminal provides utilities for managing terminal modes and states.
package terminal

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/term"
)

var (
	// ErrFailedToSetRawMode is returned when the terminal cannot be set to raw mode.
	ErrFailedToSetRawMode = errors.New("failed to set raw mode")

	errFailedToRestoreTerminalState = errors.New("failed to restore terminal state")
)

// State stores the original terminal state for restoration.
type State struct {
	state *term.State // Saved terminal state for restoration
	fd    int         // File descriptor of the terminal
}

// EnableRawMode switches the terminal to raw mode for interactive input.
// Returns the original state that should be restored later.
func EnableRawMode() (*State, error) {
	fd := int(os.Stdin.Fd())

	// Save original state
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToSetRawMode, err)
	}

	return &State{
		fd:    fd,
		state: oldState,
	}, nil
}

// Restore returns the terminal to its original state.
func (ts *State) Restore() error {
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
