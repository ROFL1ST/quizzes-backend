package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
	"time"
)

func GetDailyInfo(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	userIDUint := uint(userID)

	// Pastikan misi harian ter-assign
	utils.AssignDailyMissions(userIDUint)

	var user models.User
	config.DB.First(&user, userID)
	today := utils.StripTime(time.Now())

	// 1. CEK STATUS KLAIM LOGIN
	var todayClaim models.DailyClaim
	err := config.DB.Where("user_id = ? AND reward_type = ? AND claimed_date = ?", userID, "login", today).First(&todayClaim).Error

	// --- LOGIC BARU: TENTUKAN HARI & STATUS ---
	displayDay := 0
	displayStatus := ""

	if err == nil {

		displayDay = user.StreakCount + 1
		displayStatus = "cooldown" // Status baru: Menunggu besok
	} else {

		displayDay = user.StreakCount
		displayStatus = "claimable"
	}

	// Hitung Siklus 100 Hari
	cycleDay := displayDay % 100
	if cycleDay == 0 {
		cycleDay = 100
	}

	// Ambil Config Reward untuk hari yang ditampilkan
	var rewardConfig models.DailyRewardConfig
	if err := config.DB.Where("day = ?", cycleDay).First(&rewardConfig).Error; err != nil {
		rewardConfig.Reward = 20 // Fallback
	}

	// 2. INFO MISI HARIAN (Tetap Sama)
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

	return utils.SuccessResponse(c, fiber.StatusOK, "Info retrieved", fiber.Map{
		"streak": fiber.Map{
			"day":    displayDay,
			"reward": rewardConfig.Reward,
			"status": displayStatus,
		},
		"missions": missionResponse,
	})
}

func ClaimLoginReward(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	today := utils.StripTime(time.Now())

	// 1. CEK DUPLIKAT: Apakah sudah klaim hari ini?
	var exists int64
	config.DB.Model(&models.DailyClaim{}).
		Where("user_id = ? AND reward_type = ? AND claimed_date = ?", userID, "login", today).
		Count(&exists)

	if exists > 0 {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Login reward already claimed today", nil)
	}

	// 2. AMBIL USER DATA
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	// 3. HITUNG REWARD (Siklus 100 Hari)
	// Gunakan StreakCount user yang sudah diupdate saat login
	cycleDay := user.StreakCount % 100
	if cycleDay == 0 {
		cycleDay = 100
	}

	// Ambil config hadiah dari database sesuai hari ke-X
	var rewardConfig models.DailyRewardConfig
	if err := config.DB.Where("day = ?", cycleDay).First(&rewardConfig).Error; err != nil {
		// Fallback jika lupa seeding atau data tidak ketemu
		rewardConfig.Reward = 20
	}

	newClaim := models.DailyClaim{
		UserID:      user.ID,
		RewardType:  "login",
		ClaimedDate: today,
	}

	// Simpan Log & Update Koin User dalam satu flow
	if err := config.DB.Create(&newClaim).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to record claim", nil)
	}

	user.Coins += rewardConfig.Reward
	config.DB.Save(&user)

	// 5. RETURN RESPONSE
	return utils.SuccessResponse(c, fiber.StatusOK, "Login reward claimed", fiber.Map{
		"coins_gained": rewardConfig.Reward,
		"new_balance":  user.Coins,
		"streak_day":   user.StreakCount,
	})
}

// POST /api/daily/claim-mission
func ClaimMissionReward(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	var input struct {
		MissionID uint `json:"mission_id"`
	}
	c.BodyParser(&input)

	today := utils.StripTime(time.Now())
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

	return utils.SuccessResponse(c, fiber.StatusOK, "Mission claimed", fiber.Map{"new_coins": user.Coins})
}
