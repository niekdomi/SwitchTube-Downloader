package ui

import (
	"fmt"
	"strings"
)

// Progress bar symbols.
const (
	ProgressFilled   = "━"
	ProgressEmpty    = "─"
	ProgressBarWidth = 30
	percentageBase   = 100.0
)

// formatSpeed formats download speed in human-readable format.
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

// renderProgressBar renders a progress bar with percentage and speed.
func renderProgressBar(percentage float64, bytePerSec float64) string {
	filled := int((percentage / percentageBase) * float64(ProgressBarWidth))

	var bar strings.Builder

	for i := range ProgressBarWidth {
		if i < filled {
			bar.WriteString(Green + ProgressFilled)
		} else {
			bar.WriteString(Dim + ProgressEmpty)
		}
	}

	bar.WriteString(Reset)

	displaySpeed, unit := formatSpeed(bytePerSec)
	fmt.Fprintf(&bar, " %5.1f%% %s%6.2f %s%s", percentage, Dim, displaySpeed, unit, Reset)

	return bar.String()
}
