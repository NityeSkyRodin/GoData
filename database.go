package main

import (
	track "GoData/tracks"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)


func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("could not open SQLite database: %w", err)
	}

	pragmaSQL := `
	PRAGMA journal_mode = WAL;
	PRAGMA foreign_keys = ON;
	`
	if _, err := db.Exec(pragmaSQL); err != nil {
		return nil, fmt.Errorf("error when setting SQLite PRAGMA's: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS races (
		session_uid TEXT PRIMARY KEY,
		track_name  TEXT NOT NULL,
		track_id    INTEGER NOT NULL,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS race_results (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		session_uid     TEXT NOT NULL,
		position        INTEGER NOT NULL,
		race_number     INTEGER,
		driver_name     TEXT NOT NULL,
		num_laps        INTEGER,
		total_race_time TEXT,
		best_lap_time   TEXT,
		points          INTEGER,
		status          TEXT,
		FOREIGN KEY (session_uid) REFERENCES races(session_uid) ON DELETE CASCADE
	);`

	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("error when creating database schema: %w", err)
	}

	return db, nil
}

func SaveRaceResult(db *sql.DB, result FinalRaceResult, trackID int8) error {
	trackName := track.Name(trackID)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not start SQL transaction: %w", err)
	}
	defer tx.Rollback()

	raceSQL := `
	INSERT OR IGNORE INTO races (session_uid, track_name, track_id)
	VALUES (?, ?, ?)`

	_, err = tx.Exec(raceSQL, result.SessionUID, trackName, trackID)
	if err != nil {
		return fmt.Errorf("error when saving race data: %w", err)
	}

	driverSQL := `
	INSERT INTO race_results (
		session_uid, 
		position, 
		race_number, 
		driver_name, 
		num_laps, 
		total_race_time, 
		best_lap_time, 
		points, 
		status
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := tx.Prepare(driverSQL)
	if err != nil {
		return fmt.Errorf("error when preparing driver statement: %w", err)
	}
	defer stmt.Close()

	for _, d := range result.Drivers {
		_, err := stmt.Exec(
			result.SessionUID,
			d.Position,
			d.RaceNumber,
			d.Name,
			d.NumLaps,
			d.TotalRaceTime,
			d.BestLapTime,
			d.Points,
			d.ResultStatusText,
		)
		if err != nil {
			return fmt.Errorf("error when saving driver %s: %w", d.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error when committing transaction: %w", err)
	}

	return nil
}