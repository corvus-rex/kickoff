package player_test

import (
	"testing"

	"kickoff/internal/auth"
	"kickoff/internal/player"
	"kickoff/internal/testutil"
)

func TestPlayerService_JerseyUniqueness(t *testing.T) {
	db := testutil.Begin(t)
	svc := player.NewService(player.NewRepository(db), db)
	admin := testutil.CreateUser(t, db, "admin", "admin@t", auth.RoleAdmin)

	teamA := testutil.CreateTeam(t, db, "TeamA", nil)
	teamB := testutil.CreateTeam(t, db, "TeamB", nil)

	// Create first player in team A with jersey 10
	p1 := testutil.CreatePlayer(t, db, teamA.ID, "P1", player.PositionStriker, 10)
	_ = p1

	t.Run("duplicate jersey in same team rejected", func(t *testing.T) {
		err := svc.Create(&player.Player{
			TeamID:       teamA.ID,
			Name:         "P2",
			Position:     player.PositionMidfielder,
			JerseyNumber: 10,
		}, admin.ID, auth.RoleAdmin)
		if err != player.ErrJerseyNumberTaken {
			t.Fatalf("expected ErrJerseyNumberTaken, got %v", err)
		}
	})

	t.Run("same jersey in different team allowed", func(t *testing.T) {
		err := svc.Create(&player.Player{
			TeamID:       teamB.ID,
			Name:         "P3",
			Position:     player.PositionMidfielder,
			JerseyNumber: 10,
		}, admin.ID, auth.RoleAdmin)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("keep existing jersey on update allowed", func(t *testing.T) {
		_, err := svc.Update(&player.Player{
			ID:           p1.ID,
			Name:         "P1 Updated",
			JerseyNumber: 10,
		}, admin.ID, auth.RoleAdmin)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("change to taken jersey on update rejected", func(t *testing.T) {
		// P4 is in team B with jersey 9
		p4 := testutil.CreatePlayer(t, db, teamB.ID, "P4", player.PositionDefender, 9)

		// Try to change p4's jersey to 10 (taken in team B)
		_, err := svc.Update(&player.Player{
			ID:           p4.ID,
			Name:         "P4",
			JerseyNumber: 10,
		}, admin.ID, auth.RoleAdmin)
		if err != player.ErrJerseyNumberTaken {
			t.Fatalf("expected ErrJerseyNumberTaken, got %v", err)
		}
	})
}

func TestPlayerService_Validation(t *testing.T) {
	db := testutil.Begin(t)
	svc := player.NewService(player.NewRepository(db), db)
	admin := testutil.CreateUser(t, db, "admin", "admin@t", auth.RoleAdmin)
	teamA := testutil.CreateTeam(t, db, "TeamA", nil)

	tests := []struct {
		name    string
		player  player.Player
		wantErr error
	}{
		{"valid player", player.Player{TeamID: teamA.ID, Name: "P1", Position: player.PositionStriker, JerseyNumber: 7}, nil},
		{"name required", player.Player{TeamID: teamA.ID, Name: "", Position: player.PositionStriker, JerseyNumber: 8}, player.ErrNameRequired},
		{"invalid position", player.Player{TeamID: teamA.ID, Name: "P2", Position: "COACH", JerseyNumber: 9}, player.ErrInvalidPosition},
		{"jersey required", player.Player{TeamID: teamA.ID, Name: "P3", Position: player.PositionStriker, JerseyNumber: 0}, player.ErrJerseyNumberRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Create(&tt.player, admin.ID, auth.RoleAdmin)
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

func TestPlayerService_Authorization(t *testing.T) {
	db := testutil.Begin(t)
	svc := player.NewService(player.NewRepository(db), db)
	admin := testutil.CreateUser(t, db, "admin", "admin@t", auth.RoleAdmin)
	mgr := testutil.CreateUser(t, db, "mgr", "mgr@t", auth.RoleManager)
	user := testutil.CreateUser(t, db, "user", "user@t", auth.RoleUser)

	// Team managed by mgr
	teamA := testutil.CreateTeam(t, db, "TeamA", &mgr.ID)
	// Team with no manager
	teamB := testutil.CreateTeam(t, db, "TeamB", nil)

	tests := []struct {
		name    string
		teamID  uint
		userID  uint
		role    auth.Role
		wantErr error
	}{
		{"admin can create in any team", teamA.ID, admin.ID, auth.RoleAdmin, nil},
		{"manager can create in own team", teamA.ID, mgr.ID, auth.RoleManager, nil},
		{"manager cannot create in other team", teamB.ID, mgr.ID, auth.RoleManager, player.ErrForbidden},
		{"user cannot create", teamA.ID, user.ID, auth.RoleUser, player.ErrForbidden},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Create(&player.Player{
				TeamID:       tt.teamID,
				Name:         "Test",
				Position:     player.PositionStriker,
				JerseyNumber: 90 + i,
			}, tt.userID, tt.role)
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
