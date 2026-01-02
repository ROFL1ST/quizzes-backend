package models

import "gorm.io/gorm"

type Announcement struct {
	gorm.Model
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatorID uint   `json:"creator_id"`
	Creator   Admin  `json:"creator" gorm:"foreignKey:CreatorID"`
	Type      string `json:"type" gorm:"default:'info'"`
	Active    bool   `json:"active" gorm:"default:true"`
}
