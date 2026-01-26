// Package token provides functionality for managing access tokens to authenticate with SwitchTube.
package token

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/user"
	"strings"
	"time"

	"switchtube-downloader/internal/helper/ui/ansi"
	"switchtube-downloader/internal/helper/ui/input"
	"switchtube-downloader/internal/helper/ui/table"

	"github.com/zalando/go-keyring"
)

const (
	serviceName           = "SwitchTube"
	profileAPIURL         = "https://tube.switch.ch/api/v1/profiles/me"
	requestTimeoutSeconds = 10
)

const (
	maskThreshold    = 10
	maskVisibleChars = 5
)

var (
	// ErrTokenAlreadyExists is returned when attempting to set a token that already exists.
	ErrTokenAlreadyExists = errors.New("token already exists in keyring")

	errFailedToCloseResponse = errors.New("failed to close response body")
	errFailedToValidateToken = errors.New("failed to validate token")
	errNoToken               = errors.New("no token found in keyring - run 'token set' first")
	errTokenEmpty            = errors.New("token cannot be empty")
	errTokenInvalid          = errors.New("token authentication failed")
)

// Manager encapsulates token management logic.
type Manager struct {
	keyringService string
}

// NewTokenManager creates a new instance of Manager.
func NewTokenManager() *Manager {
	return &Manager{keyringService: serviceName}
}

// Delete removes the access token from the system keyring.
func (tm *Manager) Delete() error {
	username, err := tm.getUsername()
	if err != nil {
		return err
	}

	// Confirm deletion
	if !input.Confirm("Are you sure you want to delete the stored token?") {
		fmt.Printf("%s[CANCELLED]%s Token deletion cancelled\n", ansi.Warning, ansi.Reset)

		return nil
	}

	if err := keyring.Delete(tm.keyringService, username); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return fmt.Errorf("%w for %s", errNoToken, tm.keyringService)
		}

		return fmt.Errorf("failed to delete token: %w", err)
	}

	fmt.Printf("%s[SUCCESS]%s Token successfully deleted from keyring\n", ansi.Success, ansi.Reset)

	return nil
}

// Get retrieves the access token from the system keyring.
func (tm *Manager) Get() (string, error) {
	username, err := tm.getUsername()
	if err != nil {
		return "", err
	}

	token, err := keyring.Get(tm.keyringService, username)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", errNoToken
		}

		return token, fmt.Errorf("failed to retrieve token: %w", err)
	}

	if err := tm.validateToken(token); err != nil {
		return token, fmt.Errorf("stored token is invalid: %w", err)
	}

	return token, nil
}

// Set creates and stores a new access token in the system keyring.
func (tm *Manager) Set() error {
	if err := tm.checkExistingToken(); err != nil {
		return err
	}

	table.DisplayInstructions()

	token := input.Input("\nEnter your access token: ")
	if token == "" {
		return errTokenEmpty
	}

	fmt.Printf("\n%s[INFO]%s Validating token with SwitchTube API...\n", ansi.Info, ansi.Reset)

	if err := tm.validateToken(token); err != nil {
		fmt.Printf("\n%s[ERROR]%s Token validation failed\n", ansi.Error, ansi.Reset)
		tm.displayTokenInfo(token, false)

		return err
	}

	username, err := tm.getUsername()
	if err != nil {
		return err
	}

	if err := keyring.Set(tm.keyringService, username, token); err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	tm.displayTokenInfo(token, true)
	fmt.Printf("%s[SUCCESS]%s Token is valid and successfully stored in keyring\n", ansi.Success, ansi.Reset)

	return nil
}

// Validate validates the stored token and displays its status.
func (tm *Manager) Validate() error {
	fmt.Printf("\n%s[INFO]%s Validating token...\n", ansi.Info, ansi.Reset)

	// Get() already performs validation
	token, err := tm.Get()
	if err != nil {
		tm.displayTokenInfo(token, false)

		return err
	}

	tm.displayTokenInfo(token, true)

	return nil
}

// checkExistingToken checks if a token already exists and prompts for replacement.
func (tm *Manager) checkExistingToken() error {
	existingToken, err := tm.Get()
	if errors.Is(err, errNoToken) {
		return nil
	}

	tm.displayTokenInfo(existingToken, err == nil)

	fmt.Println()

	if !input.Confirm("Do you want to replace it?") {
		fmt.Printf("%s[CANCELLED]%s Operation cancelled\n", ansi.Warning, ansi.Reset)

		return ErrTokenAlreadyExists
	}

	return nil
}

// displayTokenInfo shows information about the token.
func (tm *Manager) displayTokenInfo(token string, valid bool) {
	username, err := tm.getUsername()
	if err != nil {
		return
	}

	var status string
	if valid {
		status = ansi.Success + "Valid" + ansi.Reset
	} else {
		status = ansi.Error + "Invalid" + ansi.Reset
	}

	table.DisplayTokenInfo(tm.keyringService, username, status, tm.maskToken(token), len(token))
}

// getUsername returns the current username.
func (tm *Manager) getUsername() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	return u.Username, nil
}

// maskToken masks the middle portion of the token.
func (tm *Manager) maskToken(token string) string {
	if len(token) <= maskThreshold {
		return strings.Repeat("*", len(token))
	}

	return token[:maskVisibleChars] +
		strings.Repeat("*", len(token)-maskThreshold) +
		token[len(token)-maskVisibleChars:]
}

// validateToken checks if the token is valid by making a request to the SwitchTube API.
func (tm *Manager) validateToken(token string) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, profileAPIURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{ //nolint:exhaustruct
		Timeout: requestTimeoutSeconds * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToValidateToken, err)
	}

	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("%w: %w", errFailedToCloseResponse, err)
	}

	if resp.StatusCode != http.StatusOK {
		return errTokenInvalid
	}

	return nil
}
