package match

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"kickoff/internal/auth"
)

var (
	ErrForbidden        = errors.New("access denied")
	ErrNotFound         = errors.New("match not found")
	ErrSameTeam         = errors.New("home team and away team must be different")
	ErrTeamNotFound     = errors.New("one or both teams not found")
	ErrInvalidStatus    = errors.New("invalid match status, must be SCHEDULED or FINISHED")
	ErrAlreadyFinished  = errors.New("match is already finished")
	ErrInvalidDate      = errors.New("invalid match date")
	ErrInvalidTime      = errors.New("match time is required (HH:MM)")
	ErrDateRequired     = errors.New("match date is required")
)

type Service struct {
	repo *Repository
	db   *gorm.DB
}

func NewService(repo *Repository, db *gorm.DB) *Service {
	return &Service{repo: repo, db: db}
}

func (s *Service) List() ([]Match, error) {
	return s.repo.FindAll()
}

func (s *Service) GetByID(id uint) (*Match, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m, nil
}

func (s *Service) Create(m *Match, role auth.Role) error {
	if role != auth.RoleAdmin {
		return ErrForbidden
	}
	if err := s.validate(m); err != nil {
		return err
	}
	if m.Status == "" {
		m.Status = StatusScheduled
	}
	return s.repo.Create(m)
}

func (s *Service) Update(updates *Match, role auth.Role) (*Match, error) {
	if role != auth.RoleAdmin {
		return nil, ErrForbidden
	}

	existing, err := s.repo.FindByID(updates.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if !updates.MatchDate.IsZero() {
		existing.MatchDate = updates.MatchDate
	}
	if updates.MatchTime != "" {
		existing.MatchTime = updates.MatchTime
	}
	if updates.HomeTeamID > 0 {
		existing.HomeTeamID = updates.HomeTeamID
	}
	if updates.AwayTeamID > 0 {
		existing.AwayTeamID = updates.AwayTeamID
	}
	if updates.Status != "" {
		existing.Status = updates.Status
	}

	if err := s.validate(existing); err != nil {
		return nil, err
	}

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) Delete(id uint, role auth.Role) error {
	if role != auth.RoleAdmin {
		return ErrForbidden
	}
	_, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

func (s *Service) Finish(id uint, role auth.Role) (*Match, error) {
	if role != auth.RoleAdmin {
		return nil, ErrForbidden
	}

	m, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if m.Status == StatusFinished {
		return nil, ErrAlreadyFinished
	}

	m.Status = StatusFinished
	if err := s.repo.Update(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Service) validate(m *Match) error {
	if m.MatchDate.IsZero() {
		return ErrDateRequired
	}
	if m.MatchTime == "" {
		return ErrInvalidTime
	}
	if _, err := time.Parse("15:04", m.MatchTime); err != nil {
		return ErrInvalidTime
	}
	if m.HomeTeamID == 0 || m.AwayTeamID == 0 {
		return ErrTeamNotFound
	}
	if m.HomeTeamID == m.AwayTeamID {
		return ErrSameTeam
	}
	if !m.Status.IsValid() {
		return ErrInvalidStatus
	}

	var count int64
	if err := s.db.Model(&struct{}{}).
		Table("teams").
		Where("id IN ? AND deleted_at IS NULL", []uint{m.HomeTeamID, m.AwayTeamID}).
		Count(&count).Error; err != nil {
		return err
	}
	if count != 2 {
		return ErrTeamNotFound
	}

	return nil
}
