package download

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"switchtube-downloader/internal/helper/dir"
	"switchtube-downloader/internal/helper/ui"
	"switchtube-downloader/internal/models"
)

// videoVariant represents a video download variant.
type videoVariant struct {
	Path      string `json:"path"`
	MediaType string `json:"mediaType"`
}

var (
	errFailedToConstructURL     = errors.New("failed to construct URL")
	errFailedToCopyVideoData    = errors.New("failed to copy video data")
	errFailedToCreateVideoFile  = errors.New("failed to create video file")
	errFailedToDecodeVariants   = errors.New("failed to decode variants")
	errFailedToDecodeVideoMeta  = errors.New("failed to decode video metadata")
	errFailedToFetchVideoStream = errors.New("failed to fetch video stream")
	errFailedToGetVideoInfo     = errors.New("failed to get video information")
	errFailedToGetVideoVariants = errors.New("failed to get video variants")
	errHTTPNotOK                = errors.New("HTTP request failed with non-OK status")
	errNoVariantsFound          = errors.New("no video variants found")
)

// videoDownloader handles the downloading of individual videos.
type videoDownloader struct {
	config   models.DownloadConfig
	progress models.ProgressInfo
	client   *Client
}

// newVideoDownloader creates a new instance of VideoDownloader.
func newVideoDownloader(
	config models.DownloadConfig,
	progress models.ProgressInfo,
	client *Client,
) *videoDownloader {
	return &videoDownloader{
		config:   config,
		progress: progress,
		client:   client,
	}
}

// downloadProcess handles the actual file download.
func (vd *videoDownloader) downloadProcess(endpoint string, file *os.File) error {
	fullURL, err := url.JoinPath(baseURL, endpoint)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToConstructURL, err)
	}

	resp, err := vd.client.makeRequest(fullURL)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToFetchVideoStream, err)
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

	currentItem := max(vd.progress.CurrentItem, 1)
	totalItems := max(vd.progress.TotalItems, 1)

	err = ui.ProgressBar(resp.Body, file, resp.ContentLength, file.Name(), currentItem, totalItems)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToCopyVideoData, err)
	}

	return nil
}

// downloadVideo downloads a video.
func (vd *videoDownloader) downloadVideo(videoID string, checkExists bool) error {
	video, err := vd.getMetadata(videoID)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToGetVideoInfo, err)
	}

	variants, err := vd.getVariants(videoID)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToGetVideoVariants, err)
	}

	if len(variants) == 0 {
		return errNoVariantsFound
	}

	filename := dir.CreateFilename(video.Title, variants[0].MediaType, video.Episode, vd.config)
	if checkExists && dir.OverwriteVideoIfExists(filename, vd.config) {
		return nil // Skip download
	}

	file, err := dir.CreateVideoFile(filename)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToCreateVideoFile, err)
	}

	// Download the video
	err = vd.downloadProcess(variants[0].Path, file)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToDownloadVideo, err)
	}

	return nil
}

// getMetadata retrieves video metadata from the API.
func (vd *videoDownloader) getMetadata(videoID string) (*models.Video, error) {
	fullURL, err := url.JoinPath(baseURL, videoAPI, videoID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToConstructURL, err)
	}

	var videoData models.Video
	if err := vd.client.makeJSONRequest(fullURL, &videoData); err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToDecodeVideoMeta, err)
	}

	return &videoData, nil
}

// getVariants retrieves available video variants from the API.
func (vd *videoDownloader) getVariants(videoID string) ([]videoVariant, error) {
	fullURL, err := url.JoinPath(baseURL, videoAPI, videoID, "video_variants")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToConstructURL, err)
	}

	var variants []videoVariant
	if err := vd.client.makeJSONRequest(fullURL, &variants); err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToDecodeVariants, err)
	}

	return variants, nil
}
