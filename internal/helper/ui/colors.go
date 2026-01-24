package ui

// ANSI color codes and modifiers.
const (
	// Reset and modifiers.
	Reset = "\033[0m"
	Bold  = "\033[1m"
	Dim   = "\033[2m"

	// Foreground colors.
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"

	// Semantic aliases.
	Success = Green
	Error   = Red
	Warning = Yellow
	Info    = Cyan

	// Cursor control.
	HideCursor = "\033[?25l"
	ShowCursor = "\033[?25h"
	ClearLine  = "\033[2K"

	// Checkbox symbols.
	CheckboxChecked   = "■"
	CheckboxUnchecked = "□"
)
