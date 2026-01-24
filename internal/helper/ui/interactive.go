package ui

import (
	"errors"
	"fmt"
	"os"

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
func (s *SelectionState) handleEvent(event InputEvent) (bool, bool) {
	switch event.Key { //nolint:exhaustive
	case KeyArrowUp:
		return s.moveUp(), false
	case KeyArrowDown:
		return s.moveDown(), false
	case KeySpace:
		s.toggleCurrent()

		return true, false
	case KeyEnter:
		return false, true
	case KeyCtrlC:
		fmt.Println()
		os.Exit(0)

		return false, false
	default:
		return false, false
	}
}

// initializeTerminal sets up the terminal for interactive input.
func initializeTerminal() (*TerminalState, error) {
	termState, err := EnableRawMode()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFailedToSetRawMode, err)
	}

	fmt.Print(HideCursor)

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

		fmt.Print(ShowCursor)
	}()

	state := newSelectionState(videos, useEpisode)
	state.render(false)

	return runEventLoop(state)
}

// moveDown moves the cursor down by one position.
func (s *SelectionState) moveDown() bool {
	if s.currentIndex < len(s.videos)-1 {
		s.currentIndex++

		return true
	}

	return false
}

// moveUp moves the cursor up by one position.
func (s *SelectionState) moveUp() bool {
	if s.currentIndex > 0 {
		s.currentIndex--

		return true
	}

	return false
}

// render displays the current selection state.
func (s *SelectionState) render(isUpdate bool) {
	if isUpdate {
		fmt.Printf("\033[%dA", len(s.videos)+1) // Move cursor up to the start of the list
	}

	fmt.Print("\r" + ClearLine)
	fmt.Printf("%s%sChoose videos to download:%s\n", Bold, Cyan, Reset)

	for i, video := range s.videos {
		renderVideoItem(video, s.selected[i], i == s.currentIndex, s.useEpisode)
	}

	fmt.Print("\r" + ClearLine)
	fmt.Printf("%sNavigation: ↑↓/j/k  Toggle: Space  Confirm: Enter%s", Dim, Reset)

	_ = os.Stdout.Sync()
}

// renderVideoItem displays a single video item.
func renderVideoItem(video models.Video, isSelected bool, isCurrent bool, useEpisode bool) {
	fmt.Print("\r" + ClearLine)

	checkbox := CheckboxUnchecked
	checkboxColor := ""

	if isSelected {
		checkbox = CheckboxChecked
		checkboxColor = Green
	}

	videoText := video.Title
	if useEpisode {
		videoText = fmt.Sprintf("%s %s", video.Episode, video.Title)
	}

	if isCurrent {
		fmt.Printf("  %s%s%s %s%s%s\n", checkboxColor, checkbox, Reset, Bold, videoText, Reset)
	} else {
		fmt.Printf("  %s%s%s %s%s%s\n", checkboxColor, checkbox, Reset, Dim, videoText, Reset)
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
