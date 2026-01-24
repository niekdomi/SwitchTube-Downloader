package ui

import (
	"fmt"
	"os"
)

// Key represents different types of keyboard input.
type Key int

// Key represents different types of keyboard input.
const (
	KeyArrowUp Key = iota
	KeyArrowDown
	KeySpace
	KeyEnter
	KeyCtrlC
	KeyChar
	KeyUnknown
)

// InputEvent represents a keyboard input event.
type InputEvent struct {
	Key  Key
	Char rune
	// Determine checkbox display
}

// ReadKey reads a single key press from stdin.
// Must be called with terminal in raw mode.
func ReadKey() (InputEvent, error) {
	buf := make([]byte, 3)

	n, err := os.Stdin.Read(buf)
	if err != nil {
		return InputEvent{Key: KeyUnknown, Char: 0}, fmt.Errorf("%w: %w", errFailedToReadKey, err)
	}

	if n == 0 {
		return InputEvent{Key: KeyUnknown, Char: 0}, nil
	}

	// Handle escape sequences (arrow keys)
	if buf[0] == '\033' {
		if n >= 3 && buf[1] == '[' { //nolint:gosec
			switch buf[2] { //nolint:gosec
			case 'A':
				return InputEvent{Key: KeyArrowUp, Char: 0}, nil
			case 'B':
				return InputEvent{Key: KeyArrowDown, Char: 0}, nil
			}
		}

		return InputEvent{Key: KeyUnknown, Char: 0}, nil
	}

	// Handle special characters
	switch buf[0] { //nolint:gosec
	case '\r', '\n':
		return InputEvent{Key: KeyEnter, Char: 0}, nil
	case ' ':
		return InputEvent{Key: KeySpace, Char: 0}, nil
	case 3: // Ctrl+C
		return InputEvent{Key: KeyCtrlC, Char: 0}, nil
	case 'j':
		return InputEvent{Key: KeyArrowDown, Char: 'j'}, nil
	case 'k':
		return InputEvent{Key: KeyArrowUp, Char: 'k'}, nil
	default:
		// Printable character
		//nolint:gosec
		if buf[0] >= 32 && buf[0] <= 126 {
			return InputEvent{Key: KeyChar, Char: rune(buf[0])}, nil
		}

		return InputEvent{Key: KeyUnknown, Char: 0}, nil
	}
}
