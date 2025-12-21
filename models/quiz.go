package models

import "gorm.io/gorm"

type Quiz struct {
	gorm.Model
	TopicID     uint       `json:"topic_id"`
	Topic       Topic      `json:"-" gorm:"foreignKey:TopicID"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Active      bool       `json:"active" gorm:"default:false"`
	CreatorID   *uint      `json:"creator_id"` // Null = Admin, Isi = User
	Creator     User       `json:"-" gorm:"foreignKey:CreatorID"`
	IsPublic    bool       `json:"is_public" gorm:"default:false"` // Muncul di pencarian?
	Status      string     `json:"status" gorm:"default:'draft'"` // draft, published, archived
	Questions   []Question `json:"-" gorm:"foreignKey:QuizID"`
}
