package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
	// "golang.org/x/crypto/bcrypt"
	"fmt"
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

func SearchUsers(c *fiber.Ctx) error {
	
	query := c.Query("q")
	if query == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Query parameter 'q' is required", nil)
	}

	var users []models.User
	
	if err := config.DB.Where("username ILIKE ? OR name ILIKE ?", "%"+query+"%", "%"+query+"%").Limit(10).Find(&users).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to search users", nil)
	}

	
	type SearchResponse struct {
		models.User
		EquippedItems []models.Item `json:"equipped_items"`
	}

	var results []SearchResponse

	// 3. Fix: Loop setiap user untuk ambil item masing-masing
	for _, u := range users {
		var items []models.Item
		
		config.DB.Table("items").
			Joins("JOIN user_items ON user_items.item_id = items.id").
			Where("user_items.user_id = ? AND user_items.is_equipped = ?", u.ID, true).
			Find(&items)

		// Masukkan ke list hasil
		results = append(results, SearchResponse{
			User:          u,
			EquippedItems: items,
		})
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Users retrieved", results)
}

func GetMyProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	var totalQuizzes int64
	var avgScore float64
	config.DB.Model(&models.History{}).Where("user_id = ?", userID).Count(&totalQuizzes)

	config.DB.Model(&models.History{}).Where("user_id = ?", userID).
		Select("COALESCE(AVG(score), 0)").Scan(&avgScore)

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

	calculatedLevel := utils.CalculateLevel(user.XP)

	if calculatedLevel > user.Level {
		user.Level = calculatedLevel
		config.DB.Save(&user) // Simpan perbaikan ke database
	}
	currentLevel := user.Level
	nextLevel := currentLevel + 1

	// Hitung batas bawah level saat ini dan batas level selanjutnya
	currentLevelBaseXP := utils.CalculateMinXPForLevel(currentLevel)
	nextLevelThreshold := utils.CalculateMinXPForLevel(nextLevel)

	levelProgress := fiber.Map{
		"current_level":    currentLevel,
		"current_xp":       user.XP,
		"level_base_xp":    currentLevelBaseXP,
		"next_level_xp":    nextLevelThreshold,
		"xp_needed":        nextLevelThreshold - user.XP,
		"progress_percent": 0,
	}

	rangeXP := nextLevelThreshold - currentLevelBaseXP
	if rangeXP > 0 {
		progress := float64(user.XP-currentLevelBaseXP) / float64(rangeXP) * 100

		if progress > 100 {
			progress = 100
		}

		levelProgress["progress_percent"] = math.Round(progress*10) / 10
	} else {
		levelProgress["progress_percent"] = 100
	}

	var equippedItems []models.Item
	config.DB.Table("items").
		Joins("JOIN user_items ON user_items.item_id = items.id").
		Where("user_items.user_id = ? AND user_items.is_equipped = ?", userID, true).
		Find(&equippedItems)
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
		"equipped_items":    equippedItems,
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Profile retrieved", response)
}

func UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Username string `json:"username"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", nil)
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	// Update Nama
	if input.Name != "" {
		user.Name = input.Name
	}

	// Update Username
	if input.Username != "" && input.Username != user.Username {
		// Cek apakah username sudah dipakai user lain
		var checkUser models.User
		if err := config.DB.Where("username = ? AND id != ?", input.Username, user.ID).First(&checkUser).Error; err == nil {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Username already in use", nil)
		}
		user.Username = input.Username
	}

	// === LOGIKA UPDATE EMAIL ===
	if input.Email != "" && input.Email != user.Email {
		// 1. Cek apakah email sudah dipakai user lain
		var checkUser models.User
		if err := config.DB.Where("email = ? AND id != ?", input.Email, user.ID).First(&checkUser).Error; err == nil {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Email already in use", nil)
		}

		// 2. Set Email baru & Reset status verifikasi
		user.Email = input.Email
		user.IsEmailVerified = false

		// 3. Generate Token menggunakan Utility baru
		user.EmailVerificationToken = utils.GenerateToken()

		// 4. Kirim Email Verifikasi (Gunakan Goroutine agar tidak blocking)
		go func(emailAddr, tokenStr string) {
			// Pastikan fungsi SendVerificationEmail sudah ada di utils/email.go
			err := utils.SendVerificationEmail(emailAddr, tokenStr)
			if err != nil {
				fmt.Println("Error sending verification email:", err)
			} else {
				fmt.Println("Verification email sent to:", emailAddr)
			}
		}(user.Email, user.EmailVerificationToken)
	}

	if err := config.DB.Save(&user).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update profile", err.Error())
	}

	// Sembunyikan token dari response
	user.EmailVerificationToken = ""

	return utils.SuccessResponse(c, fiber.StatusOK, "Profile updated. Please check email to verify.", fiber.Map{"user": user})
}

func VerifyEmail(c *fiber.Ctx) error {
	var input struct {
		Token string `json:"token"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", nil)
	}

	var user models.User
	if err := config.DB.Where("email_verification_token = ?", input.Token).First(&user).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid or expired token", nil)
	}

	user.IsEmailVerified = true
	user.EmailVerificationToken = "" // Hapus token setelah dipakai
	config.DB.Save(&user)

	return utils.SuccessResponse(c, fiber.StatusOK, "Email verified successfully", nil)
}

func GetUserProfile(c *fiber.Ctx) error {
	username := c.Params("username")

	var user models.User
	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	// 1. Statistik Dasar
	var totalQuizzes int64
	config.DB.Model(&models.History{}).Where("user_id = ?", user.ID).Count(&totalQuizzes)

	var totalWins int64
	config.DB.Model(&models.Challenge{}).Where("winner_id = ?", user.ID).Count(&totalWins)

	stats := fiber.Map{
		"xp":            user.XP,
		"level":         user.Level,
		"total_quizzes": totalQuizzes,
		"total_wins":    totalWins,        // Tambahan
		"streak_count":  user.StreakCount, // Tambahan (Duolingo style)
		"joined_at":     user.CreatedAt,
	}

	// 2. Ambil Achievements Public (Hanya yang sudah unlock)
	type PublicAchievement struct {
		Name        string `json:"name"`
		IconURL     string `json:"icon_url"`
		Description string `json:"description"`
	}
	var achievements []PublicAchievement

	config.DB.Table("user_achievements").
		Select("achievements.name, achievements.icon_url, achievements.description").
		Joins("JOIN achievements ON achievements.id = user_achievements.achievement_id").
		Where("user_achievements.user_id = ?", user.ID).
		Scan(&achievements)

	var equippedItems []models.Item
	config.DB.Table("items").
		Joins("JOIN user_items ON user_items.item_id = items.id").
		Where("user_items.user_id = ? AND user_items.is_equipped = ?", user.ID, true).
		Find(&equippedItems)

	return utils.SuccessResponse(c, fiber.StatusOK, "User profile retrieved", fiber.Map{
		"id":             user.ID, // Penting untuk cek friend status
		"name":           user.Name,
		"username":       user.Username,
		"stats":          stats,
		"achievements":   achievements, // Data baru
		"equipped_items": equippedItems,
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


func ShareProfileTrigger(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)

	utils.CheckDailyMissions(uint(userID), "social", 1, "share")

	return utils.SuccessResponse(c, fiber.StatusOK, "Share event recorded", nil)
}