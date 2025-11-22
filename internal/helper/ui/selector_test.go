package ui

import (
	"testing"

	"switchtube-downloader/internal/models"

	"github.com/stretchr/testify/assert"
)

const (
	table1Video = `┌────────┬────────┐
│ NUMBER │ TITLE  │
├────────┼────────┤
│      1 │ Video1 │
└────────┴────────┘

`
	table2Videos = `┌────────┬────────┐
│ NUMBER │ TITLE  │
├────────┼────────┤
│      1 │ Video1 │
│      2 │ Video2 │
└────────┴────────┘

`
	table3Videos = `┌────────┬────────┐
│ NUMBER │ TITLE  │
├────────┼────────┤
│      1 │ Video1 │
│      2 │ Video2 │
│      3 │ Video3 │
└────────┴────────┘

`
	episodeTable = `┌────────┬─────────┬────────┐
│ NUMBER │ EPISODE │ TITLE  │
├────────┼─────────┼────────┤
│      1 │      01 │ Video1 │
│      2 │      02 │ Video2 │
└────────┴─────────┴────────┘

`
	commonPrompt = `💡 Select videos:
   • Single: '1' or '3,5,7'
   • Range:  '1-5' or '1-3,7-9'
   • All:    Press Enter

🎯 Selection: `
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
			wantPrompt: table2Videos + commonPrompt,
		},
		{
			name:       "select single video",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:        false,
			useEpisode: false,
			input:      "1\n",
			want:       []int{0},
			wantErr:    false,
			wantPrompt: table2Videos + commonPrompt,
		},
		{
			name:       "select range",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}, {Title: "Video3"}},
			all:        false,
			useEpisode: false,
			input:      "1-3\n",
			want:       []int{0, 1, 2},
			wantErr:    false,
			wantPrompt: table3Videos + commonPrompt,
		},
		{
			name:       "select multiple videos with comma",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}, {Title: "Video3"}},
			all:        false,
			useEpisode: false,
			input:      "1, 3\n",
			want:       []int{0, 2},
			wantErr:    false,
			wantPrompt: table3Videos + commonPrompt,
		},
		{
			name:       "select multiple videos with space",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}, {Title: "Video3"}},
			all:        false,
			useEpisode: false,
			input:      "1 3\n",
			want:       []int{0, 2},
			wantErr:    false,
			wantPrompt: table3Videos + commonPrompt,
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
			wantPrompt: table1Video + commonPrompt,
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
			wantPrompt: table1Video + commonPrompt,
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
			wantPrompt: table2Videos + commonPrompt,
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
			wantPrompt: table2Videos + commonPrompt,
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
			wantPrompt: table2Videos + commonPrompt,
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
			wantPrompt: table2Videos + commonPrompt,
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
			wantPrompt: table2Videos + commonPrompt,
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
			wantPrompt: table1Video + commonPrompt,
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
			wantPrompt: table2Videos + commonPrompt,
		},
		{
			name:       "range with same start and end",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:        false,
			useEpisode: false,
			input:      "1-1\n",
			want:       []int{0},
			wantErr:    false,
			wantPrompt: table2Videos + commonPrompt,
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
			wantPrompt: table3Videos + commonPrompt,
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
			wantPrompt: table1Video + commonPrompt,
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
			wantPrompt: table1Video + commonPrompt,
		},
		{
			name:       "input with extra whitespace",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:        false,
			useEpisode: false,
			input:      "  1  ,  2  \n",
			want:       []int{0, 1},
			wantErr:    false,
			wantPrompt: table2Videos + commonPrompt,
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
			wantPrompt: table3Videos + commonPrompt,
		},
		{
			name:       "single number with leading/trailing spaces",
			videos:     []models.Video{{Title: "Video1"}, {Title: "Video2"}},
			all:        false,
			input:      "  2  \n",
			want:       []int{1},
			wantErr:    false,
			wantPrompt: table2Videos + commonPrompt,
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
			wantPrompt: episodeTable + commonPrompt,
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
