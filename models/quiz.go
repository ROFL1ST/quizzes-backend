package models

import "gorm.io/gorm"

type Quiz struct {
	gorm.Model
	TopicID     uint       `json:"topic_id"`
	Topic       Topic      `json:"-" gorm:"foreignKey:TopicID"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Questions   []Question `json:"-" gorm:"foreignKey:QuizID"`
}