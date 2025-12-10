package models

import (
	"time"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string `json:"name"`
	Username string `json:"username" gorm:"unique;not null"`
	Password string `json:"-"`
	XP               int64     `json:"xp" gorm:"default:0"`
	Level            int       `json:"level" gorm:"default:1"`
	StreakCount      int       `json:"streak_count" gorm:"default:0"`
	LastActivityDate *time.Time `json:"last_activity_date"` 
}