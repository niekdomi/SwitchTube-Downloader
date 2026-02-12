package progress

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"switchtube-downloader/internal/helper/ui/ansi"
)

const (
	refreshRateMs = 100
	minUpdateGap  = 50 * time.Millisecond
)

//nolint:gochecknoglobals
var displayMutex sync.Mutex // Prevents concurrent display updates

var errFailedToCopyData = errors.New("failed to copy data")

// progressWriter wraps an io.Writer and tracks progress.
type progressWriter struct {
	startTime       time.Time // Start time for speed calculation
	lastUpdate      time.Time // Last progress update time
	writer          io.Writer // Underlying destination writer
	filename        string    // File being downloaded
	total           int64     // Expected total bytes
	written         int64     // Bytes written so far
	rowIndex        int       // Row index for multi-line progress display
	longestFilename int       // Longest filename for alignment
}

// Write implements io.Writer and updates progress.
func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	pw.written += int64(n)

	now := time.Now()
	if now.Sub(pw.lastUpdate) >= minUpdateGap {
		pw.lastUpdate = now
		pw.displayProgress()
	}

	if err != nil {
		return n, fmt.Errorf("%w: %w", errFailedToCopyData, err)
	}

	return n, nil
}

// displayProgress renders the progress bar to stdout.
func (pw *progressWriter) displayProgress() {
	const divByZeroGuard = 0.001

	elapsed := max(time.Since(pw.startTime).Seconds(), divByZeroGuard)

	percentage := 0.0
	if pw.total > 0 {
		percentage = (float64(pw.written) / float64(pw.total)) * 100
	}

	speed := (float64(pw.written) / elapsed)

	displayMutex.Lock()
	defer displayMutex.Unlock()

	basename := filepath.Base(pw.filename)

	// Add padding for alignment if needed
	if (pw.longestFilename > 0) && (len(basename) < pw.longestFilename) {
		basename += strings.Repeat(" ", pw.longestFilename-len(basename))
	}

	if pw.rowIndex > 0 {
		fmt.Print(ansi.SaveCursor)
		fmt.Printf(ansi.MoveCursorUp, pw.rowIndex)
		fmt.Printf("%s%s%s %s", ansi.CarriageReturn, ansi.ClearLineRight, basename, renderProgressBar(percentage, speed))
		fmt.Print(ansi.RestoreCursor)
	} else {
		fmt.Printf("%s%s %s", ansi.CarriageReturn, basename, renderProgressBar(percentage, speed))
	}
}

// BarWithRow sets up a progress bar.
func BarWithRow(src io.Reader, dst io.Writer, total int64, filename string, rowIndex int, longestFilename int) error {
	pw := &progressWriter{
		writer:          dst,
		total:           total,
		written:         0,
		startTime:       time.Now(),
		lastUpdate:      time.Now(),
		filename:        filename,
		rowIndex:        rowIndex,
		longestFilename: longestFilename,
	}

	if _, err := io.Copy(pw, src); err != nil {
		return fmt.Errorf("%w: %w", errFailedToCopyData, err)
	}

	// Final update
	pw.displayProgress()

	if rowIndex == 0 {
		fmt.Println()
	}

	return nil
}
