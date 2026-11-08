package database

import (
	"database/sql"
	"encoding/json"
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
	TyreStints       []TyreStint
}

// TyreStint is één stuk race op dezelfde band: welke band en tot welke ronde.
type TyreStint struct {
	Compound string `json:"compound"`
	EndLap   int    `json:"end_lap"`
}

type RaceResult struct {
	SessionUID  string
	SessionType string
	Drivers     []DriverResult
}

type FullSeasonResult struct {
    DriverNumber int    `json:"driver_number"`
    DriverName   string `json:"driver_name"`
    TeamName     string `json:"team_name"`
    Points       int    `json:"points"`
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

	// Leeg bij oude races en kwalificatie, dan laat omitempty het veld weg uit de JSON
	TyreStints []TyreStint `json:"tyre_stints,omitempty"`
}

type RaceIndex struct {
    TrackName string `json:"track_name"`
    File      string `json:"file"`
}

// Race is één opgeslagen sessie uit de races-tabel.
type Race struct {
	SessionUID  string
	TrackName   string
	TrackID     int
	SessionType string
	CreatedAt   string
}

// Driver is een coureur zoals die in de resultaten voorkomt. Er is geen
// aparte coureurs-tabel; deze wordt afgeleid uit race_results.
type Driver struct {
	RaceNumber int
	DriverName string
	TeamName   string
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
	);
	CREATE TABLE IF NOT EXISTS pending_penalties (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		race_number INTEGER NOT NULL,
		driver_name TEXT NOT NULL,
		seconds INTEGER NOT NULL,
		origin_session_uid TEXT NOT NULL,
		applied_session_uid TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("error when creating database schema: %w", err)
	}

	// Kolommen die later zijn toegevoegd. CREATE TABLE IF NOT EXISTS voegt die
	// niet toe aan een bestaande tabel, dus dat doen we hier zelf.
	if err := addColumnIfMissing(db, "race_results", "tyre_stints", "TEXT"); err != nil {
		return nil, err
	}

	return db, nil
}

func addColumnIfMissing(db *sql.DB, table, column, columnType string) error {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return fmt.Errorf("error when reading columns of %s: %w", table, err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid, notNull, pk int
		var name, colType string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultValue, &pk); err != nil {
			return fmt.Errorf("error when scanning column of %s: %w", table, err)
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if _, err := db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, columnType)); err != nil {
		return fmt.Errorf("error when adding column %s to %s: %w", column, table, err)
	}
	return nil
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
			total_race_time, best_lap_time, points, status, session_type, team_name, tyre_stints
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("error when preparing driver statement: %w", err)
	}
	defer stmt.Close()

	for _, driver := range result.Drivers {
		stints, err := json.Marshal(driver.TyreStints)
		if err != nil {
			return fmt.Errorf("error when encoding tyre stints of %s: %w", driver.Name, err)
		}
		if _, err := stmt.Exec(
			result.SessionUID, driver.Position, driver.RaceNumber, driver.Name,
			driver.NumLaps, driver.TotalRaceTime, driver.BestLapTime, driver.Points,
			driver.ResultStatusText, result.SessionType, driver.TeamName, string(stints),
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
			rr.status,
			COALESCE(rr.tyre_stints, '')
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
		var stints string // JSON-tekst, leeg bij races van voor deze kolom

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
			&stints,
		); err != nil {
			return nil, fmt.Errorf("error when scanning row: %w", err)
		}

		if stints != "" {
			if err := json.Unmarshal([]byte(stints), &result.TyreStints); err != nil {
				return nil, fmt.Errorf("error when decoding tyre stints of %s: %w", result.DriverName, err)
			}
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
// GetRaces geeft alle opgeslagen sessies terug, oudste eerst.
func GetRaces(db *sql.DB) ([]Race, error) {
	rows, err := db.Query(`
		SELECT session_uid, track_name, track_id, session_type, created_at
		FROM races
		ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("error when querying races: %w", err)
	}
	defer rows.Close()

	var races []Race
	for rows.Next() {
		var race Race
		if err := rows.Scan(&race.SessionUID, &race.TrackName, &race.TrackID,
			&race.SessionType, &race.CreatedAt); err != nil {
			return nil, fmt.Errorf("error when scanning race: %w", err)
		}
		races = append(races, race)
	}

	return races, rows.Err()
}

// GetDrivers geeft per startnummer één coureur terug, met de naam en het team
// uit de meest recente sessie. Namen kunnen per sessie wisselen, dus groeperen
// op race_number voorkomt dat dezelfde auto twee keer in de lijst komt.
func GetDrivers(db *sql.DB) ([]Driver, error) {
	rows, err := db.Query(`
		SELECT rr.race_number, rr.driver_name, rr.team_name, MAX(r.created_at)
		FROM race_results rr
		JOIN races r ON r.session_uid = rr.session_uid
		GROUP BY rr.race_number
		ORDER BY rr.race_number`)
	if err != nil {
		return nil, fmt.Errorf("error when querying drivers: %w", err)
	}
	defer rows.Close()

	var drivers []Driver
	for rows.Next() {
		var driver Driver
		var laatsteSessie string // alleen om MAX() ergens in te scannen
		if err := rows.Scan(&driver.RaceNumber, &driver.DriverName,
			&driver.TeamName, &laatsteSessie); err != nil {
			return nil, fmt.Errorf("error when scanning driver: %w", err)
		}
		drivers = append(drivers, driver)
	}

	return drivers, rows.Err()
}

// GetTrackNames geeft alleen de circuitnamen terug, in de volgorde waarin ze
// voor het eerst gereden zijn. Voor keuzelijstjes waar je de rest niet nodig hebt.
func GetTrackNames(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`
		SELECT track_name
		FROM races
		GROUP BY track_name
		ORDER BY MIN(created_at)`)
	if err != nil {
		return nil, fmt.Errorf("error when querying track names: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("error when scanning track name: %w", err)
		}
		names = append(names, name)
	}

	return names, rows.Err()
}

// GetDriverNames geeft alleen de coureursnamen terug, één per startnummer.
func GetDriverNames(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`
		SELECT rr.driver_name, MAX(r.created_at)
		FROM race_results rr
		JOIN races r ON r.session_uid = rr.session_uid
		GROUP BY rr.race_number
		ORDER BY rr.race_number`)
	if err != nil {
		return nil, fmt.Errorf("error when querying driver names: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name, laatsteSessie string
		if err := rows.Scan(&name, &laatsteSessie); err != nil {
			return nil, fmt.Errorf("error when scanning driver name: %w", err)
		}
		names = append(names, name)
	}

	return names, rows.Err()
}
