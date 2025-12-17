package models

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	Name                   string     `json:"name"`
	Username               string     `json:"username" gorm:"unique;not null"`
	Email                  string     `json:"email" gorm:"unique;default:null"`
	IsEmailVerified        bool       `json:"is_email_verified" gorm:"default:false"`
	EmailVerificationToken string     `json:"-" gorm:"default:null"`
	Password               string     `json:"-"`
	XP                     int64      `json:"xp" gorm:"default:0"`
	Level                  int        `json:"level" gorm:"default:1"`
	StreakCount            int        `json:"streak_count" gorm:"default:0"`
	LastActivityDate       *time.Time `json:"last_activity_date"`
}

type PasswordReset struct {
	gorm.Model
	Email     string    `json:"email" gorm:"index;not null"`
	Token     string    `json:"token" gorm:"unique;not null"`
	ExpiredAt time.Time `json:"expired_at"`
}