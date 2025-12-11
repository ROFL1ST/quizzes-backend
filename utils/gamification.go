package utils

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"time"
)

func UnlockAchievement(userID uint, achievementID uint) {

	var count int64
	config.DB.Model(&models.UserAchievement{}).
		Where("user_id = ? AND achievement_id = ?", userID, achievementID).
		Count(&count)

	if count > 0 {
		return
	}

	ua := models.UserAchievement{
		UserID:        userID,
		AchievementID: achievementID,
		UnlockedAt:    time.Now(),
	}
	config.DB.Create(&ua)

	var ach models.Achievement
	config.DB.First(&ach, achievementID)

	activity := models.Activity{
		UserID:      userID,
		Type:        "achievement",
		Description: "Membuka Achievement: " + ach.Name,
	}
	config.DB.Create(&activity)

	SendNotification(userID, "🏆 Achievement Unlocked: "+ach.Name, "/profile", "success")
}

func CheckQuizAchievements(userID uint, score int) {

	UnlockAchievement(userID, 1)

	if score == 100 {
		UnlockAchievement(userID, 2)
	}

	var totalKuis int64
	config.DB.Model(&models.History{}).Where("user_id = ?", userID).Count(&totalKuis)
	if totalKuis >= 10 {
		UnlockAchievement(userID, 3)
	}

	var user models.User

	config.DB.First(&user, userID)

	if user.Level >= 5 {
		UnlockAchievement(userID, 5)
	}

	if user.StreakCount >= 3 {
		UnlockAchievement(userID, 6)
	}

	var totalWins int64
	config.DB.Model(&models.Challenge{}).Where("winner_id = ?", userID).Count(&totalWins)

	if totalWins >= 1 {
		UnlockAchievement(userID, 4)
	}
}
