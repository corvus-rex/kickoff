package goal

import (
	"time"

	"gorm.io/gorm"
)

type Goal struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	MatchID    uint           `gorm:"not null;index" json:"match_id"`
	PlayerID   uint           `gorm:"not null" json:"player_id"`
	GoalMinute int            `gorm:"not null" json:"goal_minute"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
