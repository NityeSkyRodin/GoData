package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	"GoData/database"
	"GoData/exporter"
)

func main() {
	db, err := database.InitDB("f2_results.db")
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}
	defer db.Close()

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:20777")
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error opening port 20777:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Go is listening on port 20777 for F2 data...")

	buf := make([]byte, 2048)
	participants := map[int]Participant{}
	knownDrivers := 0
	var participantSessionUID uint64
	var lastResultUID uint64

	var currentTrackID int8 = -1
	var currentSessionType string = "Unknown"

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil || n < headerSize {
			continue
		}

		var header PacketHeader
		if err := binary.Read(bytes.NewReader(buf[:headerSize]), binary.LittleEndian, &header); err != nil {
			continue
		}

		payload := buf[headerSize:n]

		switch header.PacketID {

		case 1: // PACKET_SESSION_DATA
			if len(payload) >= 9 {
				sessionTypeID := payload[6]
				trackID := int8(payload[7])

				currentTrackID = trackID
				currentSessionType = parseSessionType(sessionTypeID)
			}

		case 2:

		case 4:
			if header.SessionUID != participantSessionUID {
				participantSessionUID = header.SessionUID
				p := parseParticipants(payload)
				if len(p) == 0 {
					continue
				}

				participants = mapParticipants(p)

				if knownDrivers != len(participants) {
					knownDrivers = len(participants)
					fmt.Printf("\n[LOBBY] Driverlist data received: %d Drivers\n", knownDrivers)
				}
			}

		case 8:
			if header.SessionUID == lastResultUID {
				continue
			}

			result, err := parseFinalClassification(db, header, payload, participants, currentTrackID, currentSessionType)
			if err != nil {
				fmt.Println("\nError parsing final classification:", err)
				continue
			}
			lastResultUID = header.SessionUID

			printResult(result)

			if err := exporter.ParseFullSeasonResult(db); err != nil {
				fmt.Println("Error retrieving season results:", err)
			}

			if err := exporter.ParseRaceWeekendResults(db, result.TrackName); err != nil {
				fmt.Println("Error retrieving race weekend results:", err)
			}
		}
	}
}

func parseSessionType(sessionType uint8) string {
	switch sessionType {
	case 5, 6, 7, 8, 9:
		return "Qualifying"
	case 15:
		return "Sprint"
	case 16:
		return "Race"
	default:
		return "Practice"
	}
}

func parseFormula(formula uint8) string {
	if formula == 2 {
		return "F2"
	}
	return "F1"
}

func mapParticipants(participants map[int]Participant) map[int]Participant {
	for index, participant := range participants {
		fmt.Printf(
			"Mapping participant: %s | TeamID: %v | Type: %T\n",
			participant.Name,
			participant.TeamID,
			participant.TeamID,
		)

		participants[index] = participant
	}

	return participants
}

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
