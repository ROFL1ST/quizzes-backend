package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"math"
)

// Struct untuk Response Profile yang rapi
type ProfileStats struct {
	TotalQuizzes   int64   `json:"total_quizzes"`
	AverageScore   float64 `json:"average_score"`
	TotalWins      int64   `json:"total_wins"`
	FavoriteTopic  string  `json:"favorite_topic"`
	CompletionRate string  `json:"completion_rate"` // % Soal benar dari seluruh soal yang dijawab
}

// Struct untuk Radar Chart (Score per Topic)
type TopicPerformance struct {
	TopicName string  `json:"topic_name"`
	AvgScore  float64 `json:"avg_score"`
}

func GetMyProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)

	// 1. Ambil Data User Basic
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	// 2. Hitung Statistik Dasar (Total Kuis & Rata-rata)
	var totalQuizzes int64
	var avgScore float64
	config.DB.Model(&models.History{}).Where("user_id = ?", userID).Count(&totalQuizzes)

	// Gunakan COALESCE agar tidak error jika null (belum pernah main)
	config.DB.Model(&models.History{}).Where("user_id = ?", userID).
		Select("COALESCE(AVG(score), 0)").Scan(&avgScore)

	// 3. Hitung Kemenangan (Dari tabel Challenge)
	var totalWins int64
	config.DB.Model(&models.Challenge{}).
		Where("winner_id = ?", userID).Count(&totalWins)

	// 4. Cari Topik Terfavorit (Paling sering dimainkan)
	var favTopic struct {
		Title string
		Count int
	}
	config.DB.Table("histories").
		Select("topics.title, count(histories.id) as count").
		Joins("JOIN quizzes ON quizzes.id = histories.quiz_id").
		Joins("JOIN topics ON topics.id = quizzes.topic_id").
		Where("histories.user_id = ?", userID).
		Group("topics.title").
		Order("count desc").
		Limit(1).
		Scan(&favTopic)

	if favTopic.Title == "" {
		favTopic.Title = "-"
	}

	// 5. Analitik Per Topik (Untuk Radar Chart)
	var topicPerfs []TopicPerformance
	config.DB.Table("histories").
		Select("topics.title as topic_name, AVG(histories.score) as avg_score").
		Joins("JOIN quizzes ON quizzes.id = histories.quiz_id").
		Joins("JOIN topics ON topics.id = quizzes.topic_id").
		Where("histories.user_id = ?", userID).
		Group("topics.title").
		Scan(&topicPerfs)

	currentLevel := user.Level
	nextLevel := currentLevel + 1

	// Hitung batas bawah level saat ini dan batas level selanjutnya
	currentLevelBaseXP := utils.CalculateMinXPForLevel(currentLevel)
	nextLevelThreshold := utils.CalculateMinXPForLevel(nextLevel)

	levelProgress := fiber.Map{
		"current_level":    currentLevel,
		"current_xp":       user.XP,
		"level_base_xp":    currentLevelBaseXP,           // XP awal 
		"next_level_xp":    nextLevelThreshold,           // XP target 
		"xp_needed":        nextLevelThreshold - user.XP, // Sisa XP yg dibutuhkan
		"progress_percent": 0,
	}


	rangeXP := nextLevelThreshold - currentLevelBaseXP
	if rangeXP > 0 {
		progress := float64(user.XP-currentLevelBaseXP) / float64(rangeXP) * 100
		levelProgress["progress_percent"] = math.Round(progress*10) / 10 // bulatkan 1 desimal
	} else {
		levelProgress["progress_percent"] = 100 // Max level case
	}
	// Susun Response
	response := fiber.Map{
		"user": user,
		"stats": ProfileStats{
			TotalQuizzes:  totalQuizzes,
			AverageScore:  avgScore,
			TotalWins:     totalWins,
			FavoriteTopic: favTopic.Title,
		},
		"topic_performance": topicPerfs,
		"level_progress":    levelProgress,
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Profile retrieved", response)
}

func UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)

	var input struct {
		Name     string `json:"name"`
		Username string `json:"username"`
		Password string `json:"password"` // Opsional
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	// Update field jika dikirim
	if input.Name != "" {
		user.Name = input.Name
	}

	// Cek unique username jika diganti
	if input.Username != "" && input.Username != user.Username {
		var check models.User
		if err := config.DB.Where("username = ?", input.Username).First(&check).Error; err == nil {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Username already taken", nil)
		}
		user.Username = input.Username
	}

	// Update password jika diisi (hash ulang)
	if input.Password != "" {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
		user.Password = string(hashed)
	}

	if err := config.DB.Save(&user).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed update profile", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Profile updated", user)
}

// Fitur Tambahan: Lihat Profile Teman/Orang Lain
func GetUserProfile(c *fiber.Ctx) error {
	username := c.Params("username")

	var user models.User
	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	// Untuk profil orang lain, kita hanya tampilkan statistik umum (tanpa detail privat)
	var totalQuizzes int64
	config.DB.Model(&models.History{}).Where("user_id = ?", user.ID).Count(&totalQuizzes)

	stats := fiber.Map{
		"xp":            user.XP,
		"level":         user.Level,
		"total_quizzes": totalQuizzes,
		"joined_at":     user.CreatedAt,
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "User profile retrieved", fiber.Map{
		"name":     user.Name,
		"username": user.Username,
		"stats":    stats,
	})
}

func GetMyAchievements(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)

	// 1. Ambil semua achievement master
	var allAchievements []models.Achievement
	config.DB.Order("id asc").Find(&allAchievements)

	// 2. Ambil achievement yang sudah dimiliki user
	var userAchievements []models.UserAchievement
	config.DB.Where("user_id = ?", userID).Find(&userAchievements)

	// 3. Mapping agar frontend mudah membacanya (Unlocked: true/false)
	unlockedMap := make(map[uint]bool)
	for _, ua := range userAchievements {
		unlockedMap[ua.AchievementID] = true
	}

	type AchievementResponse struct {
		models.Achievement
		IsUnlocked bool `json:"is_unlocked"`
	}

	var response []AchievementResponse
	for _, ach := range allAchievements {
		response = append(response, AchievementResponse{
			Achievement: ach,
			IsUnlocked:  unlockedMap[ach.ID],
		})
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Achievements retrieved", response)
}
