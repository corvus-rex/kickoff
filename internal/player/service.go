package player

import (
	"errors"

	"gorm.io/gorm"

	"kickoff/internal/auth"
)

var (
	ErrForbidden            = errors.New("access denied")
	ErrNotFound             = errors.New("player not found")
	ErrTeamNotFound         = errors.New("team not found")
	ErrNameRequired         = errors.New("player name is required")
	ErrInvalidPosition      = errors.New("invalid position")
	ErrJerseyNumberTaken    = errors.New("jersey number already taken in this team")
	ErrJerseyNumberRequired = errors.New("jersey number is required")
)

type Service struct {
	repo *Repository
	db   *gorm.DB
}

func NewService(repo *Repository, db *gorm.DB) *Service {
	return &Service{repo: repo, db: db}
}

func (s *Service) ListByTeam(teamID uint) ([]Player, error) {
	return s.repo.FindByTeam(teamID)
}

func (s *Service) GetByID(id uint) (*Player, error) {
	player, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return player, nil
}

func (s *Service) Create(player *Player, userID uint, role auth.Role) error {
	if player.Name == "" {
		return ErrNameRequired
	}
	if !player.Position.IsValid() {
		return ErrInvalidPosition
	}
	if player.JerseyNumber == 0 {
		return ErrJerseyNumberRequired
	}

	if err := s.verifyTeamAccess(player.TeamID, userID, role); err != nil {
		return err
	}

	taken, err := s.repo.IsJerseyNumberTaken(player.TeamID, player.JerseyNumber, nil)
	if err != nil {
		return err
	}
	if taken {
		return ErrJerseyNumberTaken
	}

	return s.repo.Create(player)
}

func (s *Service) Update(updates *Player, userID uint, role auth.Role) (*Player, error) {
	existing, err := s.repo.FindByID(updates.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if err := s.verifyTeamAccess(existing.TeamID, userID, role); err != nil {
		return nil, err
	}

	if updates.Name == "" {
		return nil, ErrNameRequired
	}
	if updates.Position != "" && !updates.Position.IsValid() {
		return nil, ErrInvalidPosition
	}

	if updates.JerseyNumber > 0 && updates.JerseyNumber != existing.JerseyNumber {
		taken, err := s.repo.IsJerseyNumberTaken(existing.TeamID, updates.JerseyNumber, &existing.ID)
		if err != nil {
			return nil, err
		}
		if taken {
			return nil, ErrJerseyNumberTaken
		}
		existing.JerseyNumber = updates.JerseyNumber
	}

	existing.Name = updates.Name
	if updates.HeightCm > 0 {
		existing.HeightCm = updates.HeightCm
	}
	if updates.WeightKg > 0 {
		existing.WeightKg = updates.WeightKg
	}
	if updates.Position != "" {
		existing.Position = updates.Position
	}

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) Delete(id uint, userID uint, role auth.Role) error {
	player, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if err := s.verifyTeamAccess(player.TeamID, userID, role); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func (s *Service) verifyTeamAccess(teamID uint, userID uint, role auth.Role) error {
	if role == auth.RoleAdmin {
		return nil
	}
	if role != auth.RoleManager {
		return ErrForbidden
	}
	var managerUserID *uint
	err := s.db.Model(&struct{ ManagerUserID *uint }{}).
		Table("teams").
		Select("manager_user_id").
		Where("id = ? AND deleted_at IS NULL", teamID).
		Scan(&managerUserID).Error
	if err != nil {
		return ErrTeamNotFound
	}
	if managerUserID == nil || *managerUserID != userID {
		return ErrForbidden
	}
	return nil
}
