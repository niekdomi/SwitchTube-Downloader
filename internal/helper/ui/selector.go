package ui

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"switchtube-downloader/internal/models"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

const rangePartsCount = 2

var (
	errInvalidRange           = errors.New("invalid range")
	errInvalidNumber          = errors.New("invalid number")
	errInvalidEndNumber       = errors.New("invalid end number")
	errNumberOutOfRange       = errors.New("number out of range")
	errInvalidRangeFormat     = errors.New("invalid range format")
	errInvalidStartNumber     = errors.New("invalid start number")
	errNoValidSelectionsFound = errors.New("no valid selections found")
)

// SelectVideos displays the video list and handles user selection.
func SelectVideos(videos []models.Video, all bool, useEpisode bool) ([]int, error) {
	// If --all flag is used, select all videos
	if all || len(videos) == 0 {
		indices := make([]int, len(videos))
		for i := range indices {
			indices[i] = i
		}

		return indices, nil
	}

	// Use interactive selection if running in a terminal
	if IsTerminal() {
		return interactiveSelect(videos, useEpisode)
	}

	// Fall back to text-based selection for non-TTY (piped input, etc.)
	if err := renderVideoTable(videos, useEpisode); err != nil {
		return nil, err
	}

	fmt.Println("\nSelect videos:")
	fmt.Println("   • Single: '1' or '3,5,7'")
	fmt.Println("   • Range:  '1-5' or '1-3,7-9'")
	fmt.Println("   • All:    Press Enter")

	input := strings.TrimSpace(Input("\nSelection: "))
	if input == "" {
		// If input is empty, select all videos
		indices := make([]int, len(videos))
		for i := range indices {
			indices[i] = i
		}

		return indices, nil
	}

	return parseSelection(input, len(videos))
}

// renderVideoTable renders the video selection table.
func renderVideoTable(videos []models.Video, useEpisode bool) error {
	data := make([][]string, len(videos))
	for i, video := range videos {
		if useEpisode {
			data[i] = []string{strconv.Itoa(i + 1), video.Episode, video.Title}
		} else {
			data[i] = []string{strconv.Itoa(i + 1), video.Title}
		}
	}

	var aligns []tw.Align
	if useEpisode {
		aligns = []tw.Align{tw.AlignRight, tw.AlignRight, tw.AlignLeft}
	} else {
		aligns = []tw.Align{tw.AlignRight, tw.AlignLeft}
	}

	table := tablewriter.NewTable(
		os.Stdout,
		//nolint:exhaustruct
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Alignment: tw.CellAlignment{
					Global:    tw.AlignCenter,
					PerColumn: aligns,
				},
			},
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{
					Global:    tw.AlignLeft,
					PerColumn: aligns,
				},
				// Auto-wrap long titles
				Formatting: tw.CellFormatting{
					AutoWrap: tw.WrapNormal,
				},
			},
		}),
	)

	if useEpisode {
		table.Header("Number", "Episode", "Title")
	} else {
		table.Header("Number", "Title")
	}

	if err := table.Bulk(data); err != nil {
		return fmt.Errorf("failed to add table data: %w", err)
	}

	if err := table.Render(); err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}

// parseSelection parses user input and returns selected video indices.
func parseSelection(input string, availableVideos int) ([]int, error) {
	var indices []int

	seen := make(map[int]bool)

	// Split by comma, space, or both
	parts := strings.FieldsFunc(input, func(r rune) bool {
		return r == ',' || r == ' '
	})

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		var err error
		// Handle range (e.g., "1-5")
		if strings.Contains(part, "-") {
			indices, err = handleRangeSelection(part, availableVideos, indices, seen)
			if err != nil {
				return nil, err
			}
		} else {
			indices, err = handleSingleSelection(part, availableVideos, indices, seen)
			if err != nil {
				return nil, err
			}
		}
	}

	if len(indices) == 0 {
		return nil, fmt.Errorf("%w", errNoValidSelectionsFound)
	}

	sort.Ints(indices)

	return indices, nil
}

// handleRangeSelection processes a range selection like "1-5".
func handleRangeSelection(part string, availableVideos int, indices []int, seen map[int]bool) ([]int, error) {
	rangeParts := strings.Split(part, "-")
	if len(rangeParts) != rangePartsCount {
		return nil, fmt.Errorf("%w: %s", errInvalidRangeFormat, part)
	}

	start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errInvalidStartNumber, rangeParts[0])
	}

	end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errInvalidEndNumber, rangeParts[1])
	}

	if start < 1 || end > availableVideos || start > end {
		return nil, fmt.Errorf(
			"%w: %d-%d (must be 1-%d)",
			errInvalidRange,
			start,
			end,
			availableVideos,
		)
	}

	for i := start; i <= end; i++ {
		index := i - 1
		if !seen[index] {
			indices = append(indices, index)
			seen[index] = true
		}
	}

	return indices, nil
}

// handleSingleSelection processes a single number selection.
func handleSingleSelection(part string, availableVideos int, indices []int, seen map[int]bool) ([]int, error) {
	num, err := strconv.Atoi(part)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errInvalidNumber, part)
	}

	if num < 1 || num > availableVideos {
		return nil, fmt.Errorf("%w: %d (must be 1-%d)", errNumberOutOfRange, num, availableVideos)
	}

	index := num - 1
	if !seen[index] {
		indices = append(indices, index)
		seen[index] = true
	}

	return indices, nil
}
