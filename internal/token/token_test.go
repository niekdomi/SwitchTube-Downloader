package token

import (
	"os"
	"os/user"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func setupTestKeyring(t *testing.T, setupToken bool, tokenValue string) *Manager {
	t.Helper()
	keyring.MockInit()

	tokenMgr := NewTokenManager()

	if setupToken {
		currentUser, err := user.Current()
		require.NoError(t, err)
		err = keyring.Set(serviceName, currentUser.Username, tokenValue)
		require.NoError(t, err)
	}

	return tokenMgr
}

func setupTestInput(t *testing.T, input string) func() {
	t.Helper()

	tmpFile, err := os.CreateTemp(t.TempDir(), "test-input")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(input)
	require.NoError(t, err)

	_, err = tmpFile.Seek(0, 0)
	require.NoError(t, err)

	oldStdin := os.Stdin
	os.Stdin = tmpFile

	return func() {
		os.Stdin = oldStdin

		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}
}

func suppressStdout(t *testing.T) func() {
	t.Helper()

	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)

	return func() { os.Stdout = oldStdout }
}

func TestGet(t *testing.T) {
	tests := []struct {
		name        string
		setupToken  bool
		tokenValue  string
		wantToken   string
		wantErrType error
	}{
		{
			name:        "token retrieval",
			setupToken:  true,
			tokenValue:  "test-token-123",
			wantErrType: errTokenInvalid,
		},
		{
			name:        "token not found",
			setupToken:  false,
			wantErrType: errNoToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenMgr := setupTestKeyring(t, tt.setupToken, tt.tokenValue)

			token, err := tokenMgr.Get()

			if tt.wantErrType != nil {
				assert.ErrorIs(t, err, tt.wantErrType)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantToken, token)
			}
		})
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name          string
		existingToken bool
		input         string
		wantErr       bool
		wantErrType   error
	}{
		{
			name:        "new token creation",
			input:       "abcdefghijklmnopqrstuvwxyz0123456789-_ABCDE\n",
			wantErr:     true,
			wantErrType: errTokenInvalid,
		},
		{
			name:          "replace existing token",
			existingToken: true,
			input:         "n\n",
			wantErr:       true,
			wantErrType:   ErrTokenAlreadyExists,
		},
		{
			name:        "empty token input",
			input:       "\n",
			wantErr:     true,
			wantErrType: errTokenEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := suppressStdout(t)
			defer restore()

			tokenMgr := setupTestKeyring(t, tt.existingToken, "existing-token")

			restoreInput := setupTestInput(t, tt.input)
			defer restoreInput()

			err := tokenMgr.Set()
			if tt.wantErr {
				require.Error(t, err)

				if tt.wantErrType != nil {
					require.ErrorIs(t, err, tt.wantErrType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name       string
		setupToken bool
		input      string
		wantErr    bool
	}{
		{
			name:       "successful deletion",
			setupToken: true,
			input:      "y\n",
			wantErr:    false,
		},
		{
			name:       "token not found",
			setupToken: false,
			input:      "y\n",
			wantErr:    true,
		},
		{
			name:       "deletion cancelled",
			setupToken: true,
			input:      "n\n",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := suppressStdout(t)
			defer restore()

			tokenMgr := setupTestKeyring(t, tt.setupToken, "test-token")

			restoreInput := setupTestInput(t, tt.input)
			defer restoreInput()

			err := tokenMgr.Delete()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "no token found")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name       string
		setupToken bool
		tokenValue string
		wantErr    bool
	}{
		{
			name:       "token validation",
			setupToken: true,
			tokenValue: "valid-token",
			wantErr:    true,
		},
		{
			name:       "invalid token",
			setupToken: true,
			tokenValue: "invalid",
			wantErr:    true,
		},
		{
			name:    "no token",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := suppressStdout(t)
			defer restore()

			tokenMgr := setupTestKeyring(t, tt.setupToken, tt.tokenValue)

			err := tokenMgr.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMaskToken(t *testing.T) {
	tm := NewTokenManager()
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "short token",
			token:    "abc",
			expected: "***",
		},
		{
			name:     "medium token",
			token:    "abcdefghij",
			expected: "**********",
		},
		{
			name:     "long token",
			token:    "abcdefghijklmnopqrstuvwxyz0123456789-_ABCDE",
			expected: "abcde*********************************ABCDE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tm.maskToken(tt.token)
			assert.Equal(t, tt.expected, result)
		})
	}
}
