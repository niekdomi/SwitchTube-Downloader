package progress

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	xterm "github.com/charmbracelet/x/term"
)

const (
	// minBarWidth is the minimum progress bar width in characters.
	minBarWidth = 10
	// statsWidth is the fixed width of the stats suffix (e.g. " 100.0%  99.99 Gb/s").
	statsWidth = 22
)

var (
	styleDim = lipgloss.NewStyle().Faint(true)
	pb       = progress.New(
		progress.WithDefaultGradient(),
		progress.WithFillCharacters('━', '─'),
		progress.WithoutPercentage(),
	)
)

// barWidth calculates how wide the progress bar should be given the current
// terminal width and the filename column width.
func barWidth(filenameWidth int) int {
	const minPrefixGap = 1

	w, _, err := xterm.GetSize(os.Stdout.Fd())
	if err != nil || w <= 0 {
		w = 80
	}

	available := w - filenameWidth - minPrefixGap - statsWidth
	if available < minBarWidth {
		return minBarWidth
	}

	return available
}

// formatSpeed converts bytes per second to appropriate units (Gb/s, Mb/s, Kb/s, b/s).
func formatSpeed(bytePerSec float64) (float64, string) {
	const (
		Kbps = 125.0
		Mbps = Kbps * 1_000.0
		Gbps = Mbps * 1_000.0
	)

	switch {
	case bytePerSec >= Gbps:
		return bytePerSec / Gbps, "Gb/s"
	case bytePerSec >= Mbps:
		return bytePerSec / Mbps, "Mb/s"
	case bytePerSec >= Kbps:
		return bytePerSec / Kbps, "Kb/s"
	default:
		return bytePerSec, "b/s"
	}
}

// renderProgressBar renders a progress bar sized to the terminal width.
func renderProgressBar(percentage float64, bytePerSec float64, filenameWidth int) string {
	bw := barWidth(filenameWidth)

	pb.Width = bw
	renderedBar := pb.ViewAs(percentage / 100.0)

	displaySpeed, unit := formatSpeed(bytePerSec)
	return fmt.Sprintf("%s %5.1f%% %s", renderedBar, percentage, styleDim.Render(fmt.Sprintf("%6.2f %s", displaySpeed, unit)))
}
