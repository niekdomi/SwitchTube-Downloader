package ui

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"
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
	writer           io.Writer
	total            int64
	written          int64
	startTime        time.Time
	lastUpdate       time.Time
	filename         string
	rowIndex         int // Row index for multi-line progress display (0 = single line mode)
	maxFilenameWidth int // Maximum filename width for alignment (0 = no padding)
}

// Write implements io.Writer and updates progress.
func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	pw.written += int64(n)

	// Update progress display
	now := time.Now()
	if now.Sub(pw.lastUpdate) >= minUpdateGap {
		pw.lastUpdate = now
		pw.displayProgress()
	}

	return n, fmt.Errorf("%w: %w", errFailedToCopyData, err)
}

// displayProgress renders the progress bar to stdout.
func (pw *progressWriter) displayProgress() {
	elapsed := time.Since(pw.startTime).Seconds()
	if elapsed == 0 {
		elapsed = 0.001 // Prevent division by zero
	}

	percentage := 0.0
	if pw.total > 0 {
		percentage = (float64(pw.written) / float64(pw.total)) * percentageBase
	}

	speed := (float64(pw.written) / elapsed)

	displayMutex.Lock()
	defer displayMutex.Unlock()

	basename := filepath.Base(pw.filename)
	// Pad filename if maxFilenameWidth is set
	if pw.maxFilenameWidth > 0 && len(basename) < pw.maxFilenameWidth {
		basename += strings.Repeat(" ", pw.maxFilenameWidth-len(basename))
	}

	if pw.rowIndex > 0 {
		// Multi-line mode: save cursor, move up to target row, update, restore cursor
		fmt.Print("\033[s")
		fmt.Printf("\033[%dA", pw.rowIndex)
		fmt.Printf("\r\033[K%s %s", basename, renderProgressBar(percentage, speed))
		fmt.Print("\033[u")
	} else {
		// Single-line mode: use carriage return
		fmt.Printf("\r%s %s", basename, renderProgressBar(percentage, speed))
	}
}

// ProgressBar sets up a progress bar for downloading and copies data from src to dst.
func ProgressBar(src io.Reader, dst io.Writer, total int64, filename string) error {
	return ProgressBarWithRow(src, dst, total, filename, 0, 0)
}

// ProgressBarWithRow sets up a progress bar with a specific row index for multi-line display.
// rowIndex 0 means single-line mode (uses \r), rowIndex > 0 uses cursor positioning.
// maxFilenameWidth is used for padding filenames to align progress bars (0 = no padding).
func ProgressBarWithRow(src io.Reader, dst io.Writer, total int64, filename string, rowIndex int, maxFilenameWidth int) error {
	pw := &progressWriter{
		writer:           dst,
		total:            total,
		written:          0,
		startTime:        time.Now(),
		lastUpdate:       time.Now(),
		filename:         filename,
		rowIndex:         rowIndex,
		maxFilenameWidth: maxFilenameWidth,
	}

	// Copy data with progress tracking
	if _, err := io.Copy(pw, src); err != nil {
		return fmt.Errorf("%w: %w", errFailedToCopyData, err)
	}

	// Final update
	pw.displayProgress()

	return nil
}
