package download

import (
	"testing"

	"switchtube-downloader/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
