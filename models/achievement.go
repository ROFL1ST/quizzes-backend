package models

import (
	"time"
)

type Achievement struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name"`
	Description string `json:"description"` 
	IconURL     string `json:"icon_url"`
}

type UserAchievement struct {
	UserID        uint      `json:"user_id" gorm:"primaryKey"`
	AchievementID uint      `json:"achievement_id" gorm:"primaryKey"`
	UnlockedAt    time.Time `json:"unlocked_at"`
}