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
			input:      "",
			want:       []int{0, 1},
			wantErr:    false,
			wantPrompt: "",
		},
		{
			name:    "select all with empty input",
			videos:  []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:     false,
			input:   "\n",
			want:    []int{0, 1},
			wantErr: false,
			wantPrompt: "\nAvailable videos:\n1. Video1\n2. Video2\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:    "select single video",
			videos:  []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:     false,
			input:   "1\n",
			want:    []int{0},
			wantErr: false,
			wantPrompt: "\nAvailable videos:\n1. Video1\n2. Video2\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:    "select range",
			videos:  []models.Video{{Title: "Video1"}, {Title: "Video2"}, {Title: "Video3"}},
			all:     false,
			input:   "1-3\n",
			want:    []int{0, 1, 2},
			wantErr: false,
			wantPrompt: "\nAvailable videos:\n1. Video1\n2. Video2\n3. Video3\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:    "select multiple videos with comma",
			videos:  []models.Video{{Title: "Video1"}, {Title: "Video2"}, {Title: "Video3"}},
			all:     false,
			input:   "1, 3\n",
			want:    []int{0, 2},
			wantErr: false,
			wantPrompt: "\nAvailable videos:\n1. Video1\n2. Video2\n3. Video3\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:    "select multiple videos with space",
			videos:  []models.Video{{Title: "Video1"}, {Title: "Video2"}, {Title: "Video3"}},
			all:     false,
			input:   "1 3\n",
			want:    []int{0, 2},
			wantErr: false,
			wantPrompt: "\nAvailable videos:\n1. Video1\n2. Video2\n3. Video3\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:    "invalid number",
			videos:  []models.Video{{Title: "Video1"}},
			all:     false,
			input:   "abc\n",
			want:    nil,
			wantErr: true,
			err:     errInvalidNumber,
			wantPrompt: "\nAvailable videos:\n1. Video1\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:    "number out of range",
			videos:  []models.Video{{Title: "Video1"}},
			all:     false,
			input:   "2\n",
			want:    nil,
			wantErr: true,
			err:     errNumberOutOfRange,
			wantPrompt: "\nAvailable videos:\n1. Video1\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:    "invalid range format",
			videos:  []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:     false,
			input:   "1-2-3\n",
			want:    nil,
			wantErr: true,
			err:     errInvalidRangeFormat,
			wantPrompt: "\nAvailable videos:\n1. Video1\n2. Video2\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:    "invalid start number in range",
			videos:  []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:     false,
			input:   "x-2\n",
			want:    nil,
			wantErr: true,
			err:     errInvalidStartNumber,
			wantPrompt: "\nAvailable videos:\n1. Video1\n2. Video2\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:    "invalid end number in range",
			videos:  []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:     false,
			input:   "1-y\n",
			want:    nil,
			wantErr: true,
			err:     errInvalidEndNumber,
			wantPrompt: "\nAvailable videos:\n1. Video1\n2. Video2\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:    "range out of bounds",
			videos:  []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:     false,
			input:   "1-3\n",
			want:    nil,
			wantErr: true,
			err:     errInvalidRange,
			wantPrompt: "\nAvailable videos:\n1. Video1\n2. Video2\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:    "start greater than end in range",
			videos:  []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:     false,
			input:   "2-1\n",
			want:    nil,
			wantErr: true,
			err:     errInvalidRange,
			wantPrompt: "\nAvailable videos:\n1. Video1\n2. Video2\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:    "no valid selections",
			videos:  []models.Video{{Title: "Video1"}},
			all:     false,
			input:   ",,,\n",
			want:    nil,
			wantErr: true,
			err:     errNoValidSelectionsFound,
			wantPrompt: "\nAvailable videos:\n1. Video1\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:       "empty video list",
			videos:     []models.Video{},
			all:        false,
			input:      "",
			want:       []int{},
			wantErr:    false,
			wantPrompt: "",
		},
		{
			name:    "duplicate selections",
			videos:  []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:     false,
			input:   "1,1,1-2,2\n",
			want:    []int{0, 1},
			wantErr: false,
			wantPrompt: "\nAvailable videos:\n1. Video1\n2. Video2\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:    "range with same start and end",
			videos:  []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:     false,
			input:   "1-1\n",
			want:    []int{0},
			wantErr: false,
			wantPrompt: "\nAvailable videos:\n1. Video1\n2. Video2\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:    "mixed valid and invalid selections",
			videos:  []models.Video{{Title: "Video1"}, {Title: "Video2"}, {Title: "Video3"}},
			all:     false,
			input:   "1,abc,3,5\n",
			want:    nil,
			wantErr: true,
			err:     errInvalidNumber,
			wantPrompt: "\nAvailable videos:\n1. Video1\n2. Video2\n3. Video3\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:    "negative numbers",
			videos:  []models.Video{{Title: "Video1"}},
			all:     false,
			input:   "-1\n",
			want:    nil,
			wantErr: true,
			err:     errInvalidStartNumber,
			wantPrompt: "\nAvailable videos:\n1. Video1\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:    "zero as input",
			videos:  []models.Video{{Title: "Video1"}},
			all:     false,
			input:   "0\n",
			want:    nil,
			wantErr: true,
			err:     errNumberOutOfRange,
			wantPrompt: "\nAvailable videos:\n1. Video1\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:    "input with extra whitespace",
			videos:  []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:     false,
			input:   "  1  ,  2  \n",
			want:    []int{0, 1},
			wantErr: false,
			wantPrompt: "\nAvailable videos:\n1. Video1\n2. Video2\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:    "range with spaces around dash",
			videos:  []models.Video{{Title: "Video1"}, {Title: "Video2"}, {Title: "Video3"}},
			all:     false,
			input:   "1 - 3\n",
			want:    nil,
			wantErr: true,
			err:     errInvalidStartNumber,
			wantPrompt: "\nAvailable videos:\n1. Video1\n2. Video2\n3. Video3\n\n" +
				"Select videos (e.g., '1-3', '1,3,5', '1 3 5', or Enter for all):\nSelection: ",
		},
		{
			name:    "single number with leading/trailing spaces",
			videos:  []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:     false,
			input:   "  2  \n",
			want:    []int{1},
			wantErr: false,
			wantPrompt: "\nAvailable videos:\n1. Video1\n2. Video2\n\n" +
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

				result, err = SelectVideos(tt.videos, tt.all)
				capturedOutput := readOutput()

				assert.Equal(t, tt.wantPrompt, capturedOutput)
			} else {
				result, err = SelectVideos(tt.videos, tt.all)
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
