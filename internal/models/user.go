package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username           string `gorm:"uniqueIndex;not null"`
	PasswordHash       string `gorm:"not null"`
	RoleID             uint   `gorm:"not null"`
	Role               Role
	MustChangePassword bool `gorm:"not null;default:false"`
	IsActive           bool `gorm:"not null;default:true"`
}
