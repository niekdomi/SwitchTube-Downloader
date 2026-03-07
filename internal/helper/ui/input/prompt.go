package input

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
)

// Input prompts the user for a single line of text and returns the entered string.
func Input(prompt string) string {
	var value string

	_ = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(prompt).
				Value(&value),
		),
	).Run()

	return value
}

// Confirm prompts the user for a yes/no confirmation and returns true for yes.
func Confirm(format string, args ...any) bool {
	msg := fmt.Sprintf(format, args...)
	var confirmed bool

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(msg).
				Affirmative("Yes").
				Negative("No").
				Value(&confirmed),
		),
	).Run()

	if errors.Is(err, huh.ErrUserAborted) {
		return false
	}

	return confirmed
}
