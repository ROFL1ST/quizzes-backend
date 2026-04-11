package models

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

const (
	QuestionTypeMCQ         = "mcq"
	QuestionTypeEssay       = "essay"
	QuestionTypeShortAnswer = "short_answer"
	QuestionTypeBoolean     = "boolean"
	QuestionTypeMultiSelect = "multi_select"
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
	Difficulty     float64        `json:"difficulty" gorm:"default:0.5"` // 0.0 to 1.0
}


