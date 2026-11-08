package database

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

const (
	statusFinished     = "finished"
	statusDisqualified = "disqualified"

	carryOverSeconds = 5
)

var pointsPerPosition = map[string][]int{
	"Qualifying": {2},
	"Sprint":     {10, 8, 6, 5, 4, 3, 2, 1},
	"Race":       {25, 18, 15, 12, 10, 8, 6, 4, 2, 1},
}

type PenaltyInput struct {
	TrackName    string
	SessionType  string
	DriverName   string
	Seconds      int
	Disqualified bool
}

type sessionRow struct {
	ID            int
	Position      int
	RaceNumber    int
	DriverName    string
	NumLaps       int
	TotalRaceTime string
	Points        int
	Status        string
}

type pendingPenalty struct {
	ID         int
	RaceNumber int
	DriverName string
	Seconds    int
}

func ApplyPenalty(db *sql.DB, p PenaltyInput) (carriedOver bool, err error) {
	tx, err := db.Begin()
	if err != nil {
		return false, fmt.Errorf("could not start SQL transaction: %w", err)
	}
	defer tx.Rollback()

	var sessionUID string
	err = tx.QueryRow(`
		SELECT session_uid
		FROM races
		WHERE track_name = ? AND session_type = ?
		ORDER BY created_at DESC
		LIMIT 1`, p.TrackName, p.SessionType).Scan(&sessionUID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("no %s session found for %s", p.SessionType, p.TrackName)
	}
	if err != nil {
		return false, fmt.Errorf("error when looking up session: %w", err)
	}

	rows, err := getSessionRows(tx, sessionUID)
	if err != nil {
		return false, err
	}

	i := -1
	for j, row := range rows {
		if row.DriverName == p.DriverName {
			i = j
			break
		}
	}
	if i == -1 {
		return false, fmt.Errorf("driver %s did not take part in %s %s", p.DriverName, p.TrackName, p.SessionType)
	}

	driver := rows[i]

	switch {
	case p.Disqualified:
		driver.Status = statusDisqualified
		rows = append(append(rows[:i:i], rows[i+1:]...), driver)
		carriedOver = true
	case rows[i].Status != statusFinished:
		carriedOver = true
	default:
		if err := addTime(&rows[i], p.Seconds); err != nil {
			return false, err
		}
		moveDown(rows, i)
	}

	if carriedOver {
		if _, err := tx.Exec(`
			INSERT INTO pending_penalties (race_number, driver_name, seconds, origin_session_uid)
			VALUES (?, ?, ?, ?)`,
			driver.RaceNumber, driver.DriverName, carryOverSeconds, sessionUID); err != nil {
			return false, fmt.Errorf("error when saving carried over penalty: %w", err)
		}
	}

	if err := rerankAndSave(tx, rows, p.SessionType); err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("error when committing transaction: %w", err)
	}
	return carriedOver, nil
}

func ApplyPendingPenalties(db *sql.DB, sessionUID string) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("could not start SQL transaction: %w", err)
	}
	defer tx.Rollback()

	var sessionType string
	if err := tx.QueryRow(`SELECT session_type FROM races WHERE session_uid = ?`,
		sessionUID).Scan(&sessionType); err != nil {
		return 0, fmt.Errorf("error when looking up session %s: %w", sessionUID, err)
	}
	if sessionType != "Sprint" && sessionType != "Race" {
		return 0, nil
	}

	pending, err := getOpenPendingPenalties(tx)
	if err != nil {
		return 0, err
	}
	if len(pending) == 0 {
		return 0, nil
	}

	rows, err := getSessionRows(tx, sessionUID)
	if err != nil {
		return 0, err
	}

	applied := 0
	for _, pp := range pending {
		i := -1
		for j, row := range rows {
			if row.RaceNumber == pp.RaceNumber {
				i = j
				break
			}
		}
		if i == -1 || rows[i].Status != statusFinished {
			continue
		}

		if err := addTime(&rows[i], pp.Seconds); err != nil {
			return 0, err
		}
		moveDown(rows, i)

		if _, err := tx.Exec(`UPDATE pending_penalties SET applied_session_uid = ? WHERE id = ?`,
			sessionUID, pp.ID); err != nil {
			return 0, fmt.Errorf("error when closing pending penalty: %w", err)
		}
		applied++
	}

	if applied > 0 {
		if err := rerankAndSave(tx, rows, sessionType); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("error when committing transaction: %w", err)
	}
	return applied, nil
}

func CarryOverGamePenalties(db *sql.DB, result RaceResult) (int, error) {
	if result.SessionType != "Sprint" && result.SessionType != "Race" {
		return 0, nil
	}

	added := 0
	for _, driver := range result.Drivers {
		if driver.GamePenaltySeconds == 0 || driver.ResultStatusText == statusFinished {
			continue
		}

		res, err := db.Exec(`
			INSERT INTO pending_penalties (race_number, driver_name, seconds, origin_session_uid)
			SELECT ?, ?, ?, ?
			WHERE NOT EXISTS (
				SELECT 1 FROM pending_penalties
				WHERE race_number = ? AND origin_session_uid = ?
			)`,
			driver.RaceNumber, driver.Name, carryOverSeconds, result.SessionUID,
			driver.RaceNumber, result.SessionUID)
		if err != nil {
			return added, fmt.Errorf("error when saving carried over game penalty for %s: %w", driver.Name, err)
		}
		if n, _ := res.RowsAffected(); n > 0 {
			added++
		}
	}
	return added, nil
}

func getSessionRows(tx *sql.Tx, sessionUID string) ([]sessionRow, error) {
	rows, err := tx.Query(`
		SELECT id, position, race_number, driver_name, num_laps, total_race_time, points, status
		FROM race_results
		WHERE session_uid = ?
		ORDER BY position`, sessionUID)
	if err != nil {
		return nil, fmt.Errorf("error when querying session results: %w", err)
	}
	defer rows.Close()

	var result []sessionRow
	for rows.Next() {
		var row sessionRow
		if err := rows.Scan(&row.ID, &row.Position, &row.RaceNumber, &row.DriverName,
			&row.NumLaps, &row.TotalRaceTime, &row.Points, &row.Status); err != nil {
			return nil, fmt.Errorf("error when scanning session result: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func getOpenPendingPenalties(tx *sql.Tx) ([]pendingPenalty, error) {
	rows, err := tx.Query(`
		SELECT id, race_number, driver_name, seconds
		FROM pending_penalties
		WHERE applied_session_uid IS NULL
		ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("error when querying pending penalties: %w", err)
	}
	defer rows.Close()

	var result []pendingPenalty
	for rows.Next() {
		var pp pendingPenalty
		if err := rows.Scan(&pp.ID, &pp.RaceNumber, &pp.DriverName, &pp.Seconds); err != nil {
			return nil, fmt.Errorf("error when scanning pending penalty: %w", err)
		}
		result = append(result, pp)
	}
	return result, rows.Err()
}

func moveDown(rows []sessionRow, i int) {
	newTime := raceTimeMS(rows[i].TotalRaceTime)
	for i+1 < len(rows) {
		next := rows[i+1]
		if next.Status != statusFinished || next.NumLaps != rows[i].NumLaps ||
			raceTimeMS(next.TotalRaceTime) >= newTime {
			break
		}
		rows[i], rows[i+1] = rows[i+1], rows[i]
		i++
	}
}

func rerankAndSave(tx *sql.Tx, rows []sessionRow, sessionType string) error {
	bonus := make(map[int]int, len(rows))
	for _, row := range rows {
		if b := row.Points - basePoints(sessionType, row.Position); b > 0 {
			bonus[row.ID] = b
		}
	}

	stmt, err := tx.Prepare(`
		UPDATE race_results
		SET position = ?, total_race_time = ?, points = ?, status = ?
		WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("error when preparing update statement: %w", err)
	}
	defer stmt.Close()

	for i := range rows {
		rows[i].Position = i + 1
		rows[i].Points = 0
		if rows[i].Status == statusFinished {
			rows[i].Points = basePoints(sessionType, rows[i].Position)
			if rows[i].Position <= 10 {
				rows[i].Points += bonus[rows[i].ID]
			}
		}

		if _, err := stmt.Exec(rows[i].Position, rows[i].TotalRaceTime, rows[i].Points,
			rows[i].Status, rows[i].ID); err != nil {
			return fmt.Errorf("error when updating %s: %w", rows[i].DriverName, err)
		}
	}
	return nil
}

func basePoints(sessionType string, position int) int {
	table := pointsPerPosition[sessionType]
	if position < 1 || position > len(table) {
		return 0
	}
	return table[position-1]
}

func addTime(row *sessionRow, seconds int) error {
	ms, ok := parseRaceTime(row.TotalRaceTime)
	if !ok {
		return fmt.Errorf("cannot add penalty to %s: invalid race time %q", row.DriverName, row.TotalRaceTime)
	}
	row.TotalRaceTime = formatRaceTime(ms + int64(seconds)*1000)
	return nil
}

func raceTimeMS(s string) int64 {
	if ms, ok := parseRaceTime(s); ok {
		return ms
	}
	return math.MaxInt64
}

func parseRaceTime(s string) (int64, bool) {
	parts := strings.Split(s, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, false
	}

	sec, err := strconv.ParseFloat(parts[len(parts)-1], 64)
	if err != nil {
		return 0, false
	}
	total := int64(math.Round(sec * 1000))

	multiplier := int64(60_000)
	for i := len(parts) - 2; i >= 0; i-- {
		n, err := strconv.ParseInt(parts[i], 10, 64)
		if err != nil {
			return 0, false
		}
		total += n * multiplier
		multiplier *= 60
	}
	return total, true
}

func formatRaceTime(total int64) string {
	ms := total % 1000
	whole := total / 1000
	if h := whole / 3600; h > 0 {
		return fmt.Sprintf("%d:%02d:%02d.%03d", h, (whole%3600)/60, whole%60, ms)
	}
	return fmt.Sprintf("%d:%02d.%03d", whole/60, whole%60, ms)
}
