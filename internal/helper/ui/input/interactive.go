package input

import (
	"errors"
	"fmt"
	"os"

	"switchtube-downloader/internal/helper/ui/colors"
	"switchtube-downloader/internal/helper/ui/terminal"
	"switchtube-downloader/internal/models"
)

var errFailedToReadKey = errors.New("failed to read key")

// SelectionState holds the state of the interactive selection UI.
type SelectionState struct {
	videos       []models.Video
	selected     []bool
	currentIndex int
	useEpisode   bool
}

// newSelectionState creates a new selection state with all items selected by default.
func newSelectionState(videos []models.Video, useEpisode bool) *SelectionState {
	selected := make([]bool, len(videos))
	for i := range selected {
		selected[i] = true
	}

	return &SelectionState{
		videos:       videos,
		selected:     selected,
		currentIndex: 0,
		useEpisode:   useEpisode,
	}
}

// getSelectedIndices returns the indices of all selected items.
func (s *SelectionState) getSelectedIndices() []int {
	indices := make([]int, 0, len(s.selected))
	for i, sel := range s.selected {
		if sel {
			indices = append(indices, i)
		}
	}

	return indices
}

// handleEvent processes a keyboard event and returns whether to render and exit.
func (s *SelectionState) handleEvent(event Event) (bool, bool) {
	switch event.Key { //nolint:exhaustive
	case KeyArrowUp:
		s.moveUp()
	case KeyArrowDown:
		s.moveDown()
	case KeySpace:
		s.toggleCurrent()
	case KeyEnter:
		return false, true
	case KeyCtrlC:
		fmt.Println()
		os.Exit(0)

		return false, false
	}

	return true, false
}

// initializeTerminal sets up the terminal for interactive input.
func initializeTerminal() (*terminal.State, error) {
	termState, err := terminal.EnableRawMode()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", terminal.ErrFailedToSetRawMode, err)
	}

	fmt.Print(colors.HideCursor)

	return termState, nil
}

// interactiveSelect shows an interactive checkbox-based selector.
// All items are selected by default.
func interactiveSelect(videos []models.Video, useEpisode bool) ([]int, error) {
	termState, err := initializeTerminal()
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = termState.Restore()

		fmt.Print(colors.ShowCursor)
	}()

	state := newSelectionState(videos, useEpisode)
	state.render(false)

	return runEventLoop(state)
}

// moveDown moves the cursor down by one position.
func (s *SelectionState) moveDown() {
	s.currentIndex = (s.currentIndex + 1) % len(s.videos)
}

// moveUp moves the cursor up by one position.
func (s *SelectionState) moveUp() {
	s.currentIndex = (s.currentIndex - 1 + len(s.videos)) % len(s.videos)
}

// render displays the current selection state.
func (s *SelectionState) render(isUpdate bool) {
	if isUpdate {
		fmt.Printf("\033[%dA", len(s.videos)+1) // Move cursor up to the start of the list
	}

	fmt.Print("\r" + colors.ClearLine)
	fmt.Printf("%s%sChoose videos to download:%s\n", colors.Bold, colors.Cyan, colors.Reset)

	for i, video := range s.videos {
		renderVideoItem(video, s.selected[i], i == s.currentIndex, s.useEpisode)
	}

	fmt.Print("\r" + colors.ClearLine)
	fmt.Printf("%sNavigation: ↑↓/j/k  Toggle: Space  Confirm: Enter%s", colors.Dim, colors.Reset)

	_ = os.Stdout.Sync()
}

// renderVideoItem displays a single video item.
func renderVideoItem(video models.Video, isSelected bool, isCurrent bool, useEpisode bool) {
	fmt.Print("\r" + colors.ClearLine)

	checkbox := colors.CheckboxUnchecked
	checkboxColor := ""

	if isSelected {
		checkbox = colors.CheckboxChecked
		checkboxColor = colors.Green
	}

	videoText := video.Title
	if useEpisode {
		videoText = fmt.Sprintf("%s %s", video.Episode, video.Title)
	}

	if isCurrent {
		fmt.Printf("  %s%s%s %s%s%s\n", checkboxColor, checkbox, colors.Reset, colors.Bold, videoText, colors.Reset)
	} else {
		fmt.Printf("  %s%s%s %s%s%s\n", checkboxColor, checkbox, colors.Reset, colors.Dim, videoText, colors.Reset)
	}
}

// runEventLoop processes keyboard input until the user confirms or cancels.
func runEventLoop(state *SelectionState) ([]int, error) {
	for {
		event, err := ReadKey()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errFailedToReadKey, err)
		}

		shouldRender, shouldExit := state.handleEvent(event)

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
func (s *SelectionState) toggleCurrent() {
	s.selected[s.currentIndex] = !s.selected[s.currentIndex]
	s.currentIndex = (s.currentIndex + 1) % len(s.videos)
}
