package match_test

import (
	"testing"
	"time"

	"kickoff/internal/auth"
	"kickoff/internal/match"
	"kickoff/internal/testutil"
)

func TestMatchService_Create_Validation(t *testing.T) {
	db := testutil.Begin(t)
	svc := match.NewService(match.NewRepository(db), db)

	teamA := testutil.CreateTeam(t, db, "TeamA", nil)
	teamB := testutil.CreateTeam(t, db, "TeamB", nil)

	tests := []struct {
		name    string
		m       match.Match
		wantErr error
	}{
		{
			"valid match",
			match.Match{MatchDate: time.Now(), MatchTime: "20:00", HomeTeamID: teamA.ID, AwayTeamID: teamB.ID, Status: match.StatusScheduled},
			nil,
		},
		{
			"same home and away",
			match.Match{MatchDate: time.Now(), MatchTime: "20:00", HomeTeamID: teamA.ID, AwayTeamID: teamA.ID, Status: match.StatusScheduled},
			match.ErrSameTeam,
		},
		{
			"nonexistent home team",
			match.Match{MatchDate: time.Now(), MatchTime: "20:00", HomeTeamID: 999, AwayTeamID: teamB.ID, Status: match.StatusScheduled},
			match.ErrTeamNotFound,
		},
		{
			"nonexistent away team",
			match.Match{MatchDate: time.Now(), MatchTime: "20:00", HomeTeamID: teamA.ID, AwayTeamID: 999, Status: match.StatusScheduled},
			match.ErrTeamNotFound,
		},
		{
			"invalid time format",
			match.Match{MatchDate: time.Now(), MatchTime: "8pm", HomeTeamID: teamA.ID, AwayTeamID: teamB.ID, Status: match.StatusScheduled},
			match.ErrInvalidTime,
		},
		{
			"invalid status",
			match.Match{MatchDate: time.Now(), MatchTime: "20:00", HomeTeamID: teamA.ID, AwayTeamID: teamB.ID, Status: "INVALID"},
			match.ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Create(&tt.m, auth.RoleAdmin)
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

func TestMatchService_Authorization(t *testing.T) {
	db := testutil.Begin(t)
	svc := match.NewService(match.NewRepository(db), db)

	teamA := testutil.CreateTeam(t, db, "TeamA", nil)
	teamB := testutil.CreateTeam(t, db, "TeamB", nil)

	tests := []struct {
		name    string
		role    auth.Role
		wantErr error
	}{
		{"admin can create", auth.RoleAdmin, nil},
		{"manager cannot create", auth.RoleManager, match.ErrForbidden},
		{"user cannot create", auth.RoleUser, match.ErrForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Create(&match.Match{
				MatchDate: time.Now(), MatchTime: "20:00",
				HomeTeamID: teamA.ID, AwayTeamID: teamB.ID,
				Status: match.StatusScheduled,
			}, tt.role)
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

func TestMatchService_Finish(t *testing.T) {
	db := testutil.Begin(t)
	svc := match.NewService(match.NewRepository(db), db)

	teamA := testutil.CreateTeam(t, db, "TeamA", nil)
	teamB := testutil.CreateTeam(t, db, "TeamB", nil)
	m := testutil.CreateMatch(t, db, teamA.ID, teamB.ID, "2026-07-01", "20:00", match.StatusScheduled)

	t.Run("finish scheduled match", func(t *testing.T) {
		finished, err := svc.Finish(m.ID, auth.RoleAdmin)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if finished.Status != match.StatusFinished {
			t.Fatalf("expected FINISHED, got %s", finished.Status)
		}
	})

	t.Run("finish already finished match rejected", func(t *testing.T) {
		_, err := svc.Finish(m.ID, auth.RoleAdmin)
		if err != match.ErrAlreadyFinished {
			t.Fatalf("expected ErrAlreadyFinished, got %v", err)
		}
	})

	t.Run("manager cannot finish match", func(t *testing.T) {
		m2 := testutil.CreateMatch(t, db, teamA.ID, teamB.ID, "2026-07-02", "20:00", match.StatusScheduled)
		_, err := svc.Finish(m2.ID, auth.RoleManager)
		if err != match.ErrForbidden {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})
}
