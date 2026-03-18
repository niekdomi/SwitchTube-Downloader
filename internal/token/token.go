// Package token provides functionality for managing access tokens to authenticate with SwitchTube.
package token

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"strings"
	"time"

	"switchtube-downloader/internal/helper/ui/input"
	"switchtube-downloader/internal/helper/ui/table"

	"github.com/charmbracelet/huh/spinner"
	charm "github.com/charmbracelet/log"
	"github.com/charmbracelet/x/term"
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

var log = charm.NewWithOptions(os.Stderr, charm.Options{
	ReportTimestamp: false,
	ReportCaller:    false,
})

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

	if !input.Confirm("Are you sure you want to delete the stored token?") {
		log.Warn("Token deletion cancelled")

		return nil
	}

	if err := keyring.Delete(tm.keyringService, username); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return fmt.Errorf("%w for %s", errNoToken, tm.keyringService)
		}

		return fmt.Errorf("failed to delete token: %w", err)
	}

	log.Info("Token successfully deleted from keyring")

	return nil
}

// Get retrieves the access token from the system keyring and validates it.
func (tm *Manager) Get(ctx context.Context) (string, error) {
	token, err := tm.GetRaw()
	if err != nil {
		return "", err
	}

	if err := tm.validateToken(ctx, token); err != nil {
		return token, fmt.Errorf("stored token is invalid: %w", err)
	}

	return token, nil
}

// GetAndDisplay retrieves the token and shows it in the info table.
func (tm *Manager) GetAndDisplay() error {
	token, validateErr := tm.getValidated("Validating token...")

	tm.displayTokenInfo(token, validateErr == nil)

	return validateErr
}

// GetRaw retrieves the token from the keyring without any validation.
// Use this when you just need the raw token value.
func (tm *Manager) GetRaw() (string, error) {
	username, err := tm.getUsername()
	if err != nil {
		return "", err
	}

	token, err := keyring.Get(tm.keyringService, username)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", errNoToken
		}

		return "", fmt.Errorf("failed to retrieve token: %w", err)
	}

	return token, nil
}

// Set creates and stores a new access token in the system keyring.
func (tm *Manager) Set() error {
	if err := tm.checkExistingToken(); err != nil {
		return err
	}

	table.DisplayInstructions()

	token := input.Input("Enter your access token")
	if token == "" {
		return errTokenEmpty
	}

	validateErr := tm.validateWithSpinner("Validating token with SwitchTube API...", token)
	if validateErr != nil {
		log.Error("Token validation failed", "err", validateErr)
		tm.displayTokenInfo(token, false)

		return validateErr
	}

	username, err := tm.getUsername()
	if err != nil {
		return err
	}

	if err := keyring.Set(tm.keyringService, username, token); err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	tm.displayTokenInfo(token, true)
	log.Info("Token is valid and successfully stored in keyring")

	return nil
}

// Validate validates the stored token and displays its status.
func (tm *Manager) Validate() error {
	token, validateErr := tm.getValidated("Validating token...")

	tm.displayTokenInfo(token, validateErr == nil)

	return validateErr
}

// checkExistingToken checks if a token already exists and prompts for replacement.
func (tm *Manager) checkExistingToken() error {
	existingToken, err := tm.Get(context.Background())
	if errors.Is(err, errNoToken) {
		return nil
	}

	tm.displayTokenInfo(existingToken, err == nil)

	fmt.Println()

	if !input.Confirm("Do you want to replace it?") {
		log.Warn("Operation cancelled")

		return ErrTokenAlreadyExists
	}

	return nil
}

// displayTokenInfo shows information about the token in a table.
func (tm *Manager) displayTokenInfo(token string, valid bool) {
	username, err := tm.getUsername()
	if err != nil {
		return
	}

	table.DisplayTokenInfo(tm.keyringService, username, valid, tm.maskToken(token), len(token))
}

// getUsername returns the current system username.
func (tm *Manager) getUsername() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	return u.Username, nil
}

// getValidated retrieves and validates the token, using a spinner in terminal mode.
func (tm *Manager) getValidated(title string) (string, error) {
	var (
		token       string
		validateErr error
	)

	if !term.IsTerminal(os.Stdout.Fd()) {
		token, validateErr = tm.Get(context.Background())
	} else {
		_ = spinner.New().
			Title(title).
			Context(context.Background()).
			ActionWithErr(func(ctx context.Context) error {
				token, validateErr = tm.Get(ctx)

				return nil
			}).
			Run()
	}

	return token, validateErr
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
func (tm *Manager) validateToken(ctx context.Context, token string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, profileAPIURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
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

// validateWithSpinner validates a token value, using a spinner in terminal mode.
func (tm *Manager) validateWithSpinner(title string, token string) error {
	if !term.IsTerminal(os.Stdout.Fd()) {
		return tm.validateToken(context.Background(), token)
	}

	var validateErr error

	_ = spinner.New().
		Title(title).
		Context(context.Background()).
		ActionWithErr(func(ctx context.Context) error {
			validateErr = tm.validateToken(ctx, token)

			return nil
		}).
		Run()

	return validateErr
}
