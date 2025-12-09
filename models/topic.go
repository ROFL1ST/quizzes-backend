package models

import "gorm.io/gorm"

type Topic struct {
	gorm.Model
	Slug        string `json:"slug" gorm:"unique;not null"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Quizzes     []Quiz `json:"-" gorm:"foreignKey:TopicID"`
}