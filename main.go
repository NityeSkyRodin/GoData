package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

func main() {

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
	var lastResultUID uint64
	
	var currentTrackID int8 = -1 

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

		case 1:
			if len(payload) > 2 {
				currentTrackID = int8(payload[2])
			}

		case 2:
			fmt.Printf("\r[LAP DATA] Frame: %d | Player Car Index: %d  ", header.FrameIdentifier, header.PlayerCarIndex)

		case 4:
			if p := parseParticipants(payload); len(p) > 0 {
				participants = p
				if knownDrivers != len(p) {
					knownDrivers = len(p)
					fmt.Printf("\n[LOBBY] Driverlist data received: %d Drivers\n", len(p))
				}
			}

		case 8:
			if header.SessionUID == lastResultUID {
				continue
			}

			result, err := parseFinalClassification(header, payload, participants, currentTrackID)
			if err != nil {
				fmt.Println("\nError parsing final classification:", err)
				continue
			}
			lastResultUID = header.SessionUID

			printResult(result)
		}
	}
}