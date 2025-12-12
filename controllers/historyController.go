package controllers

import (
	"encoding/json"
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"strconv"
	"sync"
)

func SaveHistory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	var history models.History

	if err := c.BodyParser(&history); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid data", err.Error())
	}

	history.UserID = uint(userID)
	if err := config.DB.Create(&history).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed save history", err.Error())
	}

	var wg sync.WaitGroup

	// Update Logika Challenge (Support Battle Royale)
	wg.Add(1)
	go func(uid uint, qID uint, score int, timeTaken int) {
		defer wg.Done()

		// 1. Cari apakah User ini tergabung dalam Challenge aktif untuk kuis ini
		var participant models.ChallengeParticipant
		err := config.DB.Joins("JOIN challenges ON challenges.id = challenge_participants.challenge_id").
			Where("challenge_participants.user_id = ? AND challenges.quiz_id = ? AND challenges.status = 'active'", uid, qID).
			First(&participant).Error

		if err == nil {
			// Update skor peserta
			participant.Score = score
			participant.TimeTaken = timeTaken
			config.DB.Save(&participant)

			// Cek apakah semua peserta sudah selesai?
			var challenge models.Challenge
			config.DB.Preload("Participants").First(&challenge, participant.ChallengeID)

			allFinished := true
			for _, p := range challenge.Participants {
				if p.Score == -1 { // Masih ada yang belum selesai
					allFinished = false
					break
				}
			}

			if allFinished {
				challenge.Status = "finished"
				config.DB.Save(&challenge)
				// Disini bisa tambahkan notifikasi ke semua peserta siapa pemenangnya
			}
		}
	}(uint(userID), history.QuizID, history.Score, history.TimeTaken)

	// ... (Logika XP dan Achievement tetap sama seperti file asli) ...
	// Note: Pastikan snapshot processing tetap berjalan
	go func(quizID uint, snapshotJSON []byte) {
		// ... (Kode snapshot analysis dari file asli tetap dipakai) ...
		var questions []models.Question
		config.DB.Where("quiz_id = ?", quizID).Find(&questions)
		keyMap := make(map[uint]string)
		for _, q := range questions {
			keyMap[q.ID] = q.CorrectAnswer
		}

		var userAnswers map[string]string
		json.Unmarshal(snapshotJSON, &userAnswers)

		for qIDStr, answer := range userAnswers {
			qID, _ := strconv.Atoi(qIDStr)
			if correctAnswer, exists := keyMap[uint(qID)]; exists {
				if answer == correctAnswer {
					config.DB.Model(&models.Question{}).Where("id = ?", qID).
						UpdateColumn("correct_count", gorm.Expr("correct_count + ?", 1))
				} else {
					config.DB.Model(&models.Question{}).Where("id = ?", qID).
						UpdateColumn("incorrect_count", gorm.Expr("incorrect_count + ?", 1))
				}
			}
		}
	}(history.QuizID, history.Snapshot)
	var user models.User
	if err := config.DB.First(&user, uint(userID)).Error; err == nil {
		xpGained := history.Score
		user.XP += int64(xpGained)
		newLevel := utils.CalculateLevel(user.XP)

		if newLevel > user.Level {
			user.Level = newLevel
			activity := models.Activity{
				UserID:      user.ID,
				Type:        "level_up",
				Description: "Naik ke Level " + strconv.Itoa(newLevel),
			}
			config.DB.Create(&activity)
			utils.SendNotification(user.ID, "⭐ Level Up! Kamu naik ke Level "+strconv.Itoa(newLevel), "/profile", "success")
		}
		config.DB.Save(&user)
	}

	go func() {
		wg.Wait()
		utils.CheckQuizAchievements(history.UserID, history.Score)
	}()

	return utils.SuccessResponse(c, fiber.StatusCreated, "History saved", history)
}

func GetMyHistory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	var histories []models.History
	config.DB.Preload("User").Where("user_id = ?", userID).Order("created_at desc").Find(&histories)
	return utils.SuccessResponse(c, fiber.StatusOK, "History retrieved", histories)
}

func GetHistoryByID(c *fiber.Ctx) error {
	id := c.Params("id")

	var history models.History
	if err := config.DB.First(&history, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "History not found", nil)
	}

	var questions []models.Question
	if err := config.DB.Where("quiz_id = ?", history.QuizID).Find(&questions).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve questions", nil)
	}
	response := fiber.Map{
		"id":         history.ID,
		"quiz_title": history.QuizTitle,
		"score":      history.Score,
		"snapshot":   history.Snapshot,
		"time_taken": history.TimeTaken,
		"questions":  questions,
		"created_at": history.CreatedAt,
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "History retrieved", response)
}
