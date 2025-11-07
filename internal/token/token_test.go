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
			input:   "abcdefghijklmnopqrstuvwxyz0123456789-_ABCDE\n",
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

func TestValidateToken(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   "abcdefghijklmnopqrstuvwxyz0123456789-_ABCDE",
			wantErr: false,
		},
		{
			name:    "too short",
			token:   "short",
			wantErr: true,
		},
		{
			name:    "too long",
			token:   "abcdefghijklmnopqrstuvwxyz0123456789-_ABCDEF",
			wantErr: true,
		},
		{
			name:    "contains space",
			token:   "abcdefghijklmnopqrstuvwxyz0123456789- _ABCD",
			wantErr: true,
		},
		{
			name:    "contains special char",
			token:   "abcdefghijklmnopqrstuvwxyz0123456789💀@_ABCD",
			wantErr: true,
		},
		{
			name:    "contains tabs",
			token:   "abcdefghijklmnopqrstuvwxyz0123456789-\tABCDE",
			wantErr: true,
		},
		{
			name:    "empty string",
			token:   "",
			wantErr: true,
		},
		{
			name:    "all numbers",
			token:   "0123456789012345678901234567890123456789012",
			wantErr: false,
		},
		{
			name:    "all letters",
			token:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQ",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateToken(tt.token)
			if tt.wantErr {
				assert.ErrorIs(t, err, errInvalidToken)
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
			input:     "abcdefghijklmnopqrstuvwxyz0123456789-_ABCDE\n",
			wantToken: "abcdefghijklmnopqrstuvwxyz0123456789-_ABCDE",
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
			input:     "  abcdefghijklmnopqrstuvwxyz0123456789-_ABCDE  \n",
			wantToken: "abcdefghijklmnopqrstuvwxyz0123456789-_ABCDE",
		},
		{
			name:      "token with newlines",
			input:     "abcdefghijklmnopqrstuvwxyz0123456789-_ABCDE\nwith\nnewlines\n",
			wantToken: "abcdefghijklmnopqrstuvwxyz0123456789-_ABCDE",
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
