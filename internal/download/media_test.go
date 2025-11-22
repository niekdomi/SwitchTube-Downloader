package download

import (
	"net/http"
	"net/http/httptest"
	"os/user"
	"testing"

	"switchtube-downloader/internal/models"
	"switchtube-downloader/internal/token"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func setupTestClient(t *testing.T) *Client {
	t.Helper()
	keyring.MockInit()

	tokenMgr := token.NewTokenManager()
	currentUser, err := user.Current()
	require.NoError(t, err)
	err = keyring.Set("SwitchTube", currentUser.Username, "fake-token")
	require.NoError(t, err)

	return NewClient(tokenMgr)
}

func TestExtractIDAndType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantID   string
		wantType mediaType
		wantErr  error
	}{
		{
			name:     "video URL",
			input:    baseURL + videoPrefix + "123",
			wantID:   "123",
			wantType: videoType,
			wantErr:  nil,
		},
		{
			name:     "channel URL",
			input:    baseURL + channelPrefix + "abc",
			wantID:   "abc",
			wantType: channelType,
			wantErr:  nil,
		},
		{
			name:     "ID only (unknown type)",
			input:    "123",
			wantID:   "123",
			wantType: unknownType,
			wantErr:  nil,
		},
		{
			name:     "invalid URL",
			input:    baseURL + "invalid/123",
			wantID:   "invalid/123",
			wantType: unknownType,
			wantErr:  errInvalidURL,
		},
		{
			name:     "input with spaces",
			input:    "  " + baseURL + videoPrefix + "123  ",
			wantID:   "123",
			wantType: videoType,
			wantErr:  nil,
		},
		{
			name:     "empty input",
			input:    "",
			wantID:   "",
			wantType: unknownType,
			wantErr:  nil,
		},
		{
			name:     "base URL only",
			input:    baseURL,
			wantID:   "",
			wantType: unknownType,
			wantErr:  errInvalidURL,
		},
		{
			name:     "base URL with trailing slash",
			input:    baseURL + "/",
			wantID:   "/",
			wantType: unknownType,
			wantErr:  errInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, downloadType, err := extractIDAndType(tt.input)

			assert.Equal(t, tt.wantID, id)
			assert.Equal(t, tt.wantType, downloadType)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_makeJSONRequest(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseStatus int
		wantErr        bool
		expectedData   models.Video
	}{
		{
			name:           "request with invalid token",
			responseBody:   `{"id": "123", "title": "Test Video"}`,
			responseStatus: http.StatusOK,
			wantErr:        true,
			expectedData:   models.Video{},
		},
		{
			name:           "HTTP error",
			responseBody:   "",
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
		{
			name:           "invalid JSON",
			responseBody:   `{"invalid": json}`,
			responseStatus: http.StatusOK,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "Token fake-token", r.Header.Get(headerAuthorization))
					w.WriteHeader(tt.responseStatus)
					w.Write([]byte(tt.responseBody))
				}),
			)
			defer server.Close()

			client := setupTestClient(t)

			var result models.Video

			err := client.makeJSONRequest(server.URL, &result)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedData, result)
			}
		})
	}
}

func TestClient_makeRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Token fake-token", r.Header.Get(headerAuthorization))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := setupTestClient(t)

	resp, err := client.makeRequest(server.URL)
	assert.Error(t, err)

	if resp != nil {
		resp.Body.Close()
	}
}

func TestClient_makeRequest_TokenError(t *testing.T) {
	keyring.MockInit()

	tokenMgr := token.NewTokenManager()
	client := NewClient(tokenMgr)

	resp, err := client.makeRequest("http://example.com")
	if resp != nil {
		resp.Body.Close()
	}

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get token")
}

func TestDownload_InvalidURL(t *testing.T) {
	config := models.DownloadConfig{
		Media: baseURL + "invalid/123",
	}

	err := Download(config)
	assert.ErrorIs(t, err, errInvalidURL)
}

func TestDownload_UnknownType(t *testing.T) {
	config := models.DownloadConfig{
		Media: "some-id",
	}

	err := Download(config)

	require.Error(t, err)
	assert.NotErrorIs(t, err, errInvalidURL)
}

func TestDownload_EmptyMedia(t *testing.T) {
	config := models.DownloadConfig{
		Media: "",
	}

	err := Download(config)
	assert.Error(t, err)
}

func TestDownload_BaseURLOnly(t *testing.T) {
	config := models.DownloadConfig{
		Media: baseURL,
	}

	err := Download(config)
	assert.ErrorIs(t, err, errInvalidURL)
}
