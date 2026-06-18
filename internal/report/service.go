package report

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrMatchNotFound     = errors.New("match not found")
	ErrMatchNotFinished  = errors.New("match has not been finished yet")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetReport(matchID uint) (*MatchReport, error) {
	type matchRow struct {
		ID           uint
		MatchDate    time.Time
		MatchTime    string
		HomeTeamID   uint
		HomeTeamName string
		AwayTeamID   uint
		AwayTeamName string
		Status       string
	}

	var row matchRow
	err := s.db.Table("matches").
		Select(`matches.id, matches.match_date, matches.match_time,
			matches.home_team_id, ht.name AS home_team_name,
			matches.away_team_id, at.name AS away_team_name,
			matches.status`).
		Joins("JOIN teams ht ON ht.id = matches.home_team_id").
		Joins("JOIN teams at ON at.id = matches.away_team_id").
		Where("matches.id = ? AND matches.deleted_at IS NULL", matchID).
		Scan(&row).Error
	if err != nil || row.ID == 0 {
		return nil, ErrMatchNotFound
	}
	if row.Status != "FINISHED" {
		return nil, ErrMatchNotFinished
	}

	type goalRow struct {
		PlayerID   uint
		PlayerName string
		TeamID     uint
	}

	var goals []goalRow
	if err := s.db.Table("goals").
		Select("goals.player_id, p.name AS player_name, p.team_id").
		Joins("JOIN players p ON p.id = goals.player_id").
		Where("goals.match_id = ? AND goals.deleted_at IS NULL", matchID).
		Scan(&goals).Error; err != nil {
		return nil, err
	}

	homeScore := 0
	awayScore := 0
	scorerCounts := make(map[uint]int)
	scorerNames := make(map[uint]string)
	for _, g := range goals {
		if g.TeamID == row.HomeTeamID {
			homeScore++
		} else {
			awayScore++
		}
		scorerCounts[g.PlayerID]++
		scorerNames[g.PlayerID] = g.PlayerName
	}

	result := "DRAW"
	if homeScore > awayScore {
		result = "HOME_WIN"
	} else if awayScore > homeScore {
		result = "AWAY_WIN"
	}

	var topScorer *Scorer
	maxGoals := 0
	for pid, cnt := range scorerCounts {
		if cnt > maxGoals {
			maxGoals = cnt
			topScorer = &Scorer{
				PlayerID:   pid,
				PlayerName: scorerNames[pid],
				Goals:      cnt,
			}
		}
	}

	homeCumWins, err := s.countCumulativeWins(row.HomeTeamID, row.MatchDate, row.MatchTime)
	if err != nil {
		return nil, err
	}
	awayCumWins, err := s.countCumulativeWins(row.AwayTeamID, row.MatchDate, row.MatchTime)
	if err != nil {
		return nil, err
	}

	return &MatchReport{
		MatchID:            row.ID,
		MatchDate:          row.MatchDate.Format("2006-01-02"),
		MatchTime:          row.MatchTime,
		HomeTeam:           TeamInfo{ID: row.HomeTeamID, Name: row.HomeTeamName},
		AwayTeam:           TeamInfo{ID: row.AwayTeamID, Name: row.AwayTeamName},
		HomeScore:          homeScore,
		AwayScore:          awayScore,
		Result:             result,
		TopScorer:          topScorer,
		HomeCumulativeWins: homeCumWins,
		AwayCumulativeWins: awayCumWins,
	}, nil
}

func (s *Service) countCumulativeWins(teamID uint, upToDate time.Time, upToTime string) (int, error) {
	type matchInfo struct {
		ID         uint
		HomeTeamID uint
		AwayTeamID uint
	}

	var matches []matchInfo
	if err := s.db.Table("matches").
		Select("id, home_team_id, away_team_id").
		Where(`status = 'FINISHED' AND deleted_at IS NULL AND
			(match_date < ? OR (match_date = ? AND match_time <= ?))`,
			upToDate, upToDate, upToTime).
		Order("match_date ASC, match_time ASC").
		Scan(&matches).Error; err != nil {
		return 0, err
	}
	if len(matches) == 0 {
		return 0, nil
	}

	matchIDs := make([]uint, len(matches))
	for i, m := range matches {
		matchIDs[i] = m.ID
	}

	type goalInfo struct {
		MatchID uint
		TeamID  uint
	}

	var goals []goalInfo
	if err := s.db.Table("goals").
		Select("goals.match_id, players.team_id").
		Joins("JOIN players ON players.id = goals.player_id").
		Where("goals.match_id IN ? AND goals.deleted_at IS NULL", matchIDs).
		Scan(&goals).Error; err != nil {
		return 0, err
	}

	matchGoals := make(map[uint]map[uint]int)
	for _, g := range goals {
		if matchGoals[g.MatchID] == nil {
			matchGoals[g.MatchID] = make(map[uint]int)
		}
		matchGoals[g.MatchID][g.TeamID]++
	}

	wins := 0
	for _, m := range matches {
		homeG := matchGoals[m.ID][m.HomeTeamID]
		awayG := matchGoals[m.ID][m.AwayTeamID]

		if teamID == m.HomeTeamID && homeG > awayG {
			wins++
		} else if teamID == m.AwayTeamID && awayG > homeG {
			wins++
		}
	}

	return wins, nil
}
