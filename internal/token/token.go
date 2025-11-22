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

	"switchtube-downloader/internal/helper/ui"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/zalando/go-keyring"
)

const (
	serviceName          = "SwitchTube"
	createAccessTokenURL = "https://tube.switch.ch/access_tokens"
	profileAPIURL        = "https://tube.switch.ch/api/v1/profiles/me"
	timeout              = 10
)

const (
	tokenMaskThreshold    = 10
	tokenMaskVisibleChars = 5
)

var (
	// ErrTokenAlreadyExists is returned when attempting to set a token that already exists.
	ErrTokenAlreadyExists = errors.New("token already exists in keyring")

	errNoToken      = errors.New("no token found in keyring - run 'token set' first")
	errTokenEmpty   = errors.New("token cannot be empty")
	errTokenInvalid = errors.New("token authentication failed")
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

	if err := keyring.Delete(tm.keyringService, username); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return fmt.Errorf("%w for %s", errNoToken, tm.keyringService)
		}

		return fmt.Errorf("failed to delete token: %w", err)
	}

	fmt.Println("✅ Token successfully deleted from keyring")

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

	tm.displayInstructions()

	token := strings.TrimSpace(ui.Input("\n🔑 Enter your access token: "))
	if token == "" {
		return errTokenEmpty
	}

	fmt.Println("\n🔍 Validating token with SwitchTube API...")

	if err := tm.validateToken(token); err != nil {
		fmt.Println("\n❌ Token validation failed")
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
	fmt.Println("✅ Token is valid and successfully stored in keyring")

	return nil
}

// Validate validates the stored token and displays its status.
func (tm *Manager) Validate() error {
	fmt.Println("\n🔍 Validating token...")

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

	if !ui.Confirm("🔄 Do you want to replace it?") {
		fmt.Println("❌ Operation cancelled")

		return ErrTokenAlreadyExists
	}

	return nil
}

// createTable creates a tablewriter with standard configuration.
func (tm *Manager) createTable(header string, alignments ...tw.Align) *tablewriter.Table {
	//nolint:exhaustruct
	config := tablewriter.Config{
		Header: tw.CellConfig{
			Alignment: tw.CellAlignment{Global: tw.AlignCenter},
		},
		Row: tw.CellConfig{
			Alignment: tw.CellAlignment{Global: tw.AlignLeft},
		},
	}

	if len(alignments) > 0 {
		config.Row.Alignment = tw.CellAlignment{Global: tw.AlignCenter, PerColumn: alignments}
	}

	table := tablewriter.NewTable(os.Stdout, tablewriter.WithConfig(config))
	table.Header(header)

	return table
}

// displayInstructions shows formatted instructions for token creation.
func (tm *Manager) displayInstructions() {
	fmt.Println()

	table := tm.createTable("📋 Token Creation Instructions", tw.AlignLeft)
	_ = table.Append([]string{"1️⃣  Visit: " + createAccessTokenURL})
	_ = table.Append([]string{"2️⃣  Click 'Create New Token'"})
	_ = table.Append([]string{"3️⃣  Copy the generated token"})
	_ = table.Append([]string{"4️⃣  Paste it below"})
	_ = table.Render()
}

// displayTokenInfo shows information about the token.
func (tm *Manager) displayTokenInfo(token string, valid bool) {
	username, err := tm.getUsername()
	if err != nil {
		return
	}

	status := "🟢 Valid"
	if !valid {
		status = "🔴 Invalid"
	}

	table := tm.createTable("Token Information", tw.AlignRight, tw.AlignLeft)
	_ = table.Append([]string{"Service", tm.keyringService})
	_ = table.Append([]string{"User", username})
	_ = table.Append([]string{"Token", tm.maskToken(token)})
	_ = table.Append([]string{"Length", fmt.Sprintf("%d characters", len(token))})
	_ = table.Append([]string{"Status", status})
	_ = table.Render()
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
	if len(token) <= tokenMaskThreshold {
		return strings.Repeat("*", len(token))
	}

	return token[:tokenMaskVisibleChars] + strings.Repeat(
		"*",
		len(token)-tokenMaskThreshold,
	) + token[len(token)-tokenMaskVisibleChars:]
}

// validateToken checks if the token is valid by making a request to the SwitchTube API.
func (tm *Manager) validateToken(token string) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, profileAPIURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Timeout:       timeout * time.Second,
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errTokenInvalid
	}

	return nil
}
