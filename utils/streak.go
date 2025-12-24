package utils

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"time"
)

func DaysBetween(lastDate time.Time, nowDate time.Time) int {
	d1 := StripTime(lastDate)
	d2 := StripTime(nowDate)

	hours := d2.Sub(d1).Hours()
	return int(hours / 24)
}

func RecordActivity(userID uint) {
	now := GetJakartaTime()
	today := StripTime(now)

	var exists int64
	config.DB.Model(&models.StreakLog{}).
		Where("user_id = ? AND date = ?", userID, today).
		Count(&exists)

	if exists > 0 {
		return
	}

	logEntry := models.StreakLog{
		UserID: userID,
		Date:   today,
	}
	if err := config.DB.Create(&logEntry).Error; err != nil {
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return
	}

	yesterday := today.AddDate(0, 0, -1)
	var yesterdayExists int64
	config.DB.Model(&models.StreakLog{}).
		Where("user_id = ? AND date = ?", userID, yesterday).
		Count(&yesterdayExists)

	if yesterdayExists > 0 {

		user.StreakCount++
	} else {

		user.StreakCount = 1
	}

	user.LastActivityDate = &now

	config.DB.Save(&user)
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
