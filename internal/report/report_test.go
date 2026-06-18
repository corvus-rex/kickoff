package report_test

import (
	"testing"

	"kickoff/internal/match"
	"kickoff/internal/player"
	"kickoff/internal/report"
	"kickoff/internal/testutil"
)

func TestGetReport_FinishedMatch(t *testing.T) {
	db := testutil.Begin(t)
	svc := report.NewService(db)

	teamHome := testutil.CreateTeam(t, db, "Home Team", nil)
	teamAway := testutil.CreateTeam(t, db, "Away Team", nil)

	// Home players
	hp1 := testutil.CreatePlayer(t, db, teamHome.ID, "HScorer", player.PositionStriker, 10)
	hp2 := testutil.CreatePlayer(t, db, teamHome.ID, "HOther", player.PositionMidfielder, 8)
	// Away players
	ap1 := testutil.CreatePlayer(t, db, teamAway.ID, "AScorer", player.PositionStriker, 9)

	// Create finished match
	m := testutil.CreateMatch(t, db, teamHome.ID, teamAway.ID, "2026-07-01", "20:00", match.StatusFinished)

	// Add goals: home scores 3 (hp1 scores 2, hp2 scores 1), away scores 1
	testutil.CreateGoal(t, db, m.ID, hp1.ID, 10)
	testutil.CreateGoal(t, db, m.ID, hp1.ID, 45)
	testutil.CreateGoal(t, db, m.ID, hp2.ID, 60)
	testutil.CreateGoal(t, db, m.ID, ap1.ID, 80)

	r, err := svc.GetReport(m.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.HomeScore != 3 {
		t.Fatalf("expected HomeScore 3, got %d", r.HomeScore)
	}
	if r.AwayScore != 1 {
		t.Fatalf("expected AwayScore 1, got %d", r.AwayScore)
	}
	if r.Result != "HOME_WIN" {
		t.Fatalf("expected HOME_WIN, got %s", r.Result)
	}

	if r.TopScorer == nil {
		t.Fatal("expected top scorer, got nil")
	}
	if r.TopScorer.PlayerID != hp1.ID {
		t.Fatalf("expected top scorer ID %d, got %d", hp1.ID, r.TopScorer.PlayerID)
	}
	if r.TopScorer.Goals != 2 {
		t.Fatalf("expected top scorer goals 2, got %d", r.TopScorer.Goals)
	}

	if r.HomeTeam.ID != teamHome.ID || r.AwayTeam.ID != teamAway.ID {
		t.Fatal("team IDs mismatch")
	}
}

func TestGetReport_Draw(t *testing.T) {
	db := testutil.Begin(t)
	svc := report.NewService(db)

	th := testutil.CreateTeam(t, db, "T1", nil)
	ta := testutil.CreateTeam(t, db, "T2", nil)
	p1 := testutil.CreatePlayer(t, db, th.ID, "P1", player.PositionStriker, 10)
	p2 := testutil.CreatePlayer(t, db, ta.ID, "P2", player.PositionStriker, 9)
	m := testutil.CreateMatch(t, db, th.ID, ta.ID, "2026-07-01", "20:00", match.StatusFinished)
	testutil.CreateGoal(t, db, m.ID, p1.ID, 15)
	testutil.CreateGoal(t, db, m.ID, p2.ID, 30)

	r, err := svc.GetReport(m.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Result != "DRAW" {
		t.Fatalf("expected DRAW, got %s", r.Result)
	}
	if r.HomeScore != 1 || r.AwayScore != 1 {
		t.Fatalf("expected 1-1, got %d-%d", r.HomeScore, r.AwayScore)
	}
}

func TestGetReport_UnfinishedMatch(t *testing.T) {
	db := testutil.Begin(t)
	svc := report.NewService(db)

	th := testutil.CreateTeam(t, db, "T1", nil)
	ta := testutil.CreateTeam(t, db, "T2", nil)
	m := testutil.CreateMatch(t, db, th.ID, ta.ID, "2026-07-01", "20:00", match.StatusScheduled)

	_, err := svc.GetReport(m.ID)
	if err != report.ErrMatchNotFinished {
		t.Fatalf("expected ErrMatchNotFinished, got %v", err)
	}
}

func TestGetReport_NonExistentMatch(t *testing.T) {
	db := testutil.Begin(t)
	svc := report.NewService(db)

	_, err := svc.GetReport(999)
	if err != report.ErrMatchNotFound {
		t.Fatalf("expected ErrMatchNotFound, got %v", err)
	}
}

func TestGetReport_CumulativeWins(t *testing.T) {
	db := testutil.Begin(t)
	svc := report.NewService(db)

	t1 := testutil.CreateTeam(t, db, "Team1", nil) // home in all matches
	t2 := testutil.CreateTeam(t, db, "Team2", nil) // away in all matches

	// Match 1: t1 2-0 t2 (t1 wins)
	m1 := testutil.CreateMatch(t, db, t1.ID, t2.ID, "2026-06-01", "18:00", match.StatusFinished)
	p1 := testutil.CreatePlayer(t, db, t1.ID, "P1", player.PositionStriker, 10)
	p2 := testutil.CreatePlayer(t, db, t1.ID, "P2", player.PositionStriker, 9)
	testutil.CreateGoal(t, db, m1.ID, p1.ID, 10)
	testutil.CreateGoal(t, db, m1.ID, p2.ID, 20)

	// Match 2: t1 0-1 t2 (t2 wins)
	m2 := testutil.CreateMatch(t, db, t1.ID, t2.ID, "2026-06-15", "18:00", match.StatusFinished)
	p3 := testutil.CreatePlayer(t, db, t2.ID, "P3", player.PositionStriker, 7)
	testutil.CreateGoal(t, db, m2.ID, p3.ID, 30)

	// Current match: t1 1-1 t2 (draw)
	m3 := testutil.CreateMatch(t, db, t1.ID, t2.ID, "2026-07-01", "20:00", match.StatusFinished)
	testutil.CreateGoal(t, db, m3.ID, p1.ID, 15)
	testutil.CreateGoal(t, db, m3.ID, p3.ID, 60)

	r, err := svc.GetReport(m3.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Team1 won match 1, lost match 2, drew match 3 → 1 cumulative win
	if r.HomeCumulativeWins != 1 {
		t.Fatalf("expected HomeCumulativeWins 1, got %d", r.HomeCumulativeWins)
	}
	// Team2 lost match 1, won match 2, drew match 3 → 1 cumulative win
	if r.AwayCumulativeWins != 1 {
		t.Fatalf("expected AwayCumulativeWins 1, got %d", r.AwayCumulativeWins)
	}
}

func TestGetReport_NoGoals(t *testing.T) {
	db := testutil.Begin(t)
	svc := report.NewService(db)

	th := testutil.CreateTeam(t, db, "T1", nil)
	ta := testutil.CreateTeam(t, db, "T2", nil)
	m := testutil.CreateMatch(t, db, th.ID, ta.ID, "2026-07-01", "20:00", match.StatusFinished)

	r, err := svc.GetReport(m.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.HomeScore != 0 || r.AwayScore != 0 {
		t.Fatalf("expected 0-0, got %d-%d", r.HomeScore, r.AwayScore)
	}
	if r.Result != "DRAW" {
		t.Fatalf("expected DRAW, got %s", r.Result)
	}
	if r.TopScorer != nil {
		t.Fatal("expected nil top scorer for 0-0")
	}
}
