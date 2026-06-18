package match

import (
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindAll() ([]Match, error) {
	var matches []Match
	err := r.db.Find(&matches).Error
	return matches, err
}

func (r *Repository) FindByID(id uint) (*Match, error) {
	var match Match
	err := r.db.First(&match, id).Error
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *Repository) Create(match *Match) error {
	return r.db.Create(match).Error
}

func (r *Repository) Update(match *Match) error {
	return r.db.Save(match).Error
}

func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&Match{}, id).Error
}
