package testutil

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"kickoff/internal/auth"
	"kickoff/internal/config"
	"kickoff/internal/database"
	"kickoff/internal/goal"
	"kickoff/internal/match"
	"kickoff/internal/player"
	"kickoff/internal/team"
)

func projectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}

var globalDB *gorm.DB

func InitDB(t testing.TB) *gorm.DB {
	t.Helper()
	if globalDB != nil {
		return globalDB
	}
	_ = godotenv.Load(filepath.Join(projectRoot(), ".env"))
	cfg := config.Load()
	cfg.SeedUsers = false
	cfg.SeedDomain = false

	db, err := database.Connect(cfg)
	if err != nil {
		t.Skipf("database unavailable: %v", err)
		return nil
	}

	database.RegisterModel(&auth.User{}, &team.Team{}, &player.Player{}, &match.Match{}, &goal.Goal{})
	if err := database.RunMigrations(db); err != nil {
		t.Skipf("migration failed: %v", err)
		return nil
	}

	globalDB = db
	return db
}

func Begin(t *testing.T) *gorm.DB {
	t.Helper()
	db := InitDB(t)
	if db == nil {
		return nil
	}
	tx := db.Begin()
	t.Cleanup(func() { tx.Rollback() })
	return tx
}

func CreateUser(t *testing.T, db *gorm.DB, name, email string, role auth.Role) *auth.User {
	t.Helper()
	hash, err := auth.HashPassword("test-pass")
	if err != nil {
		t.Fatal(err)
	}
	u := &auth.User{Name: name, Email: email, PasswordHash: hash, Role: role}
	if err := db.Create(u).Error; err != nil {
		t.Fatal(err)
	}
	return u
}

func CreateTeam(t *testing.T, db *gorm.DB, name string, managerID *uint) *team.Team {
	t.Helper()
	tm := &team.Team{Name: name, FoundedYear: 2020, ManagerUserID: managerID}
	if err := db.Create(tm).Error; err != nil {
		t.Fatal(err)
	}
	return tm
}

func CreatePlayer(t *testing.T, db *gorm.DB, teamID uint, name string, position player.Position, jersey int) *player.Player {
	t.Helper()
	p := &player.Player{TeamID: teamID, Name: name, Position: position, JerseyNumber: jersey}
	if err := db.Create(p).Error; err != nil {
		t.Fatal(err)
	}
	return p
}

func CreateMatch(t *testing.T, db *gorm.DB, homeID, awayID uint, dateStr, timeStr string, status match.MatchStatus) *match.Match {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		t.Fatal(err)
	}
	m := &match.Match{
		MatchDate:  parsed,
		MatchTime:  timeStr,
		HomeTeamID: homeID,
		AwayTeamID: awayID,
		Status:     status,
	}
	if err := db.Create(m).Error; err != nil {
		t.Fatal(err)
	}
	return m
}

func CreateGoal(t *testing.T, db *gorm.DB, matchID, playerID uint, minute int) *goal.Goal {
	t.Helper()
	g := &goal.Goal{MatchID: matchID, PlayerID: playerID, GoalMinute: minute}
	if err := db.Create(g).Error; err != nil {
		t.Fatal(err)
	}
	return g
}
