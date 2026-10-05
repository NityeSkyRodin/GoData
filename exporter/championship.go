package exporter

import (
	"database/sql"
	"fmt"

	"GoData/database"
)

func ParseFullSeasonResult(db *sql.DB) error {
	seasonResults, err := database.GetFullSeasonResults(db)
	if err != nil {
		return err
	}

	fmt.Printf("Full Season Results:\n%v\n", seasonResults)

	err = ParseToJSONFile(seasonResults, "data/championship.json")
	if err != nil {
		return err
	}

	return nil
}