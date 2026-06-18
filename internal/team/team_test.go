package team_test

import (
	"testing"

	"kickoff/internal/auth"
	"kickoff/internal/team"
	"kickoff/internal/testutil"
)

func TestTeamService_Create(t *testing.T) {
	db := testutil.Begin(t)
	svc := team.NewService(team.NewRepository(db))

	tests := []struct {
		name    string
		team    team.Team
		role    auth.Role
		wantErr error
	}{
		{"admin can create", team.Team{Name: "T1", FoundedYear: 2020}, auth.RoleAdmin, nil},
		{"manager cannot create", team.Team{Name: "T2", FoundedYear: 2020}, auth.RoleManager, team.ErrForbidden},
		{"user cannot create", team.Team{Name: "T3", FoundedYear: 2020}, auth.RoleUser, team.ErrForbidden},
		{"name required", team.Team{Name: "", FoundedYear: 2020}, auth.RoleAdmin, team.ErrNameRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Create(&tt.team, tt.role)
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

func TestTeamService_Update_Ownership(t *testing.T) {
	db := testutil.Begin(t)
	svc := team.NewService(team.NewRepository(db))

	admin := testutil.CreateUser(t, db, "admin", "admin@t", auth.RoleAdmin)
	mgr1 := testutil.CreateUser(t, db, "mgr1", "mgr1@t", auth.RoleManager)
	mgr2 := testutil.CreateUser(t, db, "mgr2", "mgr2@t", auth.RoleManager)
	user := testutil.CreateUser(t, db, "user", "user@t", auth.RoleUser)

	tests := []struct {
		name    string
		userID  uint
		role    auth.Role
		isOwner bool
		wantErr error
	}{
		{"admin_updates_any", admin.ID, auth.RoleAdmin, false, nil},
		{"manager_updates_own", mgr1.ID, auth.RoleManager, true, nil},
		{"manager_not_other", mgr1.ID, auth.RoleManager, false, team.ErrForbidden},
		{"user_cannot_update", user.ID, auth.RoleUser, false, team.ErrForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := "T_" + tt.name
			tm := testutil.CreateTeam(t, db, name, nil)
			if tt.isOwner {
				db.Model(tm).Update("manager_user_id", tt.userID)
			}
			if tt.name == "manager_not_other" {
				otherName := "T_other_" + tt.name
				otherTm := testutil.CreateTeam(t, db, otherName, nil)
				db.Model(otherTm).Update("manager_user_id", mgr2.ID)
				tm = otherTm
			}
			_, err := svc.Update(&team.Team{ID: tm.ID, Name: name + "_upd", FoundedYear: 2020}, tt.userID, tt.role)
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

func TestTeamService_Delete(t *testing.T) {
	db := testutil.Begin(t)
	svc := team.NewService(team.NewRepository(db))

	team1 := testutil.CreateTeam(t, db, "TeamA", nil)

	tests := []struct {
		name    string
		id      uint
		role    auth.Role
		wantErr error
	}{
		{"admin can delete", team1.ID, auth.RoleAdmin, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Delete(tt.id, tt.role)
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

func TestTeamService_Delete_Forbidden(t *testing.T) {
	db := testutil.Begin(t)
	svc := team.NewService(team.NewRepository(db))

	team1 := testutil.CreateTeam(t, db, "TeamA", nil)

	if err := svc.Delete(team1.ID, auth.RoleManager); err != team.ErrForbidden {
		t.Fatalf("expected ErrForbidden for manager, got %v", err)
	}
	if err := svc.Delete(team1.ID, auth.RoleUser); err != team.ErrForbidden {
		t.Fatalf("expected ErrForbidden for user, got %v", err)
	}
}
