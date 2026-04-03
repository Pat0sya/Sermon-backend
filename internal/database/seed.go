package database

import (
	"errors"
	"log"
	"os"

	"sermon-backend/internal/models"
	"sermon-backend/internal/utils"

	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	var adminRole models.Role
	if err := db.FirstOrCreate(&adminRole, models.Role{Name: "admin"}).Error; err != nil {
		return err
	}

	var operatorRole models.Role
	if err := db.FirstOrCreate(&operatorRole, models.Role{Name: "operator"}).Error; err != nil {
		return err
	}

	adminPassword, err := utils.HashPassword("admin")
	if err != nil {
		return err
	}

	operatorPassword, err := utils.HashPassword("operator123")
	if err != nil {
		return err
	}

	var admin models.User
	err = db.Where("username = ?", "admin").First(&admin).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		admin = models.User{
			Username:           "admin",
			PasswordHash:       adminPassword,
			RoleID:             adminRole.ID,
			MustChangePassword: true,
			IsActive:           true,
		}
		if err := db.Create(&admin).Error; err != nil {
			return err
		}
	}

	var operator models.User
	err = db.Where("username = ?", "operator").First(&operator).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		operator = models.User{
			Username:           "operator",
			PasswordHash:       operatorPassword,
			RoleID:             operatorRole.ID,
			MustChangePassword: true,
			IsActive:           true,
		}
		if err := db.Create(&operator).Error; err != nil {
			return err
		}
	}

	if os.Getenv("APP_ENV") == "dev" {
		var server models.Server
		err = db.Where("host = ?", "127.0.0.1").First(&server).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			server = models.Server{
				Name:        "test-server-01",
				Host:        "127.0.0.1",
				OS:          "Ubuntu 22.04",
				Description: "Тестовый сервер для разработки",
				IsActive:    true,
				Status:      "normal",
			}
			if err := db.Create(&server).Error; err != nil {
				return err
			}
		}

		thresholds := []models.Threshold{
			{ServerID: server.ID, MetricType: "cpu", WarningValue: 80, CriticalValue: 90},
			{ServerID: server.ID, MetricType: "ram", WarningValue: 80, CriticalValue: 90},
			{ServerID: server.ID, MetricType: "disk", WarningValue: 85, CriticalValue: 95},
		}

		for _, t := range thresholds {
			var existing models.Threshold
			err := db.Where("server_id = ? AND metric_type = ?", t.ServerID, t.MetricType).First(&existing).Error
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				if err := db.Create(&t).Error; err != nil {
					return err
				}
			}
		}
	}

	log.Println("seed completed")
	return nil
}
