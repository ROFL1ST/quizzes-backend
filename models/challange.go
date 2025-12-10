package models

import "gorm.io/gorm"

type Challenge struct {
	gorm.Model
	ChallengerID    uint   `json:"challenger_id"`
	OpponentID      uint   `json:"opponent_id"`
	QuizID          uint   `json:"quiz_id"`
	Challenger      User   `json:"challenger" gorm:"foreignKey:ChallengerID"`
	Opponent        User   `json:"opponent" gorm:"foreignKey:OpponentID"`
	Quiz            Quiz   `json:"quiz" gorm:"foreignKey:QuizID"`
	ChallengerScore int    `json:"challenger_score" gorm:"default:-1"`
	OpponentScore   int    `json:"opponent_score" gorm:"default:-1"`
	Status          string `json:"status" gorm:"default:'pending'"`
	WinnerID        *uint  `json:"winner_id"`
}