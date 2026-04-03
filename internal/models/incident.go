package models

import "time"

type Incident struct {
	ID             uint      `gorm:"primaryKey"`
	ServerID       uint      `gorm:"not null;index"`
	Server         Server    `gorm:"foreignKey:ServerID"`
	MetricType     string    `gorm:"not null;index"` // cpu, ram, disk
	Status         string    `gorm:"not null;index"` // open, in_progress, closed
	ThresholdValue float64   `gorm:"not null"`
	ActualValue    float64   `gorm:"not null"`
	Message        string    `gorm:"not null"`
	StartedAt      time.Time `gorm:"not null"`
	ClosedAt       *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
