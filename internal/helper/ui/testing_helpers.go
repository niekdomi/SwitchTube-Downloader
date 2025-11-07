package ui

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const outputBufferSize = 1000

// SetupTestIO configures stdin and stdout for testing and returns cleanup and output capture functions.
func SetupTestIO(t *testing.T, input string) (func(), func() string) {
	t.Helper()

	tmpFile, err := os.CreateTemp(t.TempDir(), "test-input")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(input)
	require.NoError(t, err)

	_, err = tmpFile.Seek(0, 0)
	require.NoError(t, err)

	oldStdin := os.Stdin
	os.Stdin = tmpFile

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	restore := func() {
		_ = w.Close()

		os.Stdin = oldStdin
		os.Stdout = oldStdout

		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}

	readOutput := func() string {
		output := make([]byte, outputBufferSize)
		n, _ := r.Read(output)

		return string(output[:n])
	}

	return restore, readOutput
}
