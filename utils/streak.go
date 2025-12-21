package utils

import (
	"github.com/ROFL1ST/quizzes-backend/models"
	"time"
)

func DaysBetween(lastDate time.Time, nowDate time.Time) int {
	d1 := StripTime(lastDate)
	d2 := StripTime(nowDate)

	// Hitung durasi jam dibagi 24
	hours := d2.Sub(d1).Hours()
	return int(hours / 24)
}

func UpdateQuizStreak(user *models.User) {
	now := GetJakartaTime()

	if user.LastActivityDate == nil {
		user.StreakCount = 1
		user.LastActivityDate = &now
		return
	}

	diff := DaysBetween(*user.LastActivityDate, now)

	if diff == 0 {
		user.LastActivityDate = &now
	} else if diff == 1 {
		user.StreakCount++
		user.LastActivityDate = &now
	} else {
		user.StreakCount = 1 // Reset jika bolos
		user.LastActivityDate = &now
	}
}

func UpdateLoginStreak(user *models.User) {
	now := GetJakartaTime()

	if user.LastClaimDate == nil {
		user.LoginStreak = 1
		user.LastClaimDate = &now
		return
	}

	diff := DaysBetween(*user.LastClaimDate, now)

	if diff == 0 {

		user.LastClaimDate = &now
	} else {

		user.LoginStreak++
		user.LastClaimDate = &now
	}
}
