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
	if err := config.DB.Create(&ua).Error; err != nil {
		return
	}

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

	if score < 50 {
		UnlockAchievement(userID, 11)
	}

	var totalKuis int64
	config.DB.Model(&models.History{}).Where("user_id = ?", userID).Count(&totalKuis)
	if totalKuis >= 10 {
		UnlockAchievement(userID, 3)
	}

	currentHour := time.Now().Hour()
	if currentHour >= 0 && currentHour < 5 {
		UnlockAchievement(userID, 12)
	}

	var user models.User
	config.DB.First(&user, userID)

	if user.Level >= 5 {
		UnlockAchievement(userID, 5)
	}

	if user.Level >= 10 {
		UnlockAchievement(userID, 13)
	}

	if user.XP >= 5000 {
		UnlockAchievement(userID, 10)
	}

	if user.StreakCount >= 3 {
		UnlockAchievement(userID, 6)
	}

	hoursSinceJoin := time.Since(user.CreatedAt).Hours()
	if hoursSinceJoin >= (24 * 30) {
		UnlockAchievement(userID, 14)
	}

	if score == 100 {
		var perfectQuizzes int64
		config.DB.Model(&models.History{}).
			Where("user_id = ? AND score = 100", userID).
			Distinct("quiz_id").
			Count(&perfectQuizzes)

		if perfectQuizzes >= 3 {
			UnlockAchievement(userID, 15)
		}
	}

	var distinctTopics int64
	config.DB.Table("histories").
		Joins("JOIN quizzes ON quizzes.id = histories.quiz_id").
		Where("histories.user_id = ?", userID).
		Distinct("quizzes.topic_id").
		Count(&distinctTopics)

	if distinctTopics >= 3 {
		UnlockAchievement(userID, 8)
	}

	var totalWins int64
	config.DB.Model(&models.Challenge{}).Where("winner_id = ?", userID).Count(&totalWins)

	if totalWins >= 1 {
		UnlockAchievement(userID, 4)
	}

	if totalWins >= 5 {
		UnlockAchievement(userID, 9)
	}
}