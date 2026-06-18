package goal

import (
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByMatch(matchID uint) ([]Goal, error) {
	var goals []Goal
	err := r.db.Where("match_id = ?", matchID).Find(&goals).Error
	return goals, err
}

func (r *Repository) FindByID(id uint) (*Goal, error) {
	var goal Goal
	err := r.db.First(&goal, id).Error
	if err != nil {
		return nil, err
	}
	return &goal, nil
}

func (r *Repository) Create(goal *Goal) error {
	return r.db.Create(goal).Error
}

func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&Goal{}, id).Error
}
