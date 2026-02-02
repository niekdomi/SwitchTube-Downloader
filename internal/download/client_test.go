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

func setupTestClient(t *testing.T) *client {
	t.Helper()
	keyring.MockInit()

	tokenMgr := token.NewTokenManager()
	currentUser, err := user.Current()
	require.NoError(t, err)
	err = keyring.Set("SwitchTube", currentUser.Username, "fake-token")
	require.NoError(t, err)

	return newClient(tokenMgr)
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

					if _, err := w.Write([]byte(tt.responseBody)); err != nil {
						t.Errorf("failed to write response: %v", err)
					}
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
	require.Error(t, err)

	if resp != nil {
		resp.Body.Close()
	}
}

func TestClient_makeRequest_TokenError(t *testing.T) {
	keyring.MockInit()

	tokenMgr := token.NewTokenManager()
	client := newClient(tokenMgr)

	resp, err := client.makeRequest("http://example.com")
	if resp != nil {
		resp.Body.Close()
	}

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get token")
}
