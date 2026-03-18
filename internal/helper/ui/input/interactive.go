package input

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"

	"switchtube-downloader/internal/models"
)

// ErrUserAbort is returned when the user aborts an action (e.g. via Ctrl+C).
var ErrUserAbort = errors.New("aborted by user")

// SelectVideos shows an interactive multi-select for choosing videos.
// Returns slice of selected video indices and error if user aborts.
func SelectVideos(videos []models.Video, all bool, useEpisode bool) ([]int, error) {
	// If --all flag is used, select all videos
	if all || len(videos) == 0 {
		indices := make([]int, len(videos))

		for i := range indices {
			indices[i] = i
		}

		return indices, nil
	}

	options := make([]huh.Option[int], len(videos))
	for i, video := range videos {
		label := video.Title
		if useEpisode && video.Episode != "" {
			label = video.Episode + "  " + video.Title
		}

		options[i] = huh.NewOption(label, i).Selected(true)
	}

	selected := make([]int, 0, len(videos))

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[int]().
				Title("Choose videos to download").
				Options(options...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, ErrUserAbort
		}

		return nil, fmt.Errorf("failed to run selection form: %w", err)
	}

	return selected, nil
}
