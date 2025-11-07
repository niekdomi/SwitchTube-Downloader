package ui

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInput(t *testing.T) {
	tests := []struct {
		name   string
		prompt string
		input  string
		want   string
	}{
		{
			name:   "basic input",
			prompt: "Enter value: ",
			input:  "test-value\n",
			want:   "test-value",
		},
		{
			name:   "empty input",
			prompt: "Enter value: ",
			input:  "\n",
			want:   "",
		},
		{
			name:   "input with leading/trailing spaces",
			prompt: "Enter value: ",
			input:  "  spaced-value  \n",
			want:   "spaced-value",
		},
		{
			name:   "multiline input stops at first newline",
			prompt: "Enter value: ",
			input:  "first-line\nsecond-line\n",
			want:   "first-line",
		},
		{
			name:   "input with tabs",
			prompt: "Enter value: ",
			input:  "\tvalue-with-tabs\t\n",
			want:   "value-with-tabs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore, readOutput := SetupTestIO(t, tt.input)
			defer restore()

			result := Input(tt.prompt)
			capturedOutput := readOutput()

			assert.Equal(t, tt.want, result)
			assert.Contains(t, capturedOutput, tt.prompt)
		})
	}
}

func TestConfirm(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []any
		input  string
		want   bool
	}{
		{
			name:   "yes lowercase",
			format: "Continue?",
			input:  "y\n",
			want:   true,
		},
		{
			name:   "yes uppercase",
			format: "Continue?",
			input:  "Y\n",
			want:   true,
		},
		{
			name:   "yes full word lowercase",
			format: "Continue?",
			input:  "yes\n",
			want:   true,
		},
		{
			name:   "yes full word uppercase",
			format: "Continue?",
			input:  "YES\n",
			want:   true,
		},
		{
			name:   "yes with spaces",
			format: "Continue?",
			input:  "  y  \n",
			want:   true,
		},
		{
			name:   "no lowercase",
			format: "Continue?",
			input:  "n\n",
			want:   false,
		},
		{
			name:   "no uppercase",
			format: "Continue?",
			input:  "N\n",
			want:   false,
		},
		{
			name:   "no full word",
			format: "Continue?",
			input:  "no\n",
			want:   false,
		},
		{
			name:   "empty input defaults to no",
			format: "Continue?",
			input:  "\n",
			want:   false,
		},
		{
			name:   "invalid input defaults to no",
			format: "Continue?",
			input:  "💀\n",
			want:   false,
		},
		{
			name:   "format with single argument",
			format: "Delete file %s?",
			args:   []any{"test.txt"},
			input:  "y\n",
			want:   true,
		},
		{
			name:   "format with multiple arguments",
			format: "Delete %d files from %s?",
			args:   []any{5, "/tmp"},
			input:  "n\n",
			want:   false,
		},
		{
			name:   "format with no arguments but percent sign",
			format: "Continue with 100%% certainty?",
			input:  "yes\n",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore, readOutput := SetupTestIO(t, tt.input)
			defer restore()

			var result bool
			if tt.args != nil {
				result = Confirm(tt.format, tt.args...)
			} else {
				result = Confirm("%s", tt.format)
			}

			capturedOutput := readOutput()

			assert.Equal(t, tt.want, result)

			var expectedBase string
			if tt.args != nil {
				expectedBase = fmt.Sprintf(tt.format, tt.args...)
			} else {
				expectedBase = tt.format
			}

			expectedPrompt := expectedBase + " (y/N): "
			assert.Contains(t, capturedOutput, expectedPrompt)
		})
	}
}

func TestConfirmPromptFormat(t *testing.T) {
	restore, readOutput := SetupTestIO(t, "n\n")
	defer restore()

	Confirm("Test prompt")

	capturedOutput := readOutput()

	expectedPrompt := "Test prompt (y/N): "
	assert.Equal(t, expectedPrompt, capturedOutput)
}

func TestInputEmptyPrompt(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "test-input")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("test\n")
	require.NoError(t, err)

	_, err = tmpFile.Seek(0, 0)
	require.NoError(t, err)

	oldStdin := os.Stdin
	os.Stdin = tmpFile

	defer func() { os.Stdin = oldStdin }()

	result := Input("")

	assert.Equal(t, "test", result)
}
