package telemetry

import "fmt"

// Dit bestand vertaalt de id's die de game in de pakketten stuurt naar leesbare namen.

// mapTeam hoort bij het F2 2025-veld: elk team heeft twee vaste startnummers.
func mapTeam(teamID uint8) string {
	switch teamID {
	case 209:
		return "ART Grand Prix"
	case 210:
		return "Campos Racing"
	case 211:
		return "Rodin Motorsport"
	case 212:
		return "AIX Racing"
	case 213:
		return "DAMS Lucas Oil"
	case 214:
		return "Hitech Pulse-Eight"
	case 215:
		return "MP Motorsport"
	case 216:
		return "PREMA Racing"
	case 217:
		return "Trident"
	case 218:
		return "Van Amersfoort Racing"
	case 219:
		return "Invicta Racing"
	default:
		return "Unknown"
	}
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

// compoundName vertaalt de visual tyre compound. De F2-banden hebben eigen id's.
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
