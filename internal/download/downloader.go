package download

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"switchtube-downloader/internal/helper/dir"
	"switchtube-downloader/internal/helper/ui/input"
	"switchtube-downloader/internal/helper/ui/progress"
	"switchtube-downloader/internal/helper/ui/styles"
	"switchtube-downloader/internal/models"
	"switchtube-downloader/internal/token"

	"github.com/charmbracelet/x/ansi"
)

// Base URL and API endpoints for SwitchTube.
const (
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
	errFailedToCreateVideoFile     = errors.New("failed to create video file")
	errFailedToDecodeChannelMeta   = errors.New("failed to decode channel metadata")
	errFailedToDecodeChannelVideos = errors.New("failed to decode channel videos")
	errFailedToDecodeVariants      = errors.New("failed to decode variants")
	errFailedToDecodeVideoMeta     = errors.New("failed to decode video metadata")
	errFailedToDownloadChannel     = errors.New("failed to download channel")
	errFailedToDownloadVideo       = errors.New("failed to download video")
	errFailedToExtractType         = errors.New("failed to extract type")
	errFailedToFetchVideoStream    = errors.New("failed to fetch video stream")
	errFailedToGetChannelInfo      = errors.New("failed to get channel information")
	errFailedToGetChannelVideos    = errors.New("failed to get channel videos")
	errFailedToGetVideoInfo        = errors.New("failed to get video information")
	errFailedToGetVideoVariants    = errors.New("failed to get video variants")
	errFailedToSelectVideos        = errors.New("failed to select videos")
	errHTTPNotOK                   = errors.New("HTTP request failed with non-OK status")
	errInvalidID                   = errors.New("invalid id")
	errInvalidURL                  = errors.New("invalid url")
	errNoVariantsFound             = errors.New("no video variants found")
)

// videoVariant represents a video download variant.
type videoVariant struct {
	Path      string `json:"path"`       // Relative path to the video file on the server
	MediaType string `json:"media_type"` //nolint:tagliatelle // API returns snake_case
}

// channelMetadata represents channel metadata.
type channelMetadata struct {
	Name string `json:"name"` // Display name of the channel
}

// downloader handles downloading of both videos and channels.
type downloader struct {
	client *client
	config models.DownloadConfig
}

// newDownloader creates a new Downloader instance.
func newDownloader(config models.DownloadConfig, client *client) *downloader {
	return &downloader{
		config: config,
		client: client,
	}
}

// downloadChannel downloads selected videos from a channel.
// Fetches channel info, displays video list, prompts for selection, and downloads chosen videos.
func (d *downloader) downloadChannel(ctx context.Context, channelID string) error {
	channelInfo, err := d.getChannelMetadata(ctx, channelID)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToGetChannelInfo, err)
	}

	videos, err := d.getChannelVideos(ctx, channelID)
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

	d.config.OutputDir = folderName
	fmt.Printf("\r\nDownloading to folder: %s\n\n", folderName)
	d.downloadSelectedVideos(ctx, videos, selectedIndices)

	return nil
}

// downloadSelectedVideos downloads the videos at the given indices and prints a summary.
func (d *downloader) downloadSelectedVideos(ctx context.Context, videos []models.Video, selectedIndices []int) {
	var failed []string

	videosToDownload, longestVideoName := d.prepareDownloads(ctx, videos, selectedIndices, &failed)
	if len(videosToDownload) > 0 {
		failed = append(failed, d.processDownloads(ctx, videos, videosToDownload, longestVideoName)...)
	}

	d.printResults(ctx, len(selectedIndices), failed)
}

// downloadVideo downloads a single video by ID. Returns error if download fails.
// rowIndex and maxFilenameWidth are used for multi-file progress display alignment.
func (d *downloader) downloadVideo(ctx context.Context, videoID string, checkExists bool, rowIndex int, maxFilenameWidth int) error {
	video, err := d.getVideoMetadata(ctx, videoID)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToGetVideoInfo, err)
	}

	variants, err := d.getVideoVariants(ctx, videoID)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToGetVideoVariants, err)
	}

	if len(variants) == 0 {
		return errNoVariantsFound
	}

	filename := dir.CreateFilename(video.Title, variants[0].MediaType, video.Episode, d.config)
	if checkExists && !dir.OverwriteVideoIfExists(filename, d.config) {
		return nil // Skip download
	}

	file, err := dir.CreateVideoFile(filename)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToCreateVideoFile, err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("Warning: failed to close video file: %v\n", err)
		}
	}()

	// Download the video
	err = d.downloadVideoStream(ctx, variants[0].Path, file, rowIndex, maxFilenameWidth)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToDownloadVideo, err)
	}

	return nil
}

