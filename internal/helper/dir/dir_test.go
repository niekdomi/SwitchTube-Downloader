package dir

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"switchtube-downloader/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestIO(t *testing.T, input string) (func(), func() string) {
	t.Helper()

	tmpFile, err := os.CreateTemp(t.TempDir(), "test-input")
	require.NoError(t, err)

	if input != "" {
		_, err = tmpFile.WriteString(input)
		require.NoError(t, err)
		_, err = tmpFile.Seek(0, 0)
		require.NoError(t, err)
	}

	oldStdin := os.Stdin
	os.Stdin = tmpFile

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	restore := func() {
		_ = w.Close()

		os.Stdin = oldStdin
		os.Stdout = oldStdout

		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}

	captureOutput := func() string {
		output := make([]byte, 1000)
		n, _ := r.Read(output)

		return string(output[:n])
	}

	return restore, captureOutput
}

func TestCreateFilename(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		mediaType string
		episodeNr string
		want      string
		config    models.DownloadConfig
	}{
		{
			name:      "basic video",
			title:     "Test Video",
			mediaType: "video/mp4",
			episodeNr: "",
			config:    models.DownloadConfig{UseEpisode: false},
			want:      "Test_Video.mp4",
		},
		{
			name:      "video with episode number",
			title:     "Test Video",
			mediaType: "video/mp4",
			episodeNr: "E01",
			config:    models.DownloadConfig{UseEpisode: true},
			want:      "E01_Test_Video.mp4",
		},
		{
			name:      "invalid media type",
			title:     "Test Video",
			mediaType: "invalid",
			episodeNr: "",
			config:    models.DownloadConfig{UseEpisode: false},
			want:      "Test_Video.mp4",
		},
		{
			name:      "video with invalid characters",
			title:     "Test/Video:With*Invalid?Chars",
			mediaType: "video/mp4",
			episodeNr: "",
			config:    models.DownloadConfig{UseEpisode: false},
			want:      "Test-Video-WithInvalidChars.mp4",
		},
		{
			name:      "video with output path",
			title:     "Test Video",
			mediaType: "video/mp4",
			episodeNr: "",
			config:    models.DownloadConfig{OutputDir: "output", UseEpisode: false},
			want:      filepath.Join("output", "Test_Video.mp4"),
		},
		{
			name:      "empty title",
			title:     "",
			mediaType: "video/mp4",
			episodeNr: "",
			config:    models.DownloadConfig{UseEpisode: false},
			want:      ".mp4",
		},
		{
			name:      "title with only special characters",
			title:     "///:::***???",
			mediaType: "video/mp4",
			episodeNr: "",
			config:    models.DownloadConfig{UseEpisode: false},
			want:      "-.mp4",
		},
		{
			name:      "media type with more parts",
			title:     "Test Video",
			mediaType: "application/vnd.example+json",
			episodeNr: "",
			config:    models.DownloadConfig{UseEpisode: false},
			want:      "Test_Video.vnd.example+json",
		},
		{
			name:      "media type without slash",
			title:     "Test Video",
			mediaType: "mp4",
			episodeNr: "",
			config:    models.DownloadConfig{UseEpisode: false},
			want:      "Test_Video.mp4",
		},
		{
			name:      "episode with special characters",
			title:     "Test Video",
			mediaType: "video/mp4",
			episodeNr: "E01/S01",
			config:    models.DownloadConfig{UseEpisode: true},
			want:      "E01/S01_Test_Video.mp4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateFilename(tt.title, tt.mediaType, tt.episodeNr, tt.config)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOverwriteVideoIfExists(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		promptInput string
		config      models.DownloadConfig
		wantValue   bool
		createFile  bool
	}{
		{
			name:        "video exists, overwrite",
			filename:    "existing_video.mp4",
			config:      models.DownloadConfig{},
			promptInput: "y\n",
			wantValue:   true,
			createFile:  true,
		},
		{
			name:        "video exists, do not overwrite",
			filename:    "existing_video.mp4",
			config:      models.DownloadConfig{},
			promptInput: "\n",
			wantValue:   false,
			createFile:  true,
		},
		{
			name:        "video does not exist",
			filename:    "non_existing_video.mp4",
			config:      models.DownloadConfig{},
			promptInput: "",
			wantValue:   true,
			createFile:  false,
		},
		{
			name:        "video exists, force-flag set",
			filename:    "existing_video.mp4",
			config:      models.DownloadConfig{Force: true},
			promptInput: "",
			wantValue:   true,
			createFile:  true,
		},
		{
			name:        "video does not exist, force-flag set",
			filename:    "non_existing_video.mp4",
			config:      models.DownloadConfig{Force: true},
			promptInput: "",
			wantValue:   true,
			createFile:  false,
		},
		{
			name:        "video exists, skip-flag set",
			filename:    "existing_video.mp4",
			config:      models.DownloadConfig{Skip: true},
			promptInput: "",
			wantValue:   false,
			createFile:  true,
		},
		{
			name:        "video does not exist, skip-flag set",
			filename:    "non_existing_video.mp4",
			config:      models.DownloadConfig{Skip: true},
			promptInput: "",
			wantValue:   true,
			createFile:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			filename := filepath.Join(tempDir, tt.filename)

			if tt.createFile {
				require.NoError(t, os.MkdirAll(filepath.Dir(filename), dirPermissions))
				_, err := os.Create(filename)
				require.NoError(t, err)
			}

			var got bool

			if tt.promptInput != "" {
				restore, captureOutput := setupTestIO(t, tt.promptInput)
				defer restore()

				got = OverwriteVideoIfExists(filename, tt.config)
				output := captureOutput()

				adjustedOutput := strings.ReplaceAll(output, tempDir+string(os.PathSeparator), "")
				assert.Contains(t, adjustedOutput, "File")
				assert.Contains(t, adjustedOutput, "already exists")
			} else {
				got = OverwriteVideoIfExists(filename, tt.config)
			}

			assert.Equal(t, tt.wantValue, got)
		})
	}
}

func TestCreateVideoFile(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		wantErr    bool
		createFile bool
	}{
		{
			name:       "create new video",
			filename:   "test_video.mp4",
			wantErr:    false,
			createFile: false,
		},
		{
			name:       "create video in subdirectory",
			filename:   filepath.Join("sub", "test_video.mp4"),
			wantErr:    false,
			createFile: false,
		},
		{
			name:       "video already exists",
			filename:   "existing_video.mp4",
			wantErr:    false,
			createFile: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			filename := filepath.Join(tempDir, tt.filename)

			if tt.createFile {
				require.NoError(t, os.MkdirAll(filepath.Dir(filename), dirPermissions))
				_, err := os.Create(filename)
				require.NoError(t, err)
			}

			fd, err := CreateVideoFile(filename)
			if fd != nil {
				defer fd.Close()
			}

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				_, err = os.Stat(filename)
				require.NoError(t, err)
			}
		})
	}
}

func TestCreateChannelFolder(t *testing.T) {
	tests := []struct {
		name        string
		channelName string
		outputPath  string
		wantFolder  string
		wantErr     bool
	}{
		{
			name:        "basic folder creation",
			channelName: "Test Channel",
			outputPath:  "",
			wantFolder:  "Test Channel",
			wantErr:     false,
		},
		{
			name:        "folder with slashes",
			channelName: "Test/Channel",
			outputPath:  "output",
			wantFolder:  filepath.Join("output", "Test - Channel"),
			wantErr:     false,
		},
		{
			name:        "empty channel name",
			channelName: "",
			outputPath:  "",
			wantFolder:  ".",
			wantErr:     false,
		},
		{
			name:        "folder in specific output directory",
			channelName: "Test Channel",
			outputPath:  "test_path",
			wantFolder:  filepath.Join("test_path", "Test Channel"),
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			config := models.DownloadConfig{
				OutputDir: filepath.Join(tempDir, tt.outputPath),
			}

			folder, err := CreateChannelFolder(tt.channelName, config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				expectedFolder := filepath.Join(tempDir, tt.wantFolder)
				assert.Equal(t, expectedFolder, folder)
				_, err = os.Stat(folder)
				require.NoError(t, err)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "basic sanitization",
			in:   "Test Video",
			want: "Test Video",
		},
		{
			name: "invalid characters",
			in:   "Test/Video:With*Invalid?Chars<>\\|",
			want: "Test-Video-WithInvalidChars-",
		},
		{
			name: "multiple dashes",
			in:   "Test///Video",
			want: "Test-Video",
		},
		{
			name: "leading and trailing spaces",
			in:   "  Test Video  ",
			want: "Test Video",
		},
		{
			name: "empty string",
			in:   "",
			want: "",
		},
		{
			name: "only invalid characters",
			in:   "///:::***???<>|",
			want: "-",
		},
		{
			name: "unicode characters",
			in:   "Test vidéo ñoño.mp4",
			want: "Test vidéo ñoño.mp4",
		},
		{
			name: "very long string",
			in:   strings.Repeat("a", 300) + "/\\:*?\"<>|",
			want: strings.Repeat("a", 300) + "-",
		},
		{
			name: "consecutive invalid characters",
			in:   "test////video",
			want: "test-video",
		},
		{
			name: "mixed valid and invalid",
			in:   "test/video:name*question?mark<greater>less|pipe\\backslash",
			want: "test-video-namequestionmarkgreaterless-pipe-backslash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, sanitizeFilename(tt.in))
		})
	}
}
