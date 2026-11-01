package main

import (
	"fmt"
	"log"

	"GoData/database"
	"GoData/exporter"
	"GoData/git"

	"github.com/charmbracelet/huh"
)

type Penalty struct {
    Label   string
    Seconds int
}

var penalties = []Penalty{
    {"5 second time penalty", 5},
    {"10 second time penalty", 10},
	{"Drive through penalty" , 20},
	{"10 second stop and go penalty", 30},
	{"20 second stop and go penalty", 40},
    {"Disqualified", 0},
}

var sessionTypes = []string{"Qualifying", "Sprint", "Race"}

//TODO: when DSQ or DNF add 5 sec to next race.
func main() {
    db, err := database.InitDB("f2_results.db")
    if err != nil {
        log.Fatal("could not open database: ", err)
    }
    defer db.Close()

    races, err := database.GetTrackNames(db)
    if err != nil {
        log.Fatal("could not read races: ", err)
    }

    drivers, err := database.GetDriverNames(db)
    if err != nil {
        log.Fatal("could not read drivers: ", err)
    }

    var driver string

    err = huh.NewSelect[string]().
        Title("Choose driver").
        Options(huh.NewOptions(drivers...)...).
        Value(&driver).
        Run()

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Selected driver:", driver)

    var race string

    err = huh.NewSelect[string]().
        Title("Choose race").
        Options(huh.NewOptions(races...)...).
        Value(&race).
        Run()

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Selected race:", race)

    var sessionType string

    err = huh.NewSelect[string]().
        Title("Choose session type").
        Options(huh.NewOptions(sessionTypes...)...).
        Value(&sessionType).
        Run()

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Selected session type:", sessionType)

    options := make([]huh.Option[Penalty], 0, len(penalties))
    for _, p := range penalties {
        options = append(options, huh.NewOption(p.Label, p))
    }

    var penalty Penalty

    err = huh.NewSelect[Penalty]().
        Title("Choose penalty").
        Options(options...).
        Value(&penalty).
        Run()

    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("\n%s | %s | %s | %s (+%d sec)\n",
        driver, race, sessionType, penalty.Label, penalty.Seconds)


    carriedOver, err := database.ApplyPenalty(db, database.PenaltyInput{
        TrackName:    race,
        SessionType:  sessionType,
        DriverName:   driver,
        Seconds:      penalty.Seconds,
        Disqualified: penalty.Seconds == 0,
    })

    if err != nil {
        log.Fatal(err)
    }

    if carriedOver {
        fmt.Printf("%s did not finish, 5 sec penalty moves to the next race\n", driver)
    }

    if err := exporter.ParseFullSeasonResult(db); err != nil {
        log.Fatal("could not export season results: ", err)
    }

    if err := exporter.ParseRaceWeekendResults(db, race); err != nil {
        log.Fatal("could not export race weekend: ", err)
    }

    git.CommitRaceResult(race, sessionType)
}