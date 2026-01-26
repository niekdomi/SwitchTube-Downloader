package input

import (
	"errors"
	"fmt"
	"os"

	"switchtube-downloader/internal/helper/ui/ansi"
	"switchtube-downloader/internal/helper/ui/terminal"
	"switchtube-downloader/internal/models"
)

var (
	// ErrUserAbort is returned when the user aborts an action (e.g. via Ctrl+C).
	ErrUserAbort = errors.New("aborted by user")

	errFailedToReadKey = errors.New("failed to read key")
)

// selectionState holds the state of the interactive selection UI.
type selectionState struct {
	videos       []models.Video
	selected     []bool
	currentIndex int
	useEpisode   bool
}

// newSelectionState creates a new selection state with all items selected by default.
func newSelectionState(videos []models.Video, useEpisode bool) *selectionState {
	selected := make([]bool, len(videos))
	for i := range selected {
		selected[i] = true
	}

	return &selectionState{
		videos:       videos,
		selected:     selected,
		currentIndex: 0,
		useEpisode:   useEpisode,
	}
}

// SelectVideos shows an interactive checkbox-based selector. All items are selected by default.
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
	state.render(false)

	return runEventLoop(state)
}

// getSelectedIndices returns the indices of all selected items.
func (s *selectionState) getSelectedIndices() []int {
	indices := make([]int, 0, len(s.selected))
	for i, sel := range s.selected {
		if sel {
			indices = append(indices, i)
		}
	}

	return indices
}

// handleEvent processes a keyboard event and returns whether to render and exit.
func (s *selectionState) handleEvent(event Event) (bool, bool, error) {
	switch event.Key { //nolint:exhaustive
	case KeyArrowUp:
		s.moveUp()
	case KeyArrowDown:
		s.moveDown()
	case KeySpace:
		s.toggleCurrent()
	case KeyEnter:
		return false, true, nil
	case KeyCtrlC:
		return false, true, ErrUserAbort
	}

	return true, false, nil
}

// initializeTerminal sets up the terminal for interactive input.
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
func (s *selectionState) render(isUpdate bool) {
	if isUpdate {
		fmt.Printf(ansi.MoveCursorUp, len(s.videos)+1) // Move cursor up to the top of the list
	}

	fmt.Print("\r" + ansi.ClearLine)
	fmt.Printf("%s%sChoose videos to download:%s\n", ansi.Bold, ansi.Cyan, ansi.Reset)

	longestEpisodeName := 0

	if s.useEpisode {
		for _, video := range s.videos {
			longestEpisodeName = max(len(video.Episode), longestEpisodeName)
		}
	}

	for i, video := range s.videos {
		renderVideoItem(video, s.selected[i], i == s.currentIndex, s.useEpisode, longestEpisodeName)
	}

	fmt.Print("\r" + ansi.ClearLine)
	fmt.Printf("%sNavigation: ↑↓/j/k  Toggle: Space  Confirm: Enter%s", ansi.Dim, ansi.Reset)

	_ = os.Stdout.Sync()
}

// renderVideoItem displays a single video item.
func renderVideoItem(video models.Video, isSelected bool, isCurrent bool, useEpisode bool, maxEpisodeWidth int) {
	fmt.Print("\r" + ansi.ClearLine)

	checkbox := ansi.CheckboxUnchecked
	checkboxColor := ""

	if isSelected {
		checkbox = ansi.CheckboxChecked
		checkboxColor = ansi.Green
	}

	videoText := video.Title
	if useEpisode {
		videoText = fmt.Sprintf("%-*s %s", maxEpisodeWidth, video.Episode, video.Title)
	}

	if isCurrent {
		fmt.Printf("  %s%s%s %s%s%s\n", checkboxColor, checkbox, ansi.Reset, ansi.Bold, videoText, ansi.Reset)
	} else {
		fmt.Printf("  %s%s%s %s%s%s\n", checkboxColor, checkbox, ansi.Reset, ansi.Dim, videoText, ansi.Reset)
	}
}

// runEventLoop processes keyboard input until the user confirms or cancels.
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

		if shouldExit {
			fmt.Println()

			return state.getSelectedIndices(), nil
		}

		if shouldRender {
			state.render(true)
		}
	}
}

// toggleCurrent toggles the selection of the current item and moves to the next.
func (s *selectionState) toggleCurrent() {
	s.selected[s.currentIndex] = !s.selected[s.currentIndex]
	s.currentIndex = (s.currentIndex + 1) % len(s.videos)
}
