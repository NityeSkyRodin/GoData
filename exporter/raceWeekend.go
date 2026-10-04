package exporter

import (
	"database/sql"
	"fmt"

	"GoData/database"
)


func ParseRaceWeekendResults(db *sql.DB, trackName string) error {

	weekendResults, err := database.GetRaceWeekendResults(db, trackName)
	if err != nil {
		return err
	}

	fmt.Printf("Race Weekend Results for %s:\n%v\n", trackName, weekendResults)

	err = ParseToJSONFile(weekendResults, fmt.Sprintf("data/races/%s.json", trackName))
	if err != nil {
		return err
	}

	err = ExportRaceIndex(db, "data/races/index.json")
	if err != nil {
		return err
	}

	return nil
}

func ExportRaceIndex(db *sql.DB, filename string) error {
    races, err := database.GetRaceIndex(db)
    if err != nil {
        return err
    }

    return ParseToJSONFile(races, filename)
}