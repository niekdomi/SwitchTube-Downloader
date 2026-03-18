// Package download handles the downloading of videos and channels from SwitchTube.
package download

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"switchtube-downloader/internal/token"
)

var (
	errFailedToCreateRequest  = errors.New("failed to create request")
	errFailedToDecodeResponse = errors.New("failed to decode response")
	errFailedToGetToken       = errors.New("failed to get token")
	errFailedToParseBaseURL   = errors.New("failed to parse base URL")
	errUnexpectedHost         = errors.New("request URL host does not match expected base URL")
)

// client handles all API interactions.
type client struct {
	tokenManager *token.Manager // Manages authentication tokens for API requests
	client       *http.Client   // HTTP client used for making requests
	baseHost     string         // Expected host for SSRF validation
}

// newClient creates a new instance of Client.
func newClient(tm *token.Manager) (*client, error) {
	parsedBase, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToParseBaseURL, err)
	}

	return &client{
		tokenManager: tm,
		baseHost:     parsedBase.Host,
		client: &http.Client{
			Timeout:       0,
			Transport:     http.DefaultTransport,
			CheckRedirect: nil,
			Jar:           nil,
		},
	}, nil
}

// makeJSONRequest makes an authenticated HTTP request and decodes JSON response into target.
// Returns error if request fails or JSON decoding fails.
func (c *client) makeJSONRequest(ctx context.Context, reqURL string, target any) error {
	resp, err := c.makeRequest(ctx, reqURL)
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

// makeRequest makes an authenticated HTTP GET to the given URL.
func (c *client) makeRequest(ctx context.Context, reqURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToCreateRequest, err)
	}

	return c.makeRequestWithReq(req)
}

// makeRequestWithReq executes req after attaching the auth token header.
// Allows callers to supply a request with a custom context (e.g. for cancellation).
func (c *client) makeRequestWithReq(req *http.Request) (*http.Response, error) {
	// Validate request URL host to prevent SSRF
	if req.URL.Host != c.baseHost {
		return nil, fmt.Errorf("%w: got %q, want %q", errUnexpectedHost, req.URL.Host, c.baseHost)
	}

	apiToken, err := c.tokenManager.Get(req.Context())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToGetToken, err)
	}

	req.Header.Set(headerAuthorization, "Token "+apiToken)

	resp, err := c.client.Do(req) //nolint:gosec // URL host validated above against constant baseHost
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToCreateRequest, err)
	}

	return resp, nil
}
