// Package input provides functionality to read keyboard input events.
package input

import (
	"errors"
	"fmt"
	"os"
)

var errFailedToReadKey = errors.New("failed to read key")

// Key represents different types of keyboard input.
type Key uint8

// Define key constants.
const (
	KeyArrowUp Key = iota
	KeyArrowDown
	KeySpace
	KeyEnter
	KeyCtrlC
	KeyUnknown
)

var singleByteKeys = map[byte]Key{
	'\r': KeyEnter,
	'\n': KeyEnter,
	' ':  KeySpace,
	3:    KeyCtrlC, // Ctrl+C
	'j':  KeyArrowDown,
	'k':  KeyArrowUp,
}

var escapeSequences = map[byte]Key{
	'A': KeyArrowUp,
	'B': KeyArrowDown,
}

// ReadKey reads a single key press from stdin.
// Returns the key pressed and error if reading fails.
func ReadKey() (Key, error) {
	buf := make([]byte, 3)

	n, err := os.Stdin.Read(buf)
	if err != nil {
		return KeyUnknown, fmt.Errorf("%w: %w", errFailedToReadKey, err)
	}

	if n == 0 {
		return KeyUnknown, nil
	}

	if key, exists := singleByteKeys[buf[0]]; exists {
		return key, nil
	}

	// Handle ANSI escape sequences
	if buf[0] == '\033' && n >= 3 && buf[1] == '[' {
		if key, exists := escapeSequences[buf[2]]; exists {
			return key, nil
		}

		return KeyUnknown, nil
	}

	return KeyUnknown, nil
}
