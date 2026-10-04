package database

import (
	"database/sql"
	"fmt"
	"strings"

	"GoData/tracks"

	_ "modernc.org/sqlite"
)

type DriverResult struct {
	Position         int
	RaceNumber       int
	Name             string
	NumLaps          int
	TotalRaceTime    string
	BestLapTime      string
	Points           int
	ResultStatusText string
	TeamName         string
}

type RaceResult struct {
	SessionUID  string
	SessionType string
	Drivers     []DriverResult
}

type FullSeasonResult struct {
	DriverNumber int
	DriverName   string
	TeamName     string
	Points       int
}

type RaceWeekend struct {
	TrackName  string             `json:"track_name"`
	TrackID    int                `json:"track_id"`
	Qualifying []FinalRaceResult  `json:"qualifying"`
	Sprint     []FinalRaceResult  `json:"sprint"`
	Race       []FinalRaceResult  `json:"race"`
}
type FinalRaceResult struct {
	SessionUID  string `json:"-"`
	CreatedAt   string `json:"-"`
	SessionType string `json:"-"`

	Position      int    `json:"position"`
	RaceNumber    int    `json:"race_number"`
	DriverName    string `json:"driver_name"`
	TeamName      string `json:"team_name"`
	NumLaps       int    `json:"num_laps"`
	TotalRaceTime string `json:"total_race_time"`
	BestLapTime   string `json:"best_lap_time"`
	Points        int    `json:"points"`
	Status        string `json:"status"`
}

type RaceIndex struct {
    TrackName string `json:"track_name"`
    File      string `json:"file"`
}

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("could not open SQLite database: %w", err)
	}

	pragmaSQL := `
	PRAGMA journal_mode = WAL;
	PRAGMA foreign_keys = ON;`
	if _, err := db.Exec(pragmaSQL); err != nil {
		return nil, fmt.Errorf("error when setting SQLite PRAGMA's: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS races (
		session_uid TEXT PRIMARY KEY,
		track_name TEXT NOT NULL,
		track_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		session_type TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS race_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_uid TEXT NOT NULL,
		position INTEGER NOT NULL,
		race_number INTEGER,
		driver_name TEXT NOT NULL,
		num_laps INTEGER,
		total_race_time TEXT,
		best_lap_time TEXT,
		points INTEGER,
		status TEXT,
		session_type TEXT NOT NULL,
		team_name TEXT,
		FOREIGN KEY (session_uid) REFERENCES races(session_uid) ON DELETE CASCADE
	);`

	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("error when creating database schema: %w", err)
	}

	return db, nil
}

func SaveRaceResult(db *sql.DB, result RaceResult, trackID int8) error {
	trackName := track.Name(trackID)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not start SQL transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT OR IGNORE INTO races (session_uid, track_name, track_id, session_type)
		VALUES (?, ?, ?, ?)`, result.SessionUID, trackName, trackID, result.SessionType)
	if err != nil {
		return fmt.Errorf("error when saving race data: %w", err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO race_results (
			session_uid, position, race_number, driver_name, num_laps,
			total_race_time, best_lap_time, points, status, session_type, team_name
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("error when preparing driver statement: %w", err)
	}
	defer stmt.Close()

	for _, driver := range result.Drivers {
		if _, err := stmt.Exec(
			result.SessionUID, driver.Position, driver.RaceNumber, driver.Name,
			driver.NumLaps, driver.TotalRaceTime, driver.BestLapTime, driver.Points,
			driver.ResultStatusText, result.SessionType, driver.TeamName,
		); err != nil {
			return fmt.Errorf("error when saving driver %s: %w", driver.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error when committing transaction: %w", err)
	}
	return nil
}

func GetFullSeasonResults(db *sql.DB) ([]FullSeasonResult, error) {
	rows, err := db.Query(`
		SELECT race_number, driver_name, team_name, SUM(points) AS total_points
		FROM race_results
		GROUP BY driver_name
		ORDER BY total_points DESC`)
	if err != nil {
		return nil, fmt.Errorf("error when querying database: %w", err)
	}
	defer rows.Close()

	var results []FullSeasonResult
	for rows.Next() {
		var result FullSeasonResult
		if err := rows.Scan(&result.DriverNumber, &result.DriverName, &result.TeamName, &result.Points); err != nil {
			return nil, fmt.Errorf("error when scanning row: %w", err)
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error when iterating over rows: %w", err)
	}
	return results, nil
}

func GetRaceWeekendResults(db *sql.DB, trackName string) (*RaceWeekend, error) {
	rows, err := db.Query(`
		SELECT
			r.session_uid,
			r.track_name,
			r.track_id,
			r.created_at,
			r.session_type,
			rr.position,
			rr.race_number,
			rr.driver_name,
			rr.team_name,
			rr.num_laps,
			rr.total_race_time,
			rr.best_lap_time,
			rr.points,
			rr.status
		FROM races r
		JOIN race_results rr
			ON rr.session_uid = r.session_uid
		WHERE r.track_name = ?
		ORDER BY
			CASE r.session_type
				WHEN 'Qualifying' THEN 1
				WHEN 'Sprint' THEN 2
				WHEN 'Race' THEN 3
				ELSE 4
			END,
			rr.position
	`, trackName)

	if err != nil {
		return nil, fmt.Errorf("error when querying database: %w", err)
	}
	defer rows.Close()

	var weekend *RaceWeekend

	for rows.Next() {
		var result FinalRaceResult
		var resultTrackName string
		var resultTrackID int

		if err := rows.Scan(
			&result.SessionUID,
			&resultTrackName,
			&resultTrackID,
			&result.CreatedAt,
			&result.SessionType,
			&result.Position,
			&result.RaceNumber,
			&result.DriverName,
			&result.TeamName,
			&result.NumLaps,
			&result.TotalRaceTime,
			&result.BestLapTime,
			&result.Points,
			&result.Status,
		); err != nil {
			return nil, fmt.Errorf("error when scanning row: %w", err)
		}

		if weekend == nil {
			weekend = &RaceWeekend{
				TrackName: resultTrackName,
				TrackID:   resultTrackID,
			}
		}

		switch result.SessionType {
		case "Qualifying":
			weekend.Qualifying = append(weekend.Qualifying, result)

		case "Sprint":
			weekend.Sprint = append(weekend.Sprint, result)

		case "Race":
			weekend.Race = append(weekend.Race, result)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error when iterating over rows: %w", err)
	}

	return weekend, nil
}

func GetRaceIndex(db *sql.DB) ([]RaceIndex, error) {
    rows, err := db.Query(`
        SELECT DISTINCT track_name
        FROM races
        ORDER BY created_at
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var races []RaceIndex

    for rows.Next() {
        var race RaceIndex

        if err := rows.Scan(&race.TrackName); err != nil {
            return nil, err
        }

        race.File = strings.ToLower(race.TrackName) + ".json"
        races = append(races, race)
    }

    return races, rows.Err()
}