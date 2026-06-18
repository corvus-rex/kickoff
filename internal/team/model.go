package team

import (
	"time"

	"gorm.io/gorm"
)

type Team struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	Name                string         `gorm:"uniqueIndex;not null" json:"name"`
	LogoURL             string         `json:"logo_url"`
	FoundedYear         int            `gorm:"not null" json:"founded_year"`
	HeadquartersAddress string         `json:"headquarters_address"`
	HeadquartersCity    string         `json:"headquarters_city"`
	ManagerUserID       *uint          `json:"manager_user_id"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}
