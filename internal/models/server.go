package models

import "gorm.io/gorm"

type Server struct {
	gorm.Model
	Name           string `gorm:"not null"`
	Host           string `gorm:"not null;uniqueIndex"`
	OS             string `gorm:"not null"`
	Description    string
	IsActive       bool   `gorm:"not null;default:true"`
	Status         string `gorm:"not null;default:normal"`
	LastSeenAt     *int64
	AgentTokenHash string
}
