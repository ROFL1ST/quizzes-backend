package models

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Question struct {
	gorm.Model
	QuizID        uint           `json:"quiz_id"`
	Quiz          Quiz           `json:"-" gorm:"foreignKey:QuizID"`
	QuestionText  string         `json:"question"`
	Options       pq.StringArray `json:"options" gorm:"type:text[]"`
	CorrectAnswer string         `json:"correct"`
	Hint          string         `json:"hint"`
}