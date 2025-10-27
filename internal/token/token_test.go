package token

import (
	"errors"
	"os"
	"os/user"
	"testing"

	"github.com/zalando/go-keyring"
)

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
			keyring.MockInit()

			tokenMgr := NewTokenManager()

			if tt.setupToken {
				currentUser, userErr := user.Current()
				if userErr != nil {
					t.Fatalf("Failed to get current user: %v", userErr)
				}

				keyring.Set(serviceName, currentUser.Username, tt.tokenValue)
			}

			token, err := tokenMgr.Get()

			if tt.wantErrType != nil {
				if err == nil || !errors.Is(err, tt.wantErrType) {
					t.Errorf("Get() error = %v, want error type = %v", err, tt.wantErrType)
				}
			} else {
				if err != nil {
					t.Errorf("Get() error = %v, want nil", err)
				}

				if token != tt.wantToken {
					t.Errorf("Get() token = %v, want %v", token, tt.wantToken)
				}
			}
		})
	}
}

func TestSet(t *testing.T) {
	// Capture stdout to hide prompts
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)

	defer func() { os.Stdout = oldStdout }()

	tests := []struct {
		name          string
		existingToken bool
		wantErr       bool
	}{
		{
			name:    "new token creation",
			wantErr: false,
		},
		{
			name:          "replace existing token - declined",
			existingToken: true,
			wantErr:       true,
		},
		{
			name:    "empty token input",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyring.MockInit()

			tokenMgr := NewTokenManager()

			if tt.existingToken {
				currentUser, userErr := user.Current()
				if userErr != nil {
					t.Fatalf("Failed to get current user: %v", userErr)
				}

				keyring.Set(serviceName, currentUser.Username, "existing-token")
			}

			var input string

			switch tt.name {
			case "replace existing token - declined":
				input = "n\n"
			case "empty token input":
				input = "\n"
			default:
				input = "new-token\n"
			}

			tmpFile, err := os.CreateTemp(t.TempDir(), "test-input")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			tmpFile.WriteString(input)
			tmpFile.Seek(0, 0)

			oldStdin := os.Stdin
			os.Stdin = tmpFile

			defer func() { os.Stdin = oldStdin }()

			if err = tokenMgr.Set(); tt.wantErr && err == nil {
				t.Error("Set() expected error but got nil")
			} else if !tt.wantErr && err != nil {
				t.Errorf("Set() error = %v, want nil", err)
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
			keyring.MockInit()

			tokenMgr := NewTokenManager()

			if tt.setupToken {
				currentUser, userErr := user.Current()
				if userErr != nil {
					t.Fatalf("Failed to get current user: %v", userErr)
				}

				keyring.Set(serviceName, currentUser.Username, "test-token")
			}

			if err := tokenMgr.Delete(); tt.wantErrType != nil {
				if err == nil || !errors.Is(err, tt.wantErrType) {
					t.Errorf("Delete() error = %v, want error type %v", err, tt.wantErrType)
				}
			} else if err != nil {
				t.Errorf("Delete() error = %v, want nil", err)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	// Capture stdout to hide prompts
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)

	defer func() { os.Stdout = oldStdout }()

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenMgr := NewTokenManager()

			tmpFile, err := os.CreateTemp(t.TempDir(), "test-input")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err = tmpFile.WriteString(tt.input); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}

			tmpFile.Seek(0, 0)

			oldStdin := os.Stdin
			os.Stdin = tmpFile

			defer func() { os.Stdin = oldStdin }()

			token, err := tokenMgr.create()

			if token != tt.wantToken {
				t.Errorf("create() token = %v, want %v", token, tt.wantToken)
			}

			if tt.wantErrType != nil {
				if err == nil || !errors.Is(err, tt.wantErrType) {
					t.Errorf("create() error = %v, want error type %v", err, tt.wantErrType)
				}
			} else if err != nil {
				t.Errorf("create() error = %v, want nil", err)
			}
		})
	}
}
