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

	// Bandingkan tanggal terakhir main dengan hari ini (WIB)
	diff := DaysBetween(*user.LastActivityDate, now)

	if diff == 0 {
		// Masih hari yang sama: Update jam saja, streak tetap
		user.LastActivityDate = &now
	} else if diff == 1 {
		// Kemarin main, hari ini main: Streak NAIK
		user.StreakCount++
		user.LastActivityDate = &now
	} else {
		// Bolos lebih dari 1 hari: Streak RESET
		user.StreakCount = 1
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
		// Sudah klaim hari ini (seharusnya dicek di controller juga)
		user.LastClaimDate = &now
	} else if diff == 1 {
		// Kemarin klaim, hari ini klaim: Streak NAIK
		user.LoginStreak++
		user.LastClaimDate = &now
	} else {
		// Putus login: Streak RESET
		user.LoginStreak = 1
		user.LastClaimDate = &now
	}
}