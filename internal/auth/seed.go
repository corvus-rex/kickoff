package auth

import (
	"log"

	"gorm.io/gorm"
)

type seedUser struct {
	Name     string
	Email    string
	Password string
	Role     Role
}

// Seed creates one default user per role if the users table is currently empty.
func Seed(db *gorm.DB) error {
	var count int64
	if err := db.Model(&User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		log.Println("users already exist — skipping seed")
		return nil
	}

	defaults := []seedUser{
		{Name: "Default Admin", Email: "admin@xyz-football.local", Password: "ChangeMe123!", Role: RoleAdmin},
		{Name: "Default Manager", Email: "manager@xyz-football.local", Password: "ChangeMe123!", Role: RoleManager},
		{Name: "Default User", Email: "user@xyz-football.local", Password: "ChangeMe123!", Role: RoleUser},
	}

	for _, d := range defaults {
		hash, err := HashPassword(d.Password)
		if err != nil {
			return err
		}
		user := User{
			Name:         d.Name,
			Email:        normalizeEmail(d.Email),
			PasswordHash: hash,
			Role:         d.Role,
		}
		if err := db.Create(&user).Error; err != nil {
			return err
		}
	}

	log.Println("seeded default users — CHANGE THESE PASSWORDS before any non-local use:")
	for _, d := range defaults {
		log.Printf("  role=%-7s email=%-30s password=%s", d.Role, d.Email, d.Password)
	}

	return nil
}