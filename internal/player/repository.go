package player

import (
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByTeam(teamID uint) ([]Player, error) {
	var players []Player
	err := r.db.Where("team_id = ?", teamID).Find(&players).Error
	return players, err
}

func (r *Repository) FindByID(id uint) (*Player, error) {
	var player Player
	err := r.db.First(&player, id).Error
	if err != nil {
		return nil, err
	}
	return &player, nil
}

func (r *Repository) Create(player *Player) error {
	return r.db.Create(player).Error
}

func (r *Repository) Update(player *Player) error {
	return r.db.Save(player).Error
}

func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&Player{}, id).Error
}

func (r *Repository) IsJerseyNumberTaken(teamID uint, jerseyNumber int, excludeID *uint) (bool, error) {
	query := r.db.Model(&Player{}).Where("team_id = ? AND jersey_number = ?", teamID, jerseyNumber)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}
