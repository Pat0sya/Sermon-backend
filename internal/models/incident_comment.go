package models

import "time"

type IncidentComment struct {
	ID         uint     `gorm:"primaryKey"`
	IncidentID uint     `gorm:"not null;index"`
	Incident   Incident `gorm:"foreignKey:IncidentID"`
	UserID     uint     `gorm:"not null"`
	User       User     `gorm:"foreignKey:UserID"`
	Comment    string   `gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
