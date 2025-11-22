// Package token provides functionality for managing access tokens to authenticate with SwitchTube.
package token

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"strings"

	"switchtube-downloader/internal/helper/ui"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/zalando/go-keyring"
)

const (
	serviceName          = "SwitchTube"
	createAccessTokenURL = "https://tube.switch.ch/access_tokens"
)

const accessTokenLength = 43

var (
	// ErrTokenAlreadyExists is returned when trying to set a token that already
	// exists in the keyring.
	ErrTokenAlreadyExists = errors.New("token already exists in keyring")

	errFailedToDelete     = errors.New("failed to delete token from keyring")
	errFailedToGetUser    = errors.New("failed to get current user")
	errFailedToRetrieve   = errors.New("failed to retrieve token from keyring")
	errFailedToStore      = errors.New("failed to store token in keyring")
	errInvalidToken       = errors.New("invalid token provided")
	errNoTokenFoundDelete = errors.New("no token found in keyring")
	errNoTokenFound       = errors.New("no token found in keyring - run 'token set' first")
	errTokenEmpty         = errors.New("token cannot be empty")
	errUnableToCreate     = errors.New("unable to create access token")
)

// Manager encapsulates token management logic.
type Manager struct {
	keyringService string
}

// NewTokenManager creates a new instance of tokenManager.
func NewTokenManager() *Manager {
	return &Manager{
		keyringService: serviceName,
	}
}

// Delete removes the access token from the system keyring.
func (tm *Manager) Delete() error {
	userName, err := user.Current()
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToGetUser, err)
	}

	if err = keyring.Delete(tm.keyringService, userName.Username); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return fmt.Errorf("%w for %s", errNoTokenFoundDelete, tm.keyringService)
		}

		return fmt.Errorf("%w: %w", errFailedToDelete, err)
	}

	fmt.Println("✅ Token successfully deleted from keyring")

	return nil
}

// Get retrieves the access token from the system keyring.
func (tm *Manager) Get() (string, error) {
	userName, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("%w: %w", errFailedToGetUser, err)
	}

	token, err := keyring.Get(tm.keyringService, userName.Username)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", errNoTokenFound
		}

		return "", fmt.Errorf("%w: %w", errFailedToRetrieve, err)
	}

	return token, nil
}

// Set creates and stores a new access token in the system keyring.
func (tm *Manager) Set() error {
	existingToken, err := tm.Get()
	if err != nil && !errors.Is(err, errNoTokenFound) {
		return fmt.Errorf("%w: %w", errFailedToRetrieve, err)
	}

	if existingToken != "" {
		tm.displayTokenInfo(existingToken)
		fmt.Println()

		if !ui.Confirm("🔄 Do you want to replace it?") {
			fmt.Println("❌ Operation cancelled")

			return fmt.Errorf("%w", ErrTokenAlreadyExists)
		}
	}

	token, err := tm.create()
	if err != nil {
		return fmt.Errorf("%w: %w", errUnableToCreate, err)
	}

	userName, err := user.Current()
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToGetUser, err)
	}

	if err = keyring.Set(tm.keyringService, userName.Username, token); err != nil {
		return fmt.Errorf("%w: %w", errFailedToStore, err)
	}

	fmt.Println("\n✅ Token successfully stored in keyring")
	tm.displayTokenInfo(token)

	return nil
}

// create prompts the user to visit the access-token-creation URL and enter a new token.
func (tm *Manager) create() (string, error) {
	tm.displayInstructions()

	token := strings.TrimSpace(ui.Input("\n🔑 Enter your access token: "))
	if token == "" {
		return "", errTokenEmpty
	}

	if errors.Is(validateToken(token), errInvalidToken) {
		tm.displayValidationError()
		return "", errInvalidToken
	}

	return token, nil
}

// displayInstructions shows a formatted instruction table for token creation.
func (tm *Manager) displayInstructions() {
	fmt.Println()
	table := tablewriter.NewTable(
		os.Stdout,
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignCenter},
			},
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignLeft},
				Formatting: tw.CellFormatting{
					AutoWrap: tw.WrapNormal,
				},
			},
		}),
	)

	table.Header("📋 Token Creation Instructions")
	table.Append([]string{fmt.Sprintf("1️⃣  Visit: %s", createAccessTokenURL)})
	table.Append([]string{"2️⃣  Click 'Create a new access token'"})
	table.Append([]string{"3️⃣  Copy the generated token"})
	table.Append([]string{"4️⃣  Paste it below"})
	table.Render()
}

// displayTokenInfo shows information about the stored token.
func (tm *Manager) displayTokenInfo(token string) {
	userName, err := user.Current()
	if err != nil {
		return // Silently fail for display purposes
	}

	maskedToken := tm.maskToken(token)

	table := tablewriter.NewTable(
		os.Stdout,
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignCenter},
			},
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.AlignRight, tw.AlignLeft}},
			},
		}),
	)

	table.Header("Token Information")
	table.Append([]string{"Service", tm.keyringService})
	table.Append([]string{"User", userName.Username})
	table.Append([]string{"Token", maskedToken})
	table.Append([]string{"Length", fmt.Sprintf("%d characters", len(token))})
	table.Append([]string{"Status", "✓ Valid"})
	table.Render()
}

// displayValidationError shows validation requirements in a table.
func (tm *Manager) displayValidationError() {
	fmt.Println("\n❌ Invalid token format!")
	fmt.Println()

	table := tablewriter.NewTable(
		os.Stdout,
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignCenter},
			},
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{PerColumn: []tw.Align{tw.AlignRight, tw.AlignLeft}},
			},
		}),
	)

	table.Header("Token Requirements")
	table.Append([]string{"Length", "Exactly 43 characters"})
	table.Append([]string{"Characters", "A-Z, a-z, 0-9, -, _"})
	table.Append([]string{"Example", "AbC123-_xyz...  (43 chars total)"})
	table.Render()
}

// maskToken masks the middle portion of the token for security.
func (tm *Manager) maskToken(token string) string {
	if len(token) <= 10 {
		return strings.Repeat("*", len(token))
	}

	// Show first 5 and last 5 characters
	start := token[:5]
	end := token[len(token)-5:]
	middle := strings.Repeat("*", len(token)-10)

	return fmt.Sprintf("%s%s%s", start, middle, end)
}

// validateToken checks if the the token passed is valid.
//
// A token has following requirements:
//   - Length: 43 characters
//   - Valid characters: [ A-Z, a-z, 0-9, -, _ ]
//
//nolint:cyclop
func validateToken(token string) error {
	if len(token) != accessTokenLength {
		return errInvalidToken
	}

	for _, c := range token {
		if (c < 'a' || c > 'z') &&
			(c < 'A' || c > 'Z') &&
			(c < '0' || c > '9') &&
			c != '-' && c != '_' {
			return errInvalidToken
		}
	}

	return nil
}
