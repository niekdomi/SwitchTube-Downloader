// Package progress provides utilities for rendering progress bars in the terminal.
package progress

import (
	"fmt"
	"strings"

	"switchtube-downloader/internal/helper/ui/ansi"
)

// Progress bar symbols.
const (
	ProgressFilled   = "━"
	ProgressEmpty    = "─"
	ProgressBarWidth = 30
)

// formatSpeed converts bytes per second to appropriate units (Gb/s, Mb/s, Kb/s, b/s).
// Returns speed value and unit as string.
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

// renderProgressBar renders a progress bar with percentage and speed display.
// Returns formatted string with progress bar, percentage, and download speed.
func renderProgressBar(percentage float64, bytePerSec float64) string {
	filled := 0
	if percentage > 0 {
		filled = int((percentage / 100) * ProgressBarWidth)
	}

	var bar strings.Builder

	for i := range ProgressBarWidth {
		if i < filled {
			bar.WriteString(ansi.Green + ProgressFilled)
		} else {
			bar.WriteString(ansi.Dim + ProgressEmpty)
		}
	}

	bar.WriteString(ansi.Reset)

	displaySpeed, unit := formatSpeed(bytePerSec)
	fmt.Fprintf(&bar, " %5.1f%% %s%6.2f %s%s", percentage, ansi.Dim, displaySpeed, unit, ansi.Reset)

	return bar.String()
}
