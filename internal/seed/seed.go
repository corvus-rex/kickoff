package seed

import (
	"log"
	"time"

	"gorm.io/gorm"

	"kickoff/internal/match"
	"kickoff/internal/player"
	"kickoff/internal/team"
)

func Seed(db *gorm.DB) error {
	log.Println("seeding domain data...")

	clearDomain(db)

	return seedDomain(db)
}

func clearDomain(db *gorm.DB) {
	db.Exec("TRUNCATE TABLE players, matches, teams RESTART IDENTITY CASCADE")
	log.Println("cleared existing domain data and reset sequences")
}

type teamSeed struct {
	Name                string
	LogoURL             string
	FoundedYear         int
	HeadquartersAddress string
	HeadquartersCity    string
	ManagerUserID       *uint
	players             []playerSeed
}

type playerSeed struct {
	Name         string
	HeightCm     float64
	WeightKg     float64
	Position     player.Position
	JerseyNumber int
}

func seedDomain(db *gorm.DB) error {
	managerID := uint(2)

	teams := []teamSeed{
		{
			Name:             "Jakarta Mavericks",
			FoundedYear:      2015,
			HeadquartersCity: "Jakarta",
			ManagerUserID:    &managerID,
			players: []playerSeed{
				{Name: "Player A", HeightCm: 178, WeightKg: 72, Position: player.PositionStriker, JerseyNumber: 10},
				{Name: "Player B", HeightCm: 172, WeightKg: 68, Position: player.PositionDefender, JerseyNumber: 14},
				{Name: "Player C", HeightCm: 180, WeightKg: 76, Position: player.PositionGoalkeeper, JerseyNumber: 1},
				{Name: "Player D", HeightCm: 175, WeightKg: 70, Position: player.PositionMidfielder, JerseyNumber: 23},
			},
		},
		{
			Name:             "Bali Dragon",
			FoundedYear:      2014,
			HeadquartersCity: "Bali",
			players: []playerSeed{
				{Name: "Player E", HeightCm: 185, WeightKg: 78, Position: player.PositionStriker, JerseyNumber: 9},
				{Name: "Player F", HeightCm: 182, WeightKg: 74, Position: player.PositionMidfielder, JerseyNumber: 10},
				{Name: "Player G", HeightCm: 170, WeightKg: 65, Position: player.PositionMidfielder, JerseyNumber: 7},
				{Name: "Player H", HeightCm: 183, WeightKg: 80, Position: player.PositionDefender, JerseyNumber: 4},
			},
		},
		{
			Name:             "Bandung Giants",
			FoundedYear:      1933,
			HeadquartersCity: "Bandung",
			players: []playerSeed{
				{Name: "Player I", HeightCm: 186, WeightKg: 80, Position: player.PositionDefender, JerseyNumber: 3},
				{Name: "Player J", HeightCm: 168, WeightKg: 60, Position: player.PositionMidfielder, JerseyNumber: 13},
				{Name: "Player K", HeightCm: 178, WeightKg: 72, Position: player.PositionGoalkeeper, JerseyNumber: 14},
				{Name: "Player L", HeightCm: 182, WeightKg: 75, Position: player.PositionStriker, JerseyNumber: 9},
			},
		},
	}

	for _, ts := range teams {
		t := team.Team{
			Name:                ts.Name,
			LogoURL:             ts.LogoURL,
			FoundedYear:         ts.FoundedYear,
			HeadquartersAddress: ts.HeadquartersAddress,
			HeadquartersCity:    ts.HeadquartersCity,
			ManagerUserID:       ts.ManagerUserID,
		}
		if err := db.Create(&t).Error; err != nil {
			return err
		}

		for _, ps := range ts.players {
			p := player.Player{
				TeamID:       t.ID,
				Name:         ps.Name,
				HeightCm:     ps.HeightCm,
				WeightKg:     ps.WeightKg,
				Position:     ps.Position,
				JerseyNumber: ps.JerseyNumber,
			}
			if err := db.Create(&p).Error; err != nil {
				return err
			}
		}

		log.Printf("  seeded team '%s' (ID=%d) with %d players", t.Name, t.ID, len(ts.players))
	}

	if err := seedMatches(db); err != nil {
		return err
	}

	log.Println("domain seeding complete")
	return nil
}

func seedMatches(db *gorm.DB) error {
	today := time.Now().Truncate(24 * time.Hour)

	matches := []match.Match{
		{MatchDate: today, MatchTime: "19:00", HomeTeamID: 1, AwayTeamID: 2, Status: match.StatusScheduled},
		{MatchDate: today.AddDate(0, 0, 7), MatchTime: "15:30", HomeTeamID: 3, AwayTeamID: 1, Status: match.StatusScheduled},
		{MatchDate: today.AddDate(0, 0, 7), MatchTime: "19:00", HomeTeamID: 2, AwayTeamID: 3, Status: match.StatusScheduled},
	}

	for i := range matches {
		if err := db.Create(&matches[i]).Error; err != nil {
			return err
		}
		log.Printf("  seeded match #%d: team %d vs team %d", matches[i].ID, matches[i].HomeTeamID, matches[i].AwayTeamID)
	}
	return nil
}
