package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	"GoData/database"
	"GoData/exporter"
	"GoData/git"
	"GoData/telemetry"
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
	participants := map[int]telemetry.Participant{}
	knownDrivers := 0
	var participantSessionUID uint64
	var lastResultUID uint64

	var currentTrackID int8 = -1
	var currentSessionType string = "Unknown"

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil || n < telemetry.HeaderSize {
			continue
		}

		var header telemetry.PacketHeader
		if err := binary.Read(bytes.NewReader(buf[:telemetry.HeaderSize]), binary.LittleEndian, &header); err != nil {
			continue
		}

		payload := buf[telemetry.HeaderSize:n]

		switch header.PacketID {
		case 1:
			if len(payload) >= 9 {
				currentSessionType = parseSessionType(payload[6])
				currentTrackID = int8(payload[7])
			}
		case 4:
			if header.SessionUID != participantSessionUID {
				participantSessionUID = header.SessionUID
				p := telemetry.ParseParticipants(payload)
				if len(p) == 0 {
					continue
				}

				participants = p

				if knownDrivers != len(participants) {
					knownDrivers = len(participants)
					fmt.Printf("\n[LOBBY] Driverlist data received: %d Drivers\n", knownDrivers)
				}
			}
		case 8:
			if header.SessionUID == lastResultUID {
				continue
			}

			result, err := telemetry.ParseFinalClassification(db, header, payload, participants, currentTrackID, currentSessionType)
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

			git.CommitRaceResult(result.TrackName, result.SessionType)
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
