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
			name:       "successful token retrieval",
			setupToken: true,
			tokenValue: "test-token-123",
			wantToken:  "test-token-123",
		},
		{
			name:        "token not found",
			setupToken:  false,
			wantErrType: errNoTokenFound,
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

func TestGet_KeyringError(t *testing.T) {
	tokenMgr := setupTestKeyring(t, false, "")

	_, err := user.Current()
	require.NoError(t, err)

	_, err = tokenMgr.Get()
	require.Error(t, err)
	assert.ErrorIs(t, err, errNoTokenFound)
}

func TestSet(t *testing.T) {
	tests := []struct {
		name          string
		existingToken bool
		input         string
		wantErr       bool
	}{
		{
			name:    "new token creation",
			input:   "new-token\n",
			wantErr: false,
		},
		{
			name:          "replace existing token - declined",
			existingToken: true,
			input:         "n\n",
			wantErr:       true,
		},
		{
			name:    "empty token input",
			input:   "\n",
			wantErr: true,
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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name        string
		setupToken  bool
		wantErrType error
	}{
		{
			name:       "successful deletion",
			setupToken: true,
		},
		{
			name:        "token not found",
			setupToken:  false,
			wantErrType: errNoTokenFoundDelete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenMgr := setupTestKeyring(t, tt.setupToken, "test-token")

			err := tokenMgr.Delete()
			if tt.wantErrType != nil {
				assert.ErrorIs(t, err, tt.wantErrType)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_GetUserError(t *testing.T) {
	tokenMgr := setupTestKeyring(t, false, "")

	_, err := tokenMgr.Get()
	require.Error(t, err)
	assert.ErrorIs(t, err, errNoTokenFound)
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantToken   string
		wantErrType error
	}{
		{
			name:      "valid token input",
			input:     "valid-token-123\n",
			wantToken: "valid-token-123",
		},
		{
			name:        "empty token input",
			input:       "\n",
			wantToken:   "",
			wantErrType: errTokenEmpty,
		},
		{
			name:        "whitespace only token",
			input:       "   \n",
			wantToken:   "",
			wantErrType: errTokenEmpty,
		},
		{
			name:      "token with leading/trailing whitespace",
			input:     "  token-with-spaces   \n",
			wantToken: "token-with-spaces",
		},
		{
			name:      "token with special characters",
			input:     "token-123_ABC.xyz\n",
			wantToken: "token-123_ABC.xyz",
		},
		{
			name:      "long token",
			input:     "very-long-token-with-many-characters-1234567890\n",
			wantToken: "very-long-token-with-many-characters-1234567890",
		},
		{
			name:      "token with newlines",
			input:     "token\nwith\nnewlines\n",
			wantToken: "token",
		},
		{
			name:      "token with tabs",
			input:     "token\twith\ttabs\n",
			wantToken: "token\twith\ttabs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := suppressStdout(t)
			defer restore()

			tokenMgr := NewTokenManager()

			restoreInput := setupTestInput(t, tt.input)
			defer restoreInput()

			token, err := tokenMgr.create()

			assert.Equal(t, tt.wantToken, token)

			if tt.wantErrType != nil {
				assert.ErrorIs(t, err, tt.wantErrType)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
