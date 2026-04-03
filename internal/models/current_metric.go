package models

import "time"

type CurrentMetric struct {
	ID          uint      `gorm:"primaryKey"`
	ServerID    uint      `gorm:"not null;uniqueIndex"`
	Server      Server    `gorm:"foreignKey:ServerID"`
	CPUUsage    float64   `gorm:"not null"`
	RAMUsage    float64   `gorm:"not null"`
	DiskUsage   float64   `gorm:"not null"`
	CollectedAt time.Time `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
