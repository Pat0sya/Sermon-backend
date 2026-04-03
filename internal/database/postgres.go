package database

import (
	"fmt"
	"log"

	"sermon-backend/internal/config"
	"sermon-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&models.Role{},
		&models.User{},
		&models.Server{},
		&models.Threshold{},
		&models.CurrentMetric{},
		&models.MetricHistory{},
		&models.Incident{},
		&models.IncidentComment{},
	)
	if err != nil {
		return nil, err
	}

	log.Println("database connected and migrated")
	return db, nil
}
