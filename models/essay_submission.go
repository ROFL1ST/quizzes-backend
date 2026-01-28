package models

import (
	"gorm.io/gorm"
)

type EssaySubmission struct {
	gorm.Model
	HistoryID  uint     `json:"history_id"`
	QuestionID uint     `json:"question_id"`
	Question   Question `json:"question" gorm:"foreignKey:QuestionID"`
	UserAnswer string   `json:"user_answer"`
	TeacherKey string   `json:"teacher_key"` // Snapshot of key
	AIScore    float64  `json:"ai_score"`    // 0.0 - 100.0
	AIFeedback string   `json:"ai_feedback"`
	IsGraded   bool     `json:"is_graded" gorm:"default:false"`
}