// downloadVideoStream downloads video data from endpoint to file with progress tracking.
func (d *downloader) downloadVideoStream(ctx context.Context, endpoint string, file *os.File, rowIndex int, maxFilenameWidth int) error {
	fullURL, err := url.JoinPath(baseURL, endpoint)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToConstructURL, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToFetchVideoStream, err)
	}

	resp, err := d.client.makeRequestWithReq(req)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("download cancelled: %w", ctx.Err())
		}

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
		if ctx.Err() != nil {
			return fmt.Errorf("download cancelled: %w", ctx.Err())
		}

		return fmt.Errorf("%w: %w", errFailedToCopyVideoData, err)
	}

	return nil
}

// downloadVideosParallel downloads multiple videos concurrently.
// Returns slice of failed video titles.
func (d *downloader) downloadVideosParallel(ctx context.Context, videos []models.Video, indices []int, longestVideoName int) []string {
	var (
		failed []string
		wg     sync.WaitGroup
		mutex  sync.Mutex
	)

	numVideos := len(indices)

	for i, idx := range indices {
		if ctx.Err() != nil {
			break // context already cancelled
		}

		wg.Add(1)

		go func(videoIdx int, rowIndex int) {
			defer wg.Done()

			if ctx.Err() != nil {
				return // aborted before we started
			}

			video := videos[videoIdx]

			// Get variants to calculate filename width
			variants, err := d.getVideoVariants(ctx, video.ID)
			if err != nil || len(variants) == 0 {
				mutex.Lock()
				failed = append(failed, video.Title)
				mutex.Unlock()

				return
			}

			filename := dir.CreateFilename(video.Title, variants[0].MediaType, video.Episode, d.config)
			basename := filepath.Base(filename)

			mutex.Lock()
			longestVideoName = max(len(basename), longestVideoName)
			currentLongest := longestVideoName
			mutex.Unlock()

			if err := d.downloadVideo(ctx, video.ID, false, rowIndex, currentLongest); err != nil {
				if ctx.Err() == nil { // only record failure if not cancelled
					mutex.Lock()
					failed = append(failed, video.Title)
					mutex.Unlock()
				}
			}
		}(idx, numVideos-i)
	}

	wg.Wait()

	return failed
}

// getChannelMetadata retrieves channel metadata from the API.
// Returns channel metadata including name.
func (d *downloader) getChannelMetadata(ctx context.Context, channelID string) (*channelMetadata, error) {
	fullURL, err := url.JoinPath(baseURL, channelAPI, channelID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToConstructURL, err)
	}

	var data channelMetadata
	if err := d.client.makeJSONRequest(ctx, fullURL, &data); err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToDecodeChannelMeta, err)
	}

	return &data, nil
}

// getChannelVideos retrieves all videos from a channel.
// Returns slice of videos with their IDs, titles, and episode numbers.
func (d *downloader) getChannelVideos(ctx context.Context, channelID string) ([]models.Video, error) {
	fullURL, err := url.JoinPath(baseURL, channelAPI, channelID, "videos")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToConstructURL, err)
	}

	var videos []models.Video
	if err := d.client.makeJSONRequest(ctx, fullURL, &videos); err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToDecodeChannelVideos, err)
	}

	return videos, nil
}

// getVideoMetadata retrieves video metadata from the API.
// Returns video info including ID, title, and episode number.
func (d *downloader) getVideoMetadata(ctx context.Context, videoID string) (*models.Video, error) {
	fullURL, err := url.JoinPath(baseURL, videoAPI, videoID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToConstructURL, err)
	}

	var videoData models.Video
	if err := d.client.makeJSONRequest(ctx, fullURL, &videoData); err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToDecodeVideoMeta, err)
	}

	return &videoData, nil
}

