package models

import "gorm.io/gorm"

type QuizReview struct {
	gorm.Model
	UserID  uint   `json:"user_id"`
	User    User   `json:"user" gorm:"foreignKey:UserID"`
	QuizID  uint   `json:"quiz_id"`
	Quiz    Quiz   `json:"quiz" gorm:"foreignKey:QuizID"`
	Rating  int    `json:"rating"` // 1-5
	Comment string `json:"comment"`
}
