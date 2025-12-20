package models

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Question struct {
	gorm.Model
	QuizID         uint           `json:"quiz_id"`
	Quiz           Quiz           `json:"-" gorm:"foreignKey:QuizID"`
	QuestionText   string         `json:"question"`
	Options        pq.StringArray `json:"options" gorm:"type:text[]"`
	CorrectAnswer  string         `json:"correct"`
	Hint           string         `json:"hint"`
	Type           string         `json:"type" gorm:"default:'mcq'"`
	CorrectCount   int            `json:"correct_count" gorm:"default:0"`
	IncorrectCount int            `json:"incorrect_count" gorm:"default:0"`
}

type QuestionAnalysis struct {
	ID             uint   `json:"id"`
	QuestionText   string `json:"question_text"`
	CorrectCount   int    `json:"correct_count"`
	IncorrectCount int    `json:"incorrect_count"`
	TotalAttempts  int    `json:"total_attempts"`
	Difficulty     string `json:"difficulty"`
	AccuracyRate   string `json:"accuracy_rate"`
}
