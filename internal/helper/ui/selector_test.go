package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"switchtube-downloader/internal/models"
)

func TestSelectVideos(t *testing.T) {
	tests := []struct {
		name       string
		videos     []models.Video
		all        bool
		useEpisode bool
		input      string
		want       []int
		wantErr    bool
		err        error
		wantPrompt string
	}{
		{
			name:       "select all with --all flag",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:        true,
			useEpisode: false,
			input:      "",
			want:       []int{0, 1},
			wantErr:    false,
			wantPrompt: "",
		},
		{
			name:       "select all with empty input",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:        false,
			useEpisode: false,
			input:      "\n",
			want:       []int{0, 1},
			wantErr:    false,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n│      2 │ Video2 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:       "select single video",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:        false,
			useEpisode: false,
			input:      "1\n",
			want:       []int{0},
			wantErr:    false,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n│      2 │ Video2 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:       "select range",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}, {Title: "Video3"}},
			all:        false,
			useEpisode: false,
			input:      "1-3\n",
			want:       []int{0, 1, 2},
			wantErr:    false,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n│      2 │ Video2 │\n│      3 │ Video3 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:       "select multiple videos with comma",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}, {Title: "Video3"}},
			all:        false,
			useEpisode: false,
			input:      "1, 3\n",
			want:       []int{0, 2},
			wantErr:    false,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n│      2 │ Video2 │\n│      3 │ Video3 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:       "select multiple videos with space",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}, {Title: "Video3"}},
			all:        false,
			useEpisode: false,
			input:      "1 3\n",
			want:       []int{0, 2},
			wantErr:    false,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n│      2 │ Video2 │\n│      3 │ Video3 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:       "invalid number",
			videos:     []models.Video{{Title: "Video1"}},
			all:        false,
			useEpisode: false,
			input:      "abc\n",
			want:       nil,
			wantErr:    true,
			err:        errInvalidNumber,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:       "number out of range",
			videos:     []models.Video{{Title: "Video1"}},
			all:        false,
			useEpisode: false,
			input:      "2\n",
			want:       nil,
			wantErr:    true,
			err:        errNumberOutOfRange,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:       "invalid range format",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:        false,
			useEpisode: false,
			input:      "1-2-3\n",
			want:       nil,
			wantErr:    true,
			err:        errInvalidRangeFormat,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n│      2 │ Video2 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:       "invalid start number in range",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:        false,
			useEpisode: false,
			input:      "x-2\n",
			want:       nil,
			wantErr:    true,
			err:        errInvalidStartNumber,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n│      2 │ Video2 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:       "invalid end number in range",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:        false,
			useEpisode: false,
			input:      "1-y\n",
			want:       nil,
			wantErr:    true,
			err:        errInvalidEndNumber,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n│      2 │ Video2 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:       "range out of bounds",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:        false,
			useEpisode: false,
			input:      "1-3\n",
			want:       nil,
			wantErr:    true,
			err:        errInvalidRange,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n│      2 │ Video2 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:       "start greater than end in range",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:        false,
			useEpisode: false,
			input:      "2-1\n",
			want:       nil,
			wantErr:    true,
			err:        errInvalidRange,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n│      2 │ Video2 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:       "no valid selections",
			videos:     []models.Video{{Title: "Video1"}},
			all:        false,
			useEpisode: false,
			input:      ",,,\n",
			want:       nil,
			wantErr:    true,
			err:        errNoValidSelectionsFound,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:       "empty video list",
			videos:     []models.Video{},
			all:        false,
			useEpisode: false,
			input:      "",
			want:       []int{},
			wantErr:    false,
			wantPrompt: "",
		},
		{
			name:       "duplicate selections",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:        false,
			useEpisode: false,
			input:      "1,1,1-2,2\n",
			want:       []int{0, 1},
			wantErr:    false,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n│      2 │ Video2 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:       "range with same start and end",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:        false,
			useEpisode: false,
			input:      "1-1\n",
			want:       []int{0},
			wantErr:    false,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n│      2 │ Video2 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:       "mixed valid and invalid selections",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}, {Title: "Video3"}},
			all:        false,
			useEpisode: false,
			input:      "1,abc,3,5\n",
			want:       nil,
			wantErr:    true,
			err:        errInvalidNumber,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n│      2 │ Video2 │\n│      3 │ Video3 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:       "negative numbers",
			videos:     []models.Video{{Title: "Video1"}},
			all:        false,
			useEpisode: false,
			input:      "-1\n",
			want:       nil,
			wantErr:    true,
			err:        errInvalidStartNumber,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:       "zero as input",
			videos:     []models.Video{{Title: "Video1"}},
			all:        false,
			useEpisode: false,
			input:      "0\n",
			want:       nil,
			wantErr:    true,
			err:        errNumberOutOfRange,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:       "input with extra whitespace",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:        false,
			useEpisode: false,
			input:      "  1  ,  2  \n",
			want:       []int{0, 1},
			wantErr:    false,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n│      2 │ Video2 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:       "range with spaces around dash",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}, {Title: "Video3"}},
			all:        false,
			useEpisode: false,
			input:      "1 - 3\n",
			want:       nil,
			wantErr:    true,
			err:        errInvalidStartNumber,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n│      2 │ Video2 │\n│      3 │ Video3 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:    "single number with leading/trailing spaces",
			videos:  []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:     false,
			input:   "  2  \n",
			want:    []int{1},
			wantErr: false,
			wantPrompt: "┌────────┬────────┐\n│ NUMBER │ VIDEO  │\n├────────┼────────┤\n│      1 │ Video1 │\n│      2 │ Video2 │\n└────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name: "select with episode",
			videos: []models.Video{
				{Title: "Video1", Episode: "01"},
				{Title: "Video2", Episode: "02"},
			},
			all:        false,
			useEpisode: true,
			input:      "1\n",
			want:       []int{0},
			wantErr:    false,
			wantPrompt: "┌────────┬─────────┬────────┐\n│ NUMBER │ EPISODE │ VIDEO  │\n├────────┼─────────┼────────┤\n│      1 │      01 │ Video1 │\n│      2 │      02 │ Video2 │\n└────────┴─────────┴────────┘\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []int

			var err error

			if tt.wantPrompt != "" {
				restore, readOutput := SetupTestIO(t, tt.input)
				defer restore()

				result, err = SelectVideos(tt.videos, tt.all, tt.useEpisode)
				capturedOutput := readOutput()

				assert.Equal(t, tt.wantPrompt, capturedOutput)
			} else {
				result, err = SelectVideos(tt.videos, tt.all, tt.useEpisode)
			}

			assert.Equal(t, tt.want, result)

			if tt.wantErr {
				assert.ErrorIs(t, err, tt.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
