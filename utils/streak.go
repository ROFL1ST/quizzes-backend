package utils

import (
	"github.com/ROFL1ST/quizzes-backend/models"
	"time"
)

func UpdateStreak(user *models.User) {
	now := time.Now()

	if user.LastActivityDate == nil {
		user.StreakCount = 1
		user.LastActivityDate = &now
		return
	}

	last := *user.LastActivityDate

	y1, m1, d1 := last.Date()
	y2, m2, d2 := now.Date()

	dateLast := time.Date(y1, m1, d1, 0, 0, 0, 0, time.Local)
	dateNow := time.Date(y2, m2, d2, 0, 0, 0, 0, time.Local)

	daysDiff := int(dateNow.Sub(dateLast).Hours() / 24)

	if daysDiff == 1 {

		user.StreakCount++
	} else if daysDiff > 1 {

		user.StreakCount = 1
	}

	user.LastActivityDate = &now
}
