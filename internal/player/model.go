package player

import (
	"time"

	"gorm.io/gorm"
)

type Position string

const (
	PositionStriker    Position = "STRIKER"
	PositionMidfielder Position = "MIDFIELDER"
	PositionDefender   Position = "DEFENDER"
	PositionGoalkeeper Position = "GOALKEEPER"
	PositionFlex       Position = "FLEX"
)

var ValidPositions = []Position{
	PositionStriker,
	PositionMidfielder,
	PositionDefender,
	PositionGoalkeeper,
	PositionFlex,
}

func (p Position) IsValid() bool {
	for _, v := range ValidPositions {
		if p == v {
			return true
		}
	}
	return false
}

type Player struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	TeamID       uint           `gorm:"not null;uniqueIndex:idx_team_jersey" json:"team_id"`
	Name         string         `gorm:"not null" json:"name"`
	HeightCm     float64        `json:"height_cm"`
	WeightKg     float64        `json:"weight_kg"`
	Position     Position       `gorm:"type:varchar(20);not null" json:"position"`
	JerseyNumber int            `gorm:"not null;uniqueIndex:idx_team_jersey" json:"jersey_number"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
