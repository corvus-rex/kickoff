package team

import (
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindAll() ([]Team, error) {
	var teams []Team
	err := r.db.Find(&teams).Error
	return teams, err
}

func (r *Repository) FindByID(id uint) (*Team, error) {
	var team Team
	err := r.db.First(&team, id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *Repository) Create(team *Team) error {
	return r.db.Create(team).Error
}

func (r *Repository) Update(team *Team) error {
	return r.db.Save(team).Error
}

func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&Team{}, id).Error
}
