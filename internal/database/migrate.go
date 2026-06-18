package database

import (
	"log"

	"gorm.io/gorm"
)

var registeredModels []interface{}

func RegisterModel(models ...interface{}) {
	registeredModels = append(registeredModels, models...)
}

func RunMigrations(db *gorm.DB) error {
	if len(registeredModels) == 0 {
		log.Println("no models registered for migration — skipping")
		return nil
	}
	log.Printf("running migrations for %d registered model(s)", len(registeredModels))
	return db.AutoMigrate(registeredModels...)
}