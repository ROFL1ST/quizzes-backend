package models

import (
	"time"
	"gorm.io/gorm"
)

type StreakLog struct {
	gorm.Model
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"index"`
	Date      time.Time `json:"date" gorm:"type:date;index"` 
	CreatedAt time.Time `json:"created_at"`
}

