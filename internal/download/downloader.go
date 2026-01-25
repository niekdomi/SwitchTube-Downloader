// Package download handles the downloading of videos and channels from SwitchTube.
package download

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"switchtube-downloader/internal/helper/dir"
	"switchtube-downloader/internal/helper/ui/colors"
	"switchtube-downloader/internal/helper/ui/input"
	"switchtube-downloader/internal/helper/ui/progress"
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
	errFailedToConstructURL        = errors.New("failed to construct URL")
	errFailedToCopyVideoData       = errors.New("failed to copy video data")
	errFailedToCreateChannelFolder = errors.New("failed to create channel folder")
	errFailedToCreateRequest       = errors.New("failed to create request")
	errFailedToCreateVideoFile     = errors.New("failed to create video file")
	errFailedToDecodeChannelMeta   = errors.New("failed to decode channel metadata")
	errFailedToDecodeChannelVideos = errors.New("failed to decode channel videos")
	errFailedToDecodeResponse      = errors.New("failed to decode response")
	errFailedToDecodeVariants      = errors.New("failed to decode variants")
	errFailedToDecodeVideoMeta     = errors.New("failed to decode video metadata")
	errFailedToDownloadChannel     = errors.New("failed to download channel")
	errFailedToDownloadVideo       = errors.New("failed to download video")
	errFailedToExtractType         = errors.New("failed to extract type")
	errFailedToFetchVideoStream    = errors.New("failed to fetch video stream")
	errFailedToGetChannelInfo      = errors.New("failed to get channel information")
	errFailedToGetChannelVideos    = errors.New("failed to get channel videos")
	errFailedToGetToken            = errors.New("failed to get token")
	errFailedToGetVideoInfo        = errors.New("failed to get video information")
	errFailedToGetVideoVariants    = errors.New("failed to get video variants")
	errFailedToSelectVideos        = errors.New("failed to select videos")
	errHTTPNotOK                   = errors.New("HTTP request failed with non-OK status")
	errInvalidID                   = errors.New("invalid id")
	errInvalidURL                  = errors.New("invalid url")
	errNoVariantsFound             = errors.New("no video variants found")
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

// makeJSONRequest makes an authenticated HTTP request and decodes the response.
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

// Downloader handles downloading of both videos and channels.
type Downloader struct {
	config models.DownloadConfig
	client *Client
}

// New creates a new Downloader instance.
func New(config models.DownloadConfig, client *Client) *Downloader {
	return &Downloader{
		config: config,
		client: client,
	}
}

// Download initiates the download process based on the provided configuration.
func Download(config models.DownloadConfig) error {
	id, downloadType, err := extractIDAndType(config.Media)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToExtractType, err)
	}

	tokenMgr := token.NewTokenManager()
	client := NewClient(tokenMgr)
	downloader := New(config, client)

	switch downloadType {
	case videoType, unknownType:
		if err = downloader.downloadVideo(id, true, 0, 0); err == nil {
			return nil
		} else if downloadType == videoType || errors.Is(err, dir.ErrFailedToCreateFile) {
			return fmt.Errorf("%w: %w", errFailedToDownloadVideo, err)
		}

		// Fallthrough if type is unknown and try as channel
		fallthrough
	case channelType:
		if err = downloader.downloadChannel(id); err != nil {
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
	// This is the case if ht Id was passed as an argument
	prefixAndID, hasPrefix := strings.CutPrefix(input, baseURL)
	if !hasPrefix {
		return input, unknownType, nil
	}

	// Try to extract video ID
	if id, found := strings.CutPrefix(prefixAndID, videoPrefix); found {
		return id, videoType, nil
	}

	// Try to extract channel ID
	if id, found := strings.CutPrefix(prefixAndID, channelPrefix); found {
		return id, channelType, nil
	}

	return prefixAndID, unknownType, errInvalidURL
}

// videoVariant represents a video download variant.
type videoVariant struct {
	Path      string `json:"path"`
	MediaType string `json:"mediaType"`
}

// channelMetadata represents channel metadata.
type channelMetadata struct {
	Name string `json:"name"`
}

// downloadChannel downloads selected videos from a channel.
func (d *Downloader) downloadChannel(channelID string) error {
	channelInfo, err := d.getChannelMetadata(channelID)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToGetChannelInfo, err)
	}

	videos, err := d.getChannelVideos(channelID)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToGetChannelVideos, err)
	}

	if len(videos) == 0 {
		fmt.Println("No videos found in this channel")

		return nil
	}

	fmt.Printf("Found %d videos in channel: %s\n", len(videos), channelInfo.Name)

	selectedIndices, err := input.SelectVideos(videos, d.config.All, d.config.UseEpisode)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToSelectVideos, err)
	}

	if len(selectedIndices) == 0 {
		fmt.Println("No videos selected for download")

		return nil
	}

	folderName, err := dir.CreateChannelFolder(channelInfo.Name, d.config)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToCreateChannelFolder, err)
	}

	d.config.Output = folderName
	fmt.Printf("\r\nDownloading to folder: %s\n\n", folderName)
	d.downloadSelectedVideos(videos, selectedIndices)

	return nil
}

// downloadSelectedVideos downloads the selected videos and reports results.
func (d *Downloader) downloadSelectedVideos(videos []models.Video, selectedIndices []int) {
	var failed []string

	toDownload, maxWidth := d.prepareDownloads(videos, selectedIndices, &failed)
	if len(toDownload) > 0 {
		failed = append(failed, d.processDownloads(videos, toDownload, maxWidth)...)
	}

	d.printResults(len(toDownload), len(selectedIndices), failed)
}

// downloadVideo downloads a single video with optional progress bar positioning.
// rowIndex and maxFilenameWidth are used for multi-line progress display (0 = single-line mode).
func (d *Downloader) downloadVideo(videoID string, checkExists bool, rowIndex int, maxFilenameWidth int) error {
	video, err := d.getVideoMetadata(videoID)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToGetVideoInfo, err)
	}

	variants, err := d.getVideoVariants(videoID)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToGetVideoVariants, err)
	}

	if len(variants) == 0 {
		return errNoVariantsFound
	}

	filename := dir.CreateFilename(video.Title, variants[0].MediaType, video.Episode, d.config)
	if checkExists && dir.OverwriteVideoIfExists(filename, d.config) {
		return nil // Skip download
	}

	file, err := dir.CreateVideoFile(filename)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToCreateVideoFile, err)
	}

	// Download the video
	err = d.downloadVideoStream(variants[0].Path, file, rowIndex, maxFilenameWidth)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToDownloadVideo, err)
	}

	return nil
}

// downloadVideoStream handles the actual file download.
func (d *Downloader) downloadVideoStream(endpoint string, file *os.File, rowIndex int, maxFilenameWidth int) error {
	fullURL, err := url.JoinPath(baseURL, endpoint)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToConstructURL, err)
	}

	resp, err := d.client.makeRequest(fullURL)
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

	err = progress.BarWithRow(resp.Body, file, resp.ContentLength, file.Name(), rowIndex, maxFilenameWidth)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToCopyVideoData, err)
	}

	return nil
}

// downloadVideosParallel downloads multiple videos concurrently.
func (d *Downloader) downloadVideosParallel(videos []models.Video, indices []int, maxWidth int, failedLock *sync.Mutex) []string {
	var (
		failed           []string
		wg               sync.WaitGroup
		maxFilenameWidth = maxWidth
		widthMutex       sync.Mutex
	)

	numVideos := len(indices)

	for i, idx := range indices {
		wg.Add(1)

		go func(videoIdx int, rowIndex int) {
			defer wg.Done()

			video := videos[videoIdx]

			// Get variants to calculate filename width
			variants, err := d.getVideoVariants(video.ID)
			if err != nil || len(variants) == 0 {
				failedLock.Lock()

				failed = append(failed, video.Title)

				failedLock.Unlock()

				return
			}

			filename := dir.CreateFilename(video.Title, variants[0].MediaType, video.Episode, d.config)
			basename := filepath.Base(filename)

			widthMutex.Lock()

			if len(basename) > maxFilenameWidth {
				maxFilenameWidth = len(basename)
			}

			currentMaxWidth := maxFilenameWidth

			widthMutex.Unlock()

			if err := d.downloadVideo(video.ID, false, rowIndex, currentMaxWidth); err != nil {
				failedLock.Lock()

				failed = append(failed, video.Title)

				failedLock.Unlock()
			}
		}(idx, numVideos-i)
	}

	wg.Wait()

	return failed
}

// getChannelMetadata retrieves channel metadata from the API.
func (d *Downloader) getChannelMetadata(channelID string) (*channelMetadata, error) {
	fullURL, err := url.JoinPath(baseURL, channelAPI, channelID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToConstructURL, err)
	}

	var data channelMetadata
	if err := d.client.makeJSONRequest(fullURL, &data); err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToDecodeChannelMeta, err)
	}

	return &data, nil
}

// getChannelVideos retrieves all videos from a channel.
func (d *Downloader) getChannelVideos(channelID string) ([]models.Video, error) {
	fullURL, err := url.JoinPath(baseURL, channelAPI, channelID, "videos")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToConstructURL, err)
	}

	var videos []models.Video
	if err := d.client.makeJSONRequest(fullURL, &videos); err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToDecodeChannelVideos, err)
	}

	return videos, nil
}

// getVideoMetadata retrieves video metadata from the API.
func (d *Downloader) getVideoMetadata(videoID string) (*models.Video, error) {
	fullURL, err := url.JoinPath(baseURL, videoAPI, videoID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToConstructURL, err)
	}

	var videoData models.Video
	if err := d.client.makeJSONRequest(fullURL, &videoData); err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToDecodeVideoMeta, err)
	}

	return &videoData, nil
}

// getVideoVariants retrieves available video variants from the API.
func (d *Downloader) getVideoVariants(videoID string) ([]videoVariant, error) {
	fullURL, err := url.JoinPath(baseURL, videoAPI, videoID, "video_variants")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToConstructURL, err)
	}

	var variants []videoVariant
	if err := d.client.makeJSONRequest(fullURL, &variants); err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToDecodeVariants, err)
	}

	return variants, nil
}

// prepareDownloads checks which videos need to be downloaded and validates their availability.
func (d *Downloader) prepareDownloads(videos []models.Video, indices []int, failed *[]string) ([]int, int) {
	var (
		toDownload []int
		maxWidth   int
	)

	for _, idx := range indices {
		video := videos[idx]

		variants, err := d.getVideoVariants(video.ID)
		if err != nil {
			fmt.Printf("\nFailed to get video variants for %s: %v\n", video.Title, err)
			*failed = append(*failed, video.Title)

			continue
		}

		if len(variants) == 0 {
			fmt.Printf("\nNo variants found for %s\n", video.Title)
			*failed = append(*failed, video.Title)

			continue
		}

		filename := dir.CreateFilename(video.Title, variants[0].MediaType, video.Episode, d.config)
		if !dir.OverwriteVideoIfExists(filename, d.config) {
			toDownload = append(toDownload, idx)

			basename := filepath.Base(filename)
			if len(basename) > maxWidth {
				maxWidth = len(basename)
			}
		}
	}

	return toDownload, maxWidth
}

// printResults displays the download results summary.
func (d *Downloader) printResults(downloadCount int, selectedCount int, failed []string) {
	successCount := downloadCount - len(failed)
	fmt.Printf("\n%s[SUCCESS]%s Download complete! %d/%d videos successful\n", colors.Success, colors.Reset, successCount, selectedCount)

	if len(failed) > 0 {
		fmt.Printf("%s[ERROR]%s Failed downloads:\n", colors.Error, colors.Reset)

		for _, title := range failed {
			fmt.Printf("  - %s\n", title)
		}
	}
}

// processDownloads performs the actual video downloads in parallel and returns failed video titles.
func (d *Downloader) processDownloads(videos []models.Video, indices []int, maxWidth int) []string {
	var (
		failed     []string
		failedLock sync.Mutex
	)

	numVideos := len(indices)

	fmt.Print("\033[?25l") // Hide cursor

	for range numVideos {
		fmt.Println() // Reserve a line for each video
	}

	failed = d.downloadVideosParallel(videos, indices, maxWidth, &failedLock)

	fmt.Print("\033[?25h") // Show cursor

	return failed
}