// getVideoVariants retrieves available video variants from the API.
// Returns slice of variants with download paths and media types.
func (d *downloader) getVideoVariants(ctx context.Context, videoID string) ([]videoVariant, error) {
	fullURL, err := url.JoinPath(baseURL, videoAPI, videoID, "video_variants")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToConstructURL, err)
	}

	var variants []videoVariant
	if err := d.client.makeJSONRequest(ctx, fullURL, &variants); err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToDecodeVariants, err)
	}

	return variants, nil
}

// prepareDownloads checks which videos need to be downloaded and validates their availability.
// Returns indices of videos to download and longest filename width for alignment.
func (d *downloader) prepareDownloads(ctx context.Context, videos []models.Video, indices []int, failed *[]string) ([]int, int) {
	var (
		videosToDownload []int
		longestVideoName int
	)

	for _, idx := range indices {
		video := videos[idx]

		variants, err := d.getVideoVariants(ctx, video.ID)
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
		if dir.OverwriteVideoIfExists(filename, d.config) {
			videosToDownload = append(videosToDownload, idx)

			basename := filepath.Base(filename)
			longestVideoName = max(len(basename), longestVideoName)
		}
	}

	return videosToDownload, longestVideoName
}

// printResults displays the download results summary.
func (d *downloader) printResults(ctx context.Context, selectedCount int, failed []string) {
	if ctx.Err() != nil {
		fmt.Printf("\n%s Download aborted by user\n", styles.Error.Render("[ERROR]"))

		return
	}

	successCount := selectedCount - len(failed)
	fmt.Printf("\nDownload complete! %d/%d videos successful\n", successCount, selectedCount)

	if len(failed) > 0 {
		fmt.Printf("%s Failed downloads:\n", styles.Error.Render("[ERROR]"))

		for _, title := range failed {
			fmt.Printf("  - %s\n", title)
		}
	}
}

// processDownloads performs the actual video downloads in parallel.
// Returns slice of failed video titles.
func (d *downloader) processDownloads(ctx context.Context, videos []models.Video, indices []int, longestVideoName int) []string {
	var failed []string

	numVideos := len(indices)

	fmt.Print(ansi.HideCursor)

	for range numVideos {
		fmt.Println() // Reserve a line for each video
	}

	failed = d.downloadVideosParallel(ctx, videos, indices, longestVideoName)

	fmt.Print(ansi.ShowCursor)

	return failed
}

// Download initiates the download process based on the provided configuration.
// Extracts ID and type from media field, then downloads video or channel accordingly.
func Download(config models.DownloadConfig) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cancel context on SIGINT (Ctrl+C) for clean abort
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	defer signal.Stop(sigCh)

	go func() {
		<-sigCh
		cancel()
	}()

	id, downloadType, err := extractIDAndType(config.Media)
	if err != nil {
		return fmt.Errorf("%w: %w", errFailedToExtractType, err)
	}

	tokenMgr := token.NewTokenManager()

	client, err := newClient(tokenMgr)
	if err != nil {
		return err
	}

	downloader := newDownloader(config, client)

	switch downloadType {
	case videoType, unknownType:
		if err = downloader.downloadVideo(ctx, id, true, 0, 0); err == nil {
			return nil
		}

		if ctx.Err() != nil {
			return input.ErrUserAbort
		}

		if downloadType == videoType || errors.Is(err, dir.ErrFailedToCreateFile) {
			return fmt.Errorf("%w: %w", errFailedToDownloadVideo, err)
		}

		fallthrough // Fallthrough if type is unknown and try as channel
	case channelType:
		if err = downloader.downloadChannel(ctx, id); err != nil {
			if ctx.Err() != nil {
				return input.ErrUserAbort
			}

			if downloadType == unknownType {
				return fmt.Errorf("%w", errInvalidID)
			}

			return fmt.Errorf("%w: %w", errFailedToDownloadChannel, err)
		}
	}

	return nil
}

// extractIDAndType extracts the ID and determines if it's a video or channel.
// Returns ID, media type (video/channel/unknown), and error if URL is invalid.
func extractIDAndType(media string) (string, mediaType, error) {
	media = strings.TrimSpace(media)

	// If input doesn't start with baseURL, return as unknown type. This is the
	// case when the Id was passed as an argument
	prefixAndID, hasPrefix := strings.CutPrefix(media, baseURL)
	if !hasPrefix {
		return media, unknownType, nil
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
