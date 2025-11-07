// Package token provides functionality for managing access tokens to authenticate with SwitchTube.
package token

import (
	"errors"
	"fmt"
	"os/user"

	"switchtube-downloader/internal/helper/ui"

	"github.com/zalando/go-keyring"
)

const (
	serviceName          = "SwitchTube"
	createAccessTokenURL = "https://tube.switch.ch/access_tokens"
)

var (
	// ErrTokenAlreadyExists is returned when trying to set a token that already
	// exists in the keyring.
	ErrTokenAlreadyExists = errors.New("token already exists in keyring")

	errFailedToDelete     = errors.New("failed to delete token from keyring")
	errFailedToGetUser    = errors.New("failed to get current user")
	errFailedToRetrieve   = errors.New("failed to retrieve token from keyring")
	errFailedToStore      = errors.New("failed to store token in keyring")
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
		fmt.Println("Token already exists in keyring")

		if !ui.Confirm("Do you want to replace it?") {
			fmt.Println("Operation cancelled")

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

	return nil
}

// create prompts the user to visit the access-token-creation URL and enter a new token.
func (tm *Manager) create() (string, error) {
	fmt.Printf("Please visit: %s\n", createAccessTokenURL)
	fmt.Printf("Create a new access token and paste it below\n\n")

	token := ui.Input("Enter your access token: ")
	// TODO: Validate token
	if token == "" {
		return "", errTokenEmpty
	}

	return token, nil
}
