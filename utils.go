package main

import (
	"bytes"
	"fmt"
	"math"
	"strings"
)

func printResult(r FinalRaceResult) {
	fmt.Printf("\n\n=== FINAL CLASSIFICATION (%d cars) ===\n", r.NumCars)
	fmt.Printf("%-4s %-4s %-22s %5s %13s %12s %4s  %s\n",
		"Pos", "Nr", "Driver", "Laps", "Race Time", "Best Lap", "Pts", "Status")

	for _, d := range r.Drivers {
		fmt.Printf("%-4d %-4d %-22s %5d %13s %12s %4d  %s\n",
			d.Position, d.RaceNumber, truncate(d.Name, 22), d.NumLaps,
			d.TotalRaceTime, d.BestLapTime, d.Points, d.ResultStatusText)
	}
	fmt.Println()
}


func cString(b []byte) string {
	if i := bytes.IndexByte(b, 0); i >= 0 {
		b = b[:i]
	}
	return strings.TrimSpace(string(b))
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "."
}

func formatLapTime(ms uint32) string {
	if ms == 0 {
		return "-"
	}
	return fmt.Sprintf("%d:%02d.%03d", ms/60000, (ms%60000)/1000, ms%1000)
}

func formatRaceTime(sec float64) string {
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

func resultStatusName(status uint8) string {
	switch status {
	case 0:
		return "invalid"
	case 1:
		return "inactive"
	case 2:
		return "active"
	case 3:
		return "finished"
	case 4:
		return "didnotfinish"
	case 5:
		return "disqualified"
	case 6:
		return "not classified"
	case 7:
		return "retired"
	}
	return fmt.Sprintf("unknown (%d)", status)
}

func compoundName(visual uint8) string {
	switch visual {
	case 7:
		return "Intermediate"
	case 8, 15:
		return "Wet"
	case 16, 20:
		return "Soft"
	case 17, 21:
		return "Medium"
	case 18, 22:
		return "Hard"
	case 19:
		return "Super Soft"
	}
	return fmt.Sprintf("unknown (%d)", visual)
}