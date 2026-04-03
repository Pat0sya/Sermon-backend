package models

import "gorm.io/gorm"

type Threshold struct {
	gorm.Model
	ServerID      uint   `gorm:"not null;index:idx_server_metric,unique"`
	Server        Server `gorm:"foreignKey:ServerID"`
	MetricType    string `gorm:"not null;index:idx_server_metric,unique"` // cpu, ram, disk
	WarningValue  int    `gorm:"not null"`
	CriticalValue int    `gorm:"not null"`
}
