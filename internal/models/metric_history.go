package models

import "time"

type MetricHistory struct {
	ID          uint      `gorm:"primaryKey"`
	ServerID    uint      `gorm:"not null;index"`
	Server      Server    `gorm:"foreignKey:ServerID"`
	MetricType  string    `gorm:"not null;index"` // cpu, ram, disk
	MetricValue float64   `gorm:"not null"`
	CollectedAt time.Time `gorm:"not null;index"`
	CreatedAt   time.Time
}
