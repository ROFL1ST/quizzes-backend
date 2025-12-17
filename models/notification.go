package models

import (
	"time"
	"gorm.io/gorm"
)

type Notification struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"index"` 
	Type      string         `json:"type"`  // e.g., "info", "warning", "success"             
	Title     string         `json:"title"`
	Message   string         `json:"message"`
	Link      string         `json:"link"`                 
	IsRead    bool           `json:"is_read" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}