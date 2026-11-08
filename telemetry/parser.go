package telemetry

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"GoData/database"
	"GoData/tracks"
)

// cString leest een naam uit een vast byte-veld: alles tot de eerste nul-byte.
func cString(b []byte) string {
	if i := bytes.IndexByte(b, 0); i >= 0 {
		b = b[:i]
	}
	return strings.TrimSpace(string(b))
}

func ParseParticipants(payload []byte) map[int]Participant {
	out := map[int]Participant{}
	if len(payload) < 1 {
		return out
	}

	numCars := int(payload[0])
	body := payload[1:]
	perCar := len(body) / maxCars
	if perCar < 55 {
		return out
	}
	if numCars > maxCars {
		numCars = maxCars
	}

	for i := 0; i < numCars; i++ {
		rec := body[i*perCar : (i+1)*perCar]
		out[i] = Participant{
			DriverID:   int(rec[1]),
			TeamID:     rec[3],
			RaceNumber: int(rec[5]),
			Name:       cString(rec[7:55]),
		}
	}
	return out
}

const (
	statusActualCompound = 25
	statusVisualCompound = 26
	statusTyresAge       = 27
)

func ParseTyres(payload []byte) map[int]TyreStatus {
	out := map[int]TyreStatus{}
	perCar := len(payload) / maxCars
	if perCar <= statusTyresAge {
		return out
	}

	for i := 0; i < maxCars; i++ {
		rec := payload[i*perCar : (i+1)*perCar]
		visual := rec[statusVisualCompound]
		if visual == 0 {
			continue
		}
		out[i] = TyreStatus{
			ActualCompound: rec[statusActualCompound],
			Compound:       compoundName(visual),
			AgeLaps:        int(rec[statusTyresAge]),
		}
	}
	return out
}

// ParseLapData leest het Lap Data-blok van één auto. Achter de 22 auto's staan nog
// 2 bytes voor time trial, die tellen niet mee in de grootte per auto.
func ParseLapData(payload []byte, car int) (LapData, error) {
	var lap LapData
	perCar := (len(payload) - 2) / maxCars
	if perCar < binary.Size(lap) {
		return lap, fmt.Errorf("lap data too short: %d bytes", len(payload))
	}
	if car < 0 || car >= maxCars {
		return lap, fmt.Errorf("car index %d out of range", car)
	}

	rec := bytes.NewReader(payload[car*perCar : (car+1)*perCar])
	err := binary.Read(rec, binary.LittleEndian, &lap)
	return lap, err
}

// tyreStints zet de stints uit de eindklassering om naar band + eindronde.
// De arrays hebben altijd 8 plekken, alleen de eerste NumTyreStints zijn gevuld.
func tyreStints(tail classificationTail) []database.TyreStint {
	count := int(tail.NumTyreStints)
	if count > len(tail.TyreStintsVisual) {
		count = len(tail.TyreStintsVisual)
	}

	stints := make([]database.TyreStint, 0, count)
	for s := 0; s < count; s++ {
		stints = append(stints, database.TyreStint{
			Compound: compoundName(tail.TyreStintsVisual[s]),
			EndLap:   int(tail.TyreStintsEndLaps[s]),
		})
	}
	return stints
}

func ParseFinalClassification(db *sql.DB, header PacketHeader, payload []byte, participants map[int]Participant, currentTrackID int8, currentSessionType string) (FinalRaceResult, error) {

	if currentTrackID == -1 {
		return FinalRaceResult{}, fmt.Errorf("currentTrackID is not set, cannot determine track name")
	}

	result := FinalRaceResult{
		SessionUID:  fmt.Sprintf("%d", header.SessionUID),
		RecordedAt:  time.Now(),
		NumCars:     0,
		TrackName:   track.Name(currentTrackID),
		Drivers:     []DriverResult{},
		SessionType: currentSessionType,
	}

	if len(payload) < 1 {
		return result, fmt.Errorf("packet too small: %d bytes", len(payload))
	}

	numCars := int(payload[0])
	body := payload[1:]
	perCar := len(body) / maxCars

	headSize := binary.Size(classificationHead{})
	tailSize := binary.Size(classificationTail{})

	if perCar < headSize+tailSize {
		return result, fmt.Errorf("unexpected packet size: %d bytes per car, expected at least %d", perCar, headSize+tailSize)
	}

	hasResultReason := perCar >= headSize+tailSize+1

	if numCars > maxCars {
		numCars = maxCars
	}

	result.NumCars = numCars

	for i := 0; i < numCars; i++ {
		rec := bytes.NewReader(body[i*perCar : (i+1)*perCar])

		var head classificationHead
		var reason uint8
		var tail classificationTail

		if err := binary.Read(rec, binary.LittleEndian, &head); err != nil {
			return result, err
		}

		if hasResultReason {
			if err := binary.Read(rec, binary.LittleEndian, &reason); err != nil {
				return result, err
			}
		}

		if err := binary.Read(rec, binary.LittleEndian, &tail); err != nil {
			return result, err
		}

		result.Drivers = append(result.Drivers, DriverResult{
			Position:         int(head.Position),
			RaceNumber:       participants[i].RaceNumber,
			Name:             participants[i].Name,
			NumLaps:          int(head.NumLaps),
			TotalRaceTime:    FormatRaceTime(tail.TotalRaceTime),
			BestLapTime:      FormatLapTime(tail.BestLapTimeInMS),
			Points:           int(head.Points),
			ResultStatusText: resultStatusName(head.ResultStatus),
			TeamName:         mapTeam(participants[i].TeamID),
			TyreStints:       tyreStints(tail),
		})
	}

	databaseResult := database.RaceResult{
		SessionUID:  result.SessionUID,
		SessionType: result.SessionType,
	}
	for _, driver := range result.Drivers {
		databaseResult.Drivers = append(databaseResult.Drivers, database.DriverResult{
			Position:         driver.Position,
			RaceNumber:       driver.RaceNumber,
			Name:             driver.Name,
			NumLaps:          driver.NumLaps,
			TotalRaceTime:    driver.TotalRaceTime,
			BestLapTime:      driver.BestLapTime,
			Points:           driver.Points,
			ResultStatusText: driver.ResultStatusText,
			TeamName:         driver.TeamName,
			TyreStints:       driver.TyreStints,
		})
	}

	if err := database.SaveRaceResult(db, databaseResult, currentTrackID); err != nil {
		return result, err
	}

	n, err := database.ApplyPendingPenalties(db, databaseResult.SessionUID)
	
	if err != nil {
    	return result, err
	}
	
	if n > 0 {
    	fmt.Printf("%d carried over penalties applied\n", n)
	}

	return result, nil
}
