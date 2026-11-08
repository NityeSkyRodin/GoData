package telemetry

import (
	"time"

	"GoData/database"
)

// HeaderSize is de vaste grootte van de pakket-header in bytes.
const HeaderSize = 29
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

// LapData is het blok van één auto uit het Lap Data-pakket (packet 2),
// in de volgorde van de F1 25-specificatie. 57 bytes, zonder opvulling ertussen.
type LapData struct {
	LastLapTimeInMS              uint32
	CurrentLapTimeInMS           uint32
	Sector1TimeMSPart            uint16
	Sector1TimeMinutesPart       uint8
	Sector2TimeMSPart            uint16
	Sector2TimeMinutesPart       uint8
	DeltaToCarInFrontMSPart      uint16
	DeltaToCarInFrontMinutesPart uint8
	DeltaToRaceLeaderMSPart      uint16
	DeltaToRaceLeaderMinutesPart uint8
	LapDistance                  float32 // meter in deze ronde
	TotalDistance                float32 // meter in de hele sessie
	SafetyCarDelta               float32
	CarPosition                  uint8
	CurrentLapNum                uint8
	PitStatus                    uint8 // 0 = baan, 1 = pitlane, 2 = pitbox
	NumPitStops                  uint8
	Sector                       uint8 // 0, 1 of 2
	CurrentLapInvalid            uint8
	Penalties                    uint8 // opgebouwde tijdstraf in seconden
	TotalWarnings                uint8
	CornerCuttingWarnings        uint8
	NumUnservedDriveThroughPens  uint8
	NumUnservedStopGoPens        uint8
	GridPosition                 uint8
	DriverStatus                 uint8 // 0 = garage, 1 = vliegende ronde, 2 = in-lap, 3 = out-lap, 4 = op de baan
	ResultStatus                 uint8
	PitLaneTimerActive           uint8
	PitLaneTimeInLaneInMS        uint16
	PitStopTimerInMS             uint16
	PitStopShouldServePen        uint8
	SpeedTrapFastestSpeed        float32 // km/u
	SpeedTrapFastestLap          uint8
}

// TyreStatus is de band van één auto uit het Car Status-pakket (packet 7).
type TyreStatus struct {
	ActualCompound uint8  // ruw id van de echte compound, nog niet vertaald
	Compound       string // wat je in de game ziet: Soft, Medium, Hard, ...
	AgeLaps        int
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
	TyreStints       []database.TyreStint
}

type FinalRaceResult struct {
	SessionUID  string
	RecordedAt  time.Time
	NumCars     int
	SessionType string
	TrackName   string
	Drivers     []DriverResult
}
