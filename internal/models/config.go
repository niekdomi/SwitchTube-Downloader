// Package models defines the structures used in the application.
package models

// DownloadConfig holds configuration options for the Download function.
type DownloadConfig struct {
	Media      string // Video or channel ID/URL
	OutputDir  string // Output directory
	UseEpisode bool   // Whether to use episode numbers in filenames
	Skip       bool   // Whether to skip existing files
	Force      bool   // Whether to force overwrite existing files
	All        bool   // Whether to download all videos
}
