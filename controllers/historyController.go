package controllers

import (
	"encoding/json"
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"strconv"
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
	go utils.CheckQuizAchievements(history.UserID, history.Score)
	go func(uid uint, qID uint, score int) {
		var challenge models.Challenge
		err := config.DB.Where(
			"quiz_id = ? AND status = 'active' AND (challenger_id = ? OR opponent_id = ?)",
			qID, uid, uid,
		).First(&challenge).Error

		if err == nil {
			// Update skor user yang mengerjakan
			if challenge.ChallengerID == uid {
				challenge.ChallengerScore = score
			} else {
				challenge.OpponentScore = score
			}

			// Cek apakah kedua pemain sudah bermain (skor != -1)
			if challenge.ChallengerScore != -1 && challenge.OpponentScore != -1 {
				challenge.Status = "finished"

				// Tentukan Pemenang
				if challenge.ChallengerScore > challenge.OpponentScore {
					winner := challenge.ChallengerID
					challenge.WinnerID = &winner
				} else if challenge.OpponentScore > challenge.ChallengerScore {
					winner := challenge.OpponentID
					challenge.WinnerID = &winner
				}
				// Jika seri, WinnerID tetap null
			}
			config.DB.Save(&challenge)
		}
	}(uint(userID), history.QuizID, history.Score)
	go func(quizID uint, snapshotJSON []byte) {
		// Ambil kunci jawaban
		var questions []models.Question
		config.DB.Where("quiz_id = ?", quizID).Find(&questions)
		keyMap := make(map[uint]string)
		for _, q := range questions {
			keyMap[q.ID] = q.CorrectAnswer
		}

		// Parsing jawaban user
		var userAnswers map[string]string
		json.Unmarshal(snapshotJSON, &userAnswers)

		// Update counter benar/salah
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

		// Tambah XP
		xpGained := history.Score
		user.XP += int64(xpGained)

		// Cek Level Up (Rumus: Level = XP / 1000 + 1)
		newLevel := int((user.XP / 1000)) + 1
		if newLevel > user.Level {
			user.Level = newLevel

			// Simpan Activity Level Up
			activity := models.Activity{
				UserID:      user.ID,
				Type:        "level_up",
				Description: "Naik ke Level " + strconv.Itoa(newLevel),
			}
			config.DB.Create(&activity)
		}
		config.DB.Save(&user)

		// Simpan Activity Quiz Completed
		feed := models.Activity{
			UserID:      user.ID,
			Type:        "quiz_completed",
			Description: "Menyelesaikan kuis " + history.QuizTitle + " dengan skor " + strconv.Itoa(history.Score),
		}
		config.DB.Create(&feed)
	}

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
		"questions":  questions,
		"created_at": history.CreatedAt,
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "History retrieved", response)
}
