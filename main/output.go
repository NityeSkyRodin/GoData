package main

import (
	"fmt"

	"GoData/telemetry"
)

func printResult(r telemetry.FinalRaceResult) {
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

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "."
}
