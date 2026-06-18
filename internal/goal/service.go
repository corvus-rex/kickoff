package goal

import (
	"errors"

	"gorm.io/gorm"

	"kickoff/internal/auth"
)

var (
	ErrForbidden          = errors.New("access denied")
	ErrNotFound           = errors.New("goal not found")
	ErrMatchNotFound      = errors.New("match not found")
	ErrPlayerNotFound     = errors.New("player not found")
	ErrPlayerNotInMatch   = errors.New("player does not belong to either team in this match")
	ErrInvalidMinute      = errors.New("goal minute must be greater than 0")
)

type Service struct {
	repo *Repository
	db   *gorm.DB
}

func NewService(repo *Repository, db *gorm.DB) *Service {
	return &Service{repo: repo, db: db}
}

func (s *Service) ListByMatch(matchID uint) ([]Goal, error) {
	return s.repo.FindByMatch(matchID)
}

func (s *Service) Create(goal *Goal, role auth.Role) error {
	if role != auth.RoleAdmin {
		return ErrForbidden
	}
	if err := s.validateGoal(goal.MatchID, goal.PlayerID, goal.GoalMinute); err != nil {
		return err
	}
	return s.repo.Create(goal)
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

func (s *Service) validateGoal(matchID, playerID uint, goalMinute int) error {
	if goalMinute <= 0 {
		return ErrInvalidMinute
	}

	var matchCount int64
	if err := s.db.Table("matches").
		Where("id = ? AND deleted_at IS NULL", matchID).
		Count(&matchCount).Error; err != nil {
		return err
	}
	if matchCount == 0 {
		return ErrMatchNotFound
	}

	var homeTeamID, awayTeamID uint
	row := s.db.Table("matches").
		Select("home_team_id, away_team_id").
		Where("id = ? AND deleted_at IS NULL", matchID).
		Row()
	if err := row.Scan(&homeTeamID, &awayTeamID); err != nil {
		return ErrMatchNotFound
	}

	var playerTeamID uint
	row2 := s.db.Table("players").
		Select("team_id").
		Where("id = ? AND deleted_at IS NULL", playerID).
		Row()
	if err := row2.Scan(&playerTeamID); err != nil {
		return ErrPlayerNotFound
	}

	if playerTeamID != homeTeamID && playerTeamID != awayTeamID {
		return ErrPlayerNotInMatch
	}

	return nil
}
