package telemetry

import (
	"fmt"
	"math"
)

// FormatLapTime zet een rondetijd in milliseconden om naar m:ss.mmm.
func FormatLapTime(ms uint32) string {
	if ms == 0 {
		return "-"
	}
	return fmt.Sprintf("%d:%02d.%03d", ms/60000, (ms%60000)/1000, ms%1000)
}

// FormatRaceTime zet een racetijd in seconden om naar [u:]mm:ss.mmm.
func FormatRaceTime(sec float64) string {
	if sec <= 0 {
		return "-"
	}
	total := int64(math.Round(sec * 1000))
	ms := total % 1000
	whole := total / 1000
	if h := whole / 3600; h > 0 {
		return fmt.Sprintf("%d:%02d:%02d.%03d", h, (whole%3600)/60, whole%60, ms)
	}
	return fmt.Sprintf("%d:%02d.%03d", whole/60, whole%60, ms)
}
