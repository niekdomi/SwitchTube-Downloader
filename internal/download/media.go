// Package download handles the downloading of videos and channels from SwitchTube.
package download

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"switchtube-downloader/internal/helper/dir"
	"switchtube-downloader/internal/models"
	"switchtube-downloader/internal/token"
)

const (
	// Base URL and API endpoints for SwitchTube.
	baseURL             = "https://tube.switch.ch/"
	videoAPI            = "api/v1/browse/videos/"
	channelAPI          = "api/v1/browse/channels/"
	videoPrefix         = "videos/"
	channelPrefix       = "channels/"
	headerAuthorization = "Authorization"
)

type mediaType int

const (
	unknownType mediaType = iota
	videoType
	channelType
)

var (
	errFailedToCreateRequest   = errors.New("failed to create request")
	errFailedToDecodeResponse  = errors.New("failed to decode response")
	errFailedToDownloadChannel = errors.New("failed to download channel")
	errFailedToDownloadVideo   = errors.New("failed to download video")
	errFailedToExtractType     = errors.New("failed to extract type")
	errFailedToGetToken        = errors.New("failed to get token")
	errHTTPNotOK               = errors.New("HTTP request failed with non-OK status")
	errInvalidID               = errors.New("invalid id")
	errInvalidURL              = errors.New("invalid url")
)

// Client handles all API interactions.
type Client struct {
	tokenManager *token.Manager
	client       *http.Client
}

// NewClient creates a new instance of Client.
func NewClient(tm *token.Manager) *Client {
	return &Client{
		tokenManager: tm,
		client: &http.Client{
			Timeout:       0,
			Transport:     http.DefaultTransport,
			CheckRedirect: nil,
			Jar:           nil,
		},
	}
}

// makeRequest makes an authenticated HTTP request and decodes the response.
func (c *Client) makeJSONRequest(url string, target any) error {
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
func (c *Client) makeRequest(url string) (*http.Response, error) {
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

// Download initiates the download process based on the provided configuration.
func Download(config models.DownloadConfig) error {
	id, downloadType, err := extractIDAndType(config.Media)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToExtractType, err)
	}

	tokenMgr := token.NewTokenManager()
	client := NewClient(tokenMgr)

	switch downloadType {
	case videoType, unknownType:
		downloader := newVideoDownloader(config, client)

		if err = downloader.download(id, true, 0, 0); err == nil {
			return nil
		} else if downloadType == videoType || errors.Is(err, dir.ErrFailedToCreateFile) {
			return fmt.Errorf("%w: %w", errFailedToDownloadVideo, err)
		}

		// Fallthrough if type is unknown and try as channel again
		fallthrough
	case channelType:
		downloader := newChannelDownloader(config, client)

		if err = downloader.download(id); err != nil {
			if downloadType == unknownType {
				return fmt.Errorf("%w", errInvalidID)
			}

			return fmt.Errorf("%w: %w", errFailedToDownloadChannel, err)
		}
	}

	return nil
}

// extractIDAndType extracts the id and determines if it's a video or channel.
func extractIDAndType(input string) (string, mediaType, error) {
	input = strings.TrimSpace(input)

	// If input doesn't start with baseURL, return as unknown type
	// This is the case if the Id was passed as an argument
	if !strings.HasPrefix(input, baseURL) {
		return input, unknownType, nil
	}

	switch prefixAndID := strings.TrimPrefix(input, baseURL); {
	case strings.HasPrefix(prefixAndID, videoPrefix):
		return strings.TrimPrefix(prefixAndID, videoPrefix), videoType, nil
	case strings.HasPrefix(prefixAndID, channelPrefix):
		return strings.TrimPrefix(prefixAndID, channelPrefix), channelType, nil
	default:
		return prefixAndID, unknownType, errInvalidURL
	}
}
