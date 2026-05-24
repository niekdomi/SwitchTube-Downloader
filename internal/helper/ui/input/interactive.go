// Package input provides functions for interactive user input.
package input

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"

	"switchtube-downloader/internal/models"
)

// ErrUserAbort is returned when the user aborts an action (e.g. via Ctrl+C).
var ErrUserAbort = errors.New("aborted by user")

func buildVideoSelectForm(header string, options []huh.Option[int], selected *[]int) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[int]().
				Title("Choose videos to download").
				Description(header).
				Options(options...).
				Value(selected),
		),
	)
}

// SelectVideos shows an interactive multi-select for choosing videos.
// episodeColWidth controls the episode column width; 0 means no episode column is shown.
// Returns slice of selected video indices and error if user aborts.
func SelectVideos(videos []models.Video, episodeColWidth int) ([]int, error) {
	const (
		colSepSpace     = "  "
		rowPrefixIndent = "    "
	)

	options := make([]huh.Option[int], len(videos))
	for i, video := range videos {
		label := video.Title
		if episodeColWidth > 0 && video.Episode != "" {
			label = fmt.Sprintf("%-*s", episodeColWidth, video.Episode) + colSepSpace + video.Title
		}

		options[i] = huh.NewOption(label, i).Selected(true)
	}

	header := rowPrefixIndent
	if episodeColWidth > 0 {
		header += fmt.Sprintf("%-*s", episodeColWidth, "Episode") + colSepSpace
	}
	header += "Title"

	selected := make([]int, 0, len(videos))
	form := buildVideoSelectForm(header, options, &selected)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, ErrUserAbort
		}

		return nil, fmt.Errorf("failed to run selection form: %w", err)
	}

	return selected, nil
}
