// Package download handles the downloading of videos and channels from SwitchTube.
package download

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"switchtube-downloader/internal/token"
)

var (
	errFailedToCreateRequest  = errors.New("failed to create request")
	errFailedToDecodeResponse = errors.New("failed to decode response")
	errFailedToGetToken       = errors.New("failed to get token")
)

// client handles all API interactions.
type client struct {
	tokenManager *token.Manager // Manages authentication tokens for API requests
	client       *http.Client   // HTTP client used for making requests
}

// newClient creates a new instance of Client.
func newClient(tm *token.Manager) *client {
	return &client{
		tokenManager: tm,
		client: &http.Client{
			Timeout:       0,
			Transport:     http.DefaultTransport,
			CheckRedirect: nil,
			Jar:           nil,
		},
	}
}

// makeJSONRequest makes an authenticated HTTP request and decodes the response.
func (c *client) makeJSONRequest(url string, target any) error {
	resp, err := c.makeRequest(url)
	if err != nil {
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Warning: failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d: %s",
			errHTTPNotOK,
			resp.StatusCode,
			http.StatusText(resp.StatusCode))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("%w: %w", errFailedToDecodeResponse, err)
	}

	return nil
}

// makeRequest makes an authenticated HTTP request.
func (c *client) makeRequest(url string) (*http.Response, error) {
	apiToken, err := c.tokenManager.Get()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToGetToken, err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToCreateRequest, err)
	}

	req.Header.Set(headerAuthorization, "Token "+apiToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToCreateRequest, err)
	}

	return resp, nil
}
