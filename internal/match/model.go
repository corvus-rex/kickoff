package match

import (
	"time"

	"gorm.io/gorm"
)

type MatchStatus string

const (
	StatusScheduled MatchStatus = "SCHEDULED"
	StatusFinished  MatchStatus = "FINISHED"
)

func (s MatchStatus) IsValid() bool {
	return s == StatusScheduled || s == StatusFinished
}

type Match struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	MatchDate  time.Time      `gorm:"type:date;not null" json:"match_date"`
	MatchTime  string         `gorm:"type:varchar(5);not null" json:"match_time"`
	HomeTeamID uint           `gorm:"not null" json:"home_team_id"`
	AwayTeamID uint           `gorm:"not null" json:"away_team_id"`
	Status     MatchStatus    `gorm:"type:varchar(20);not null;default:SCHEDULED" json:"status"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
