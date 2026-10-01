package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

func parseParticipants(payload []byte) map[int]Participant {
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
			DriverID:     int(rec[1]),
			TeamID:       int(rec[3]),
			RaceNumber:   int(rec[5]),
			Name:         cString(rec[7:55]),
		}
	}
	return out
}

func parseFinalClassification(header PacketHeader, payload []byte, participants map[int]Participant, currentTrackID int8) (FinalRaceResult, error) {

	if currentTrackID == -1 {
		return FinalRaceResult{}, fmt.Errorf("currentTrackID is not set, cannot determine track name")
	}

	result := FinalRaceResult{
		SessionUID:  fmt.Sprintf("%d", header.SessionUID),
		RecordedAt:  time.Now(),
		NumCars:     0,
		Drivers:     []DriverResult{},
	}

	if len(payload) < 1 {
		return result, fmt.Errorf("packet too small: %d bytes", len(payload))
	}

	numCars := int(payload[0])
	body := payload[1:]
	perCar  := len(body) / maxCars

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
    NumLaps:          int(head.NumLaps), // <-- Deze toevoegen!
    TotalRaceTime:    formatRaceTime(tail.TotalRaceTime),
    BestLapTime:      formatLapTime(tail.BestLapTimeInMS),
    Points:           int(head.Points),
    ResultStatusText: resultStatusName(head.ResultStatus),
})
	}

	database, err := InitDB("f2_results.db")
	if err != nil {
		return result, fmt.Errorf("error initializing database: %w", err)
	}

	SaveRaceResult(database, result, currentTrackID)


	return result, nil
}
