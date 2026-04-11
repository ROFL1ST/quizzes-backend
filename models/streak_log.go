package models

import (
	"gorm.io/gorm"
	"time"
)

type StreakLog struct {
	gorm.Model
	UserID    uint      `json:"user_id" gorm:"index"`
	Date      time.Time `json:"date" gorm:"type:date;index"`
}
