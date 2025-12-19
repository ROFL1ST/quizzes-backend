package models

import (
	"time"
)

type DailyRewardConfig struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	Day    int  `json:"day" gorm:"unique"`
	Reward int  `json:"reward"`
}

type DailyClaim struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserID uint `json:"user_id" gorm:"index"`

	RewardType string `json:"reward_type" gorm:"index"`

	ClaimedDate time.Time `json:"claimed_date"`

	CreatedAt time.Time `json:"created_at"`
}

type Mission struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Key         string `json:"key" gorm:"unique"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Target      int    `json:"target"`
	Reward      int    `json:"reward"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
}


type UserMission struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `json:"user_id" gorm:"index"`
	MissionID   uint      `json:"mission_id" gorm:"index"`
	Mission     Mission   `json:"mission" gorm:"foreignKey:MissionID"`
	Progress    int       `json:"progress" gorm:"default:0"`
	IsClaimed   bool      `json:"is_claimed" gorm:"default:false"`
	ResetDate   time.Time `json:"reset_date"`
	UpdatedAt   time.Time `json:"updated_at"`
}