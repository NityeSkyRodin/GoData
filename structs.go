package main

import "time"

const headerSize = 29
const maxCars = 22

type PacketHeader struct {
	PacketFormat            uint16
	GameYear                uint8
	GameMajorVersion        uint8
	GameMinorVersion        uint8
	PacketVersion           uint8
	PacketID                uint8
	SessionUID              uint64
	SessionTime             float32
	FrameIdentifier         uint32
	OverallFrameIdentifier  uint32
	PlayerCarIndex          uint8
	SecondaryPlayerCarIndex uint8
}

type classificationHead struct {
	Position     uint8
	NumLaps      uint8
	GridPosition uint8
	Points       uint8
	NumPitStops  uint8
	ResultStatus uint8
}

type classificationTail struct {
	BestLapTimeInMS   uint32
	TotalRaceTime     float64
	PenaltiesTime     uint8
	NumPenalties      uint8
	NumTyreStints     uint8
	TyreStintsActual  [8]uint8
	TyreStintsVisual  [8]uint8
	TyreStintsEndLaps [8]uint8
}

type Participant struct {
	Name       string
	RaceNumber int
	TeamID     uint8
	DriverID   int
}

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

type FinalRaceResult struct {
	SessionUID  string
	RecordedAt  time.Time
	NumCars     int
	SessionType string
	TrackName   string
	Drivers     []DriverResult
}
