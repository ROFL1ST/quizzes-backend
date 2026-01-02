package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type History struct {
	gorm.Model
	UserID       uint           `json:"user_id"`
	User         User           `json:"user" gorm:"foreignKey:UserID"`
	QuizID       uint           `json:"quiz_id"`
	QuizTitle    string         `json:"quiz_title"`
	Score        int            `json:"score"`
	TotalSoal    int            `json:"total_soal"`
	TimeTaken    int            `json:"time_taken"`
	Snapshot     datatypes.JSON `json:"snapshot"`
	AssignmentID *uint          `json:"assignment_id,omitempty"`
	ClassroomID  *uint          `json:"classroom_id,omitempty"`
}
