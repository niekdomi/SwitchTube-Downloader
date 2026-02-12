package input

import (
	"errors"
	"fmt"
	"os"

	"switchtube-downloader/internal/helper/ui/ansi"
	"switchtube-downloader/internal/helper/ui/terminal"
	"switchtube-downloader/internal/models"
)

// ErrUserAbort is returned when the user aborts an action (e.g. via Ctrl+C).
var ErrUserAbort = errors.New("aborted by user")

// selectionState holds the state of the interactive selection UI.
type selectionState struct {
	videos           []models.Video // All videos to choose from
	selected         []bool         // Selection status for each video
	currentIndex     int            // Currently highlighted index
	useEpisode       bool           // Whether to show episode numbers
	highlightEnabled bool           // Whether highlighting is enabled
	rendered         bool           // Track if initial render has occurred
}

// newSelectionState creates a new selection state with all videos selected by default.
func newSelectionState(videos []models.Video, useEpisode bool) *selectionState {
	selectedVideo := make([]bool, len(videos))

	for i := range selectedVideo {
		selectedVideo[i] = true
	}

	return &selectionState{
		videos:           videos,
		selected:         selectedVideo,
		currentIndex:     0,
		useEpisode:       useEpisode,
		highlightEnabled: true,
		rendered:         false,
	}
}

// SelectVideos shows an interactive checkbox-based selector for choosing videos.
// Returns slice of selected video indices and error if user aborts or terminal fails.
func SelectVideos(videos []models.Video, all bool, useEpisode bool) ([]int, error) {
	// If --all flag is used, select all videos
	if all || len(videos) == 0 {
		indices := make([]int, len(videos))

		for i := range indices {
			indices[i] = i
		}

		return indices, nil
	}

	termState, err := initializeTerminal()
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = termState.Restore()
		fmt.Print(ansi.ShowCursor)
	}()

	state := newSelectionState(videos, useEpisode)
	state.render()

	return runEventLoop(state)
}

// getSelectedIndices returns the indices of all selected videos.
func (s *selectionState) getSelectedIndices() []int {
	indices := make([]int, 0, len(s.selected))

	for i, selectedVideo := range s.selected {
		if selectedVideo {
			indices = append(indices, i)
		}
	}

	return indices
}

// handleEvent processes a keyboard event.
// Returns shouldRender, shouldExit flag and error if user aborts.
func (s *selectionState) handleEvent(key Key) (bool, bool, error) {
	switch key {
	case KeyArrowUp:
		s.moveUp()
	case KeyArrowDown:
		s.moveDown()
	case KeySpace:
		s.toggleCurrent()
	case KeyEnter:
		s.highlightEnabled = false

		return true, true, nil
	case KeyCtrlC:
		return false, true, ErrUserAbort
	case KeyUnknown:
		return false, false, nil
	}

	return true, false, nil
}

// initializeTerminal sets up the terminal for interactive input.
// Returns terminal state for restoration and error if setup fails.
func initializeTerminal() (*terminal.State, error) {
	termState, err := terminal.EnableRawMode()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", terminal.ErrFailedToSetRawMode, err)
	}

	fmt.Print(ansi.HideCursor)

	return termState, nil
}

// moveDown moves the cursor down by one position.
func (s *selectionState) moveDown() {
	s.currentIndex = (s.currentIndex + 1) % len(s.videos)
}

// moveUp moves the cursor up by one position.
func (s *selectionState) moveUp() {
	s.currentIndex = (s.currentIndex - 1 + len(s.videos)) % len(s.videos)
}

// render displays the current selection state.
func (s *selectionState) render() {
	if s.rendered {
		// Move cursor up to the top of the list
		fmt.Printf(ansi.MoveCursorUp, len(s.videos))
	} else {
		fmt.Print("\r" + ansi.ClearLine)
		fmt.Printf("%s%sChoose videos to download:%s\n", ansi.Bold, ansi.Cyan, ansi.Reset)
	}

	longestEpisodeName := 0

	if s.useEpisode {
		for _, video := range s.videos {
			longestEpisodeName = max(len(video.Episode), longestEpisodeName)
		}
	}

	for i, video := range s.videos {
		isCurrent := s.highlightEnabled && (i == s.currentIndex)
		s.renderVideo(video, s.selected[i], isCurrent, longestEpisodeName)
	}

	if !s.rendered {
		fmt.Print("\r" + ansi.ClearLine)
		fmt.Printf("%sNavigation: ↑↓/j/k  Toggle: Space  Confirm: Enter%s", ansi.Dim, ansi.Reset)

		s.rendered = true
	}

	_ = os.Stdout.Sync()
}

// renderVideo displays a single video in the selection.
func (s *selectionState) renderVideo(video models.Video, isSelected bool, isCurrent bool, maxEpisodeWidth int) {
	fmt.Print("\r" + ansi.ClearLine)

	checkbox := ansi.CheckboxUnchecked
	checkboxColor := ""

	if isSelected {
		checkbox = ansi.CheckboxChecked
		checkboxColor = ansi.Green
	}

	videoText := video.Title
	if s.useEpisode {
		videoText = fmt.Sprintf("%-*s %s", maxEpisodeWidth, video.Episode, video.Title)
	}

	style := ansi.Dim
	if isCurrent {
		style = ansi.Bold
	}

	fmt.Printf("  %s%s%s %s%s%s\n", checkboxColor, checkbox, ansi.Reset, style, videoText, ansi.Reset)
}

// toggleCurrent toggles the selection of the current video and moves to the next.
func (s *selectionState) toggleCurrent() {
	s.selected[s.currentIndex] = !s.selected[s.currentIndex]
	s.currentIndex = (s.currentIndex + 1) % len(s.videos)
}

// runEventLoop processes keyboard input until the user confirms or cancels.
// Returns selected video indices and error if user aborts or key reading fails.
func runEventLoop(state *selectionState) ([]int, error) {
	for {
		event, err := ReadKey()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errFailedToReadKey, err)
		}

		shouldRender, shouldExit, err := state.handleEvent(event)
		if err != nil {
			return nil, err
		}

		if shouldRender {
			state.render()
		}

		if shouldExit {
			fmt.Println()

			return state.getSelectedIndices(), nil
		}
	}
}
