// Package dir provides functions to create video files and channel folders.
package dir

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"switchtube-downloader/internal/helper/ui/input"
	"switchtube-downloader/internal/models"
)

const (
	// File and directory permissions.
	dirPermissions = 0o755
)

var (
	// ErrFailedToCreateFile is returned when file creation fails.
	ErrFailedToCreateFile = errors.New("failed to create file")

	errFailedToCreateFolder = errors.New("failed to create folder")
)

// CreateFilename creates a sanitized filename from video title and media type.
func CreateFilename(title string, mediaType string, episodeNr string, config models.DownloadConfig) string {
	// Extract extension from media type (e.g., "video/mp4" -> "mp4")
	_, extension, found := strings.Cut(mediaType, "/")
	if !found {
		extension = "mp4" // default fallback
	}

	sanitizedTitle := sanitizeFilename(title)
	sanitizedTitle = strings.ReplaceAll(sanitizedTitle, " ", "_")

	// Add episode prefix if episode flag is set
	var filename string
	if config.UseEpisode && episodeNr != "" {
		filename = fmt.Sprintf("%s_%s.%s", episodeNr, sanitizedTitle, extension)
	} else {
		filename = fmt.Sprintf("%s.%s", sanitizedTitle, extension)
	}

	if config.Output != "" {
		filename = filepath.Join(config.Output, filename)
	}

	return filepath.Clean(filename)
}

// OverwriteVideoIfExists checks if a video file exists and prompts to overwrite
// it. Returns false if the file doesn't exist or if overwriting is declined.
func OverwriteVideoIfExists(filename string, config models.DownloadConfig) bool {
	if !config.Force {
		if _, err := os.Stat(filename); err == nil {
			if config.Skip || !input.Confirm("File %s already exists. Overwrite?", filename) {
				return true
			}
		}
	}

	return false
}

// CreateVideoFile creates a video file on disk with the specified filename.
func CreateVideoFile(filename string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(filename), dirPermissions); err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToCreateFolder, err)
	}

	fd, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToCreateFile, err)
	}

	return fd, nil
}

// CreateChannelFolder creates a folder for the channel using its name.
func CreateChannelFolder(channelName string, config models.DownloadConfig) (string, error) {
	folderName := strings.ReplaceAll(channelName, "/", " - ")
	folderName = filepath.Clean(folderName)

	if config.Output != "" {
		folderName = filepath.Join(config.Output, folderName)
	}

	if err := os.MkdirAll(folderName, dirPermissions); err != nil {
		return "", fmt.Errorf("%w: %w", errFailedToCreateFolder, err)
	}

	return folderName, nil
}

// sanitizeFilename removes or replaces invalid characters in filenames.
func sanitizeFilename(filename string) string {
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "",
		"?", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "-",
	)

	sanitized := replacer.Replace(filename)
	sanitized = strings.TrimSpace(sanitized)

	for strings.Contains(sanitized, "--") {
		sanitized = strings.ReplaceAll(sanitized, "--", "-")
	}

	return sanitized
}
