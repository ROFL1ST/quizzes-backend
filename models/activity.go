package models

import "gorm.io/gorm"

type Activity struct {
	gorm.Model
	UserID      uint   `json:"user_id"`
	User        User   `json:"user" gorm:"foreignKey:UserID"`
	Type        string `json:"type"` 
	Description string `json:"description"`
}