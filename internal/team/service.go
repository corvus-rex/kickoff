package team

import (
	"errors"

	"gorm.io/gorm"

	"kickoff/internal/auth"
)

var (
	ErrForbidden    = errors.New("access denied")
	ErrNotFound     = errors.New("team not found")
	ErrNameRequired = errors.New("team name is required")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List() ([]Team, error) {
	return s.repo.FindAll()
}

func (s *Service) GetByID(id uint) (*Team, error) {
	team, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return team, nil
}

func (s *Service) Create(team *Team, role auth.Role) error {
	if role != auth.RoleAdmin {
		return ErrForbidden
	}
	if team.Name == "" {
		return ErrNameRequired
	}
	return s.repo.Create(team)
}

func (s *Service) Update(updates *Team, userID uint, role auth.Role) (*Team, error) {
	existing, err := s.repo.FindByID(updates.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if role != auth.RoleAdmin {
		if existing.ManagerUserID == nil || *existing.ManagerUserID != userID {
			return nil, ErrForbidden
		}
	}

	if updates.Name == "" {
		return nil, ErrNameRequired
	}

	existing.Name = updates.Name
	existing.LogoURL = updates.LogoURL
	existing.FoundedYear = updates.FoundedYear
	existing.HeadquartersAddress = updates.HeadquartersAddress
	existing.HeadquartersCity = updates.HeadquartersCity
	existing.ManagerUserID = updates.ManagerUserID

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
