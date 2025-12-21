package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
)

func GetDailyInfo(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	today := utils.StripTime(utils.GetJakartaTime())

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	// 1. Cek Status Klaim Hadiah (Login Reward)
	var todayClaim models.DailyClaim
	err := config.DB.Where("user_id = ? AND reward_type = ? AND claimed_date = ?", userID, "login", today).First(&todayClaim).Error

	displayDay := 0
	displayStatus := ""

	if err == nil {
		// SUDAH KLAIM HARI INI
		displayDay = user.LoginStreak
		displayStatus = "cooldown"
	} else {
		// BELUM KLAIM HARI INI
		lastClaim := utils.GetJakartaTime().AddDate(0, 0, -100) // Default lama
		if user.LastClaimDate != nil {
			lastClaim = *user.LastClaimDate
		}

		diff := utils.DaysBetween(lastClaim, utils.GetJakartaTime())

		if diff > 1 {
			displayDay = 1 // Bakal Reset
		} else {
			displayDay = user.LoginStreak + 1 // Bakal Lanjut
		}
		displayStatus = "claimable"
	}

	// Hitung Hadiah Berdasarkan Siklus 100 Hari
	cycleDay := displayDay % 100
	if cycleDay == 0 {
		cycleDay = 100
	}

	var rewardConfig models.DailyRewardConfig
	if err := config.DB.Where("day = ?", cycleDay).First(&rewardConfig).Error; err != nil {
		rewardConfig.Reward = 20 // Default reward
	}

	// 2. Cek Misi Harian (Assign jika belum ada)
	utils.AssignDailyMissions(uint(userID))

	var userMissions []models.UserMission
	config.DB.Preload("Mission").
		Where("user_id = ? AND reset_date = ?", userID, today).
		Find(&userMissions)

	var missionResponse []fiber.Map
	for _, um := range userMissions {
		status := "locked"
		if um.IsClaimed {
			status = "claimed"
		} else if um.Progress >= um.Mission.Target {
			status = "claimable"
		}

		missionResponse = append(missionResponse, fiber.Map{
			"id":          um.MissionID,
			"title":       um.Mission.Title,
			"description": um.Mission.Description,
			"reward":      um.Mission.Reward,
			"target":      um.Mission.Target,
			"progress":    um.Progress,
			"status":      status,
		})
	}

	// [BARU] Cek apakah Quiz Streak sudah dikerjakan hari ini?
	isQuizDone := false
	if user.LastActivityDate != nil {
		lastActivity := utils.StripTime(*user.LastActivityDate)
		// Bandingkan tanggal aktivitas terakhir dengan hari ini (WIB)
		if lastActivity.Equal(today) {
			isQuizDone = true
		}
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Info retrieved", fiber.Map{
		"streak": fiber.Map{
			"day":          displayDay, // Hari login (untuk UI Card Hadiah)
			"reward":       rewardConfig.Reward,
			"status":       displayStatus,
			"quiz_streak":  user.StreakCount, // Total streak kuis
			"is_quiz_done": isQuizDone,       // Status pengerjaan hari ini
		},
		"missions": missionResponse,
	})
}

func ClaimLoginReward(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)

	// Gunakan Waktu Jakarta
	today := utils.StripTime(utils.GetJakartaTime())

	// 1. Cek Duplikat
	var exists int64
	config.DB.Model(&models.DailyClaim{}).
		Where("user_id = ? AND reward_type = ? AND claimed_date = ?", userID, "login", today).
		Count(&exists)

	if exists > 0 {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Login reward already claimed today", nil)
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	// 2. UPDATE LOGIN STREAK
	// Panggil fungsi yang sudah kita buat di utils/streak.go
	utils.UpdateLoginStreak(&user)

	// 3. Hitung Hadiah
	cycleDay := user.LoginStreak % 100
	if cycleDay == 0 {
		cycleDay = 100
	}

	var rewardConfig models.DailyRewardConfig
	if err := config.DB.Where("day = ?", cycleDay).First(&rewardConfig).Error; err != nil {
		rewardConfig.Reward = 20
	}

	// 4. Simpan Log & Update User
	newClaim := models.DailyClaim{
		UserID:      user.ID,
		RewardType:  "login",
		ClaimedDate: today,
	}
	config.DB.Create(&newClaim)

	user.Coins += rewardConfig.Reward
	config.DB.Save(&user) // Menyimpan LoginStreak, LastClaimDate, dan Coins

	return utils.SuccessResponse(c, fiber.StatusOK, "Login reward claimed", fiber.Map{
		"coins_gained": rewardConfig.Reward,
		"new_balance":  user.Coins,
		"login_streak": user.LoginStreak,
		"quiz_streak":  user.StreakCount,
	})
}

// POST /api/daily/claim-mission
func ClaimMissionReward(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	var input struct {
		MissionID uint `json:"mission_id"`
	}
	c.BodyParser(&input)

	today := utils.StripTime(utils.GetJakartaTime())
	var um models.UserMission

	// Cari misi spesifik yg ditugaskan hari ini
	if err := config.DB.Preload("Mission").Where("user_id = ? AND mission_id = ? AND reset_date = ?", userID, input.MissionID, today).First(&um).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Mission not active today", nil)
	}

	if um.IsClaimed {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Already claimed", nil)
	}
	if um.Progress < um.Mission.Target {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Not finished", nil)
	}

	// Klaim
	um.IsClaimed = true
	config.DB.Save(&um)

	var user models.User
	config.DB.First(&user, userID)
	user.Coins += um.Mission.Reward
	config.DB.Save(&user)

	return utils.SuccessResponse(c, fiber.StatusOK, "Mission claimed", fiber.Map{"new_coins": user.Coins, "reward": um.Mission.Reward})
}
