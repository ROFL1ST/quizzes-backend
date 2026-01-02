package models

import "gorm.io/gorm"

type Challenge struct {
	gorm.Model
	QuizID    *uint `json:"quiz_id" gorm:"default:null"`
	Quiz      Quiz  `json:"quiz" gorm:"foreignKey:QuizID"`
	CreatorID uint  `json:"creator_id"`
	Creator   User  `json:"creator" gorm:"foreignKey:CreatorID"`

	// Settings Baru
	Mode         string                 `json:"mode" gorm:"default:'1v1'"`
	TimeLimit    int                    `json:"time_limit"`
	IsRealtime   bool                   `json:"is_realtime" gorm:"default:false"`
	Status       string                 `json:"status" gorm:"default:'pending'"` // pending, active, finished
	Participants []ChallengeParticipant `json:"participants" gorm:"foreignKey:ChallengeID"`
	WagerAmount  int                    `json:"wager_amount" gorm:"default:0"`
	WinnerID     *uint                  `json:"winner_id"`    // Nullable (Pointer) karena bisa DRAW atau Team Win
	WinningTeam  string                 `json:"winning_team"` // "A", "B", atau "DRAW" (Khusus 2v2)
}

type ChallengeParticipant struct {
	gorm.Model
	ChallengeID uint   `json:"challenge_id"`
	UserID      uint   `json:"user_id"`
	User        User   `json:"user" gorm:"foreignKey:UserID"`
	Team        string `json:"team" gorm:"default:'solo'"`
	Status      string `json:"status" gorm:"default:'pending'"` // pending, accepted, rejected
	Score       int    `json:"score" gorm:"default:-1"`         // -1 artinya belum main
	TimeTaken   int    `json:"time_taken" gorm:"default:0"`
	IsFinished  bool   `json:"is_finished" gorm:"default:false"`
}
