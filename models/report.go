package models

import "gorm.io/gorm"

type Report struct {
	gorm.Model
	ReporterID uint   `json:"reporter_id"`
	Reporter   User   `json:"reporter" gorm:"foreignKey:ReporterID"`
	TargetID   uint   `json:"target_id"`   // ID of the question, user, or quiz being reported
	TargetType string `json:"target_type"` // "question", "user", "quiz"
	Reason     string `json:"reason"`
	Status     string `json:"status" gorm:"default:'pending'"` // "pending", "resolved", "dismissed"
}
