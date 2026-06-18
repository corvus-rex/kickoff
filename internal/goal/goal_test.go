package goal_test

import (
	"testing"

	"kickoff/internal/auth"
	"kickoff/internal/goal"
	"kickoff/internal/match"
	"kickoff/internal/player"
	"kickoff/internal/testutil"
)

func TestGoalService_ScorerValidation(t *testing.T) {
	db := testutil.Begin(t)
	svc := goal.NewService(goal.NewRepository(db), db)

	teamHome := testutil.CreateTeam(t, db, "Home", nil)
	teamAway := testutil.CreateTeam(t, db, "Away", nil)
	teamOther := testutil.CreateTeam(t, db, "Other", nil)

	homePlayer := testutil.CreatePlayer(t, db, teamHome.ID, "HomeP", player.PositionStriker, 10)
	awayPlayer := testutil.CreatePlayer(t, db, teamAway.ID, "AwayP", player.PositionMidfielder, 8)
	otherPlayer := testutil.CreatePlayer(t, db, teamOther.ID, "OtherP", player.PositionDefender, 4)

	m := testutil.CreateMatch(t, db, teamHome.ID, teamAway.ID, "2026-07-01", "20:00", match.StatusFinished)

	tests := []struct {
		name     string
		matchID  uint
		playerID uint
		minute   int
		wantErr  error
	}{
		{"home player scores", m.ID, homePlayer.ID, 30, nil},
		{"away player scores", m.ID, awayPlayer.ID, 45, nil},
		{"player not in match", m.ID, otherPlayer.ID, 10, goal.ErrPlayerNotInMatch},
		{"invalid minute", m.ID, homePlayer.ID, 0, goal.ErrInvalidMinute},
		{"nonexistent match", 999, homePlayer.ID, 10, goal.ErrMatchNotFound},
		{"nonexistent player", m.ID, 999, 10, goal.ErrPlayerNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Create(&goal.Goal{
				MatchID:    tt.matchID,
				PlayerID:   tt.playerID,
				GoalMinute: tt.minute,
			}, auth.RoleAdmin)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestGoalService_ScorerValidation_FinishedMatch(t *testing.T) {
	db := testutil.Begin(t)
	svc := goal.NewService(goal.NewRepository(db), db)

	teamHome := testutil.CreateTeam(t, db, "Home", nil)
	teamAway := testutil.CreateTeam(t, db, "Away", nil)
	homePlayer := testutil.CreatePlayer(t, db, teamHome.ID, "HP", player.PositionStriker, 10)
	m := testutil.CreateMatch(t, db, teamHome.ID, teamAway.ID, "2026-07-01", "20:00", match.StatusFinished)

	err := svc.Create(&goal.Goal{MatchID: m.ID, PlayerID: homePlayer.ID, GoalMinute: 5}, auth.RoleAdmin)
	if err != nil {
		t.Fatalf("expected goal allowed on finished match, got %v", err)
	}
}

func TestGoalService_Authorization(t *testing.T) {
	db := testutil.Begin(t)
	svc := goal.NewService(goal.NewRepository(db), db)

	teamHome := testutil.CreateTeam(t, db, "Home", nil)
	teamAway := testutil.CreateTeam(t, db, "Away", nil)
	homePlayer := testutil.CreatePlayer(t, db, teamHome.ID, "HP", player.PositionStriker, 10)
	m := testutil.CreateMatch(t, db, teamHome.ID, teamAway.ID, "2026-07-01", "20:00", match.StatusScheduled)

	t.Run("admin can create goal", func(t *testing.T) {
		err := svc.Create(&goal.Goal{MatchID: m.ID, PlayerID: homePlayer.ID, GoalMinute: 10}, auth.RoleAdmin)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("manager cannot create goal", func(t *testing.T) {
		err := svc.Create(&goal.Goal{MatchID: m.ID, PlayerID: homePlayer.ID, GoalMinute: 20}, auth.RoleManager)
		if err != goal.ErrForbidden {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})
}
