package models

import (
	"time"
)

type UserAdaptivity struct {
	UserID         uint      `json:"user_id" gorm:"primaryKey;autoIncrement:false"`
	AdaptiveRating float64   `json:"adaptive_rating" gorm:"default:0.5"` // 0.0 - 1.0 (Theta / Ability)
	Confidence     float64   `json:"confidence" gorm:"default:0.0"`      // Seberapa yakin AI dengan rating ini (Optional)
	LastDiff       float64   `json:"last_diff" gorm:"default:0.5"`       // Kesulitan soal terakhir yang dikerjakan
	UpdatedAt      time.Time `json:"updated_at"`
}
