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

	

	// Simpan history
	history.UserID = uint(userID)
	if err := config.DB.Create(&history).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed save history", err.Error())
	}

	var wg sync.WaitGroup
	wg.Add(1)
    go func(uid uint, qID uint, score int, timeTaken int) {
        defer wg.Done()

        var participant models.ChallengeParticipant

        // 1. Cari Partisipan (Query tetap sama supaya aman dari ambiguous ID)
        err := config.DB.Table("challenge_participants").
            Joins("JOIN challenges ON challenges.id = challenge_participants.challenge_id").
            Where("challenge_participants.user_id = ? AND challenges.quiz_id = ? AND challenges.status = 'active'", uid, qID).
            Order("challenges.created_at DESC"). // <--- TAMBAHAN KECIL TAPI PENTING
            Select("challenge_participants.*").
            First(&participant).Error

        if err == nil {
            // --- UPDATE 1: Set Flag IsFinished ---
            participant.Score = score
            participant.TimeTaken = timeTaken
            participant.IsFinished = true // <--- PENTING: Tandai sudah selesai
            
            // Simpan perubahan participant
            config.DB.Save(&participant)

            // 2. Cek apakah Challenge Selesai (Semua peserta 'accepted' sudah 'finished')
            var challenge models.Challenge
            // Preload participants untuk cek status teman mabar
            if err := config.DB.Preload("Participants").First(&challenge, participant.ChallengeID).Error; err == nil {
                
                allFinished := true
                for _, p := range challenge.Participants {
                    // Hanya cek user yang statusnya 'accepted'.
                    // User 'pending' atau 'rejected' tidak dihitung.
                    if p.Status == "accepted" {
                        // Jika ada SATU saja yang belum finish, maka game belum over.
                        if !p.IsFinished { 
                            allFinished = false
                            break
                        }
                    }
                }

                // --- UPDATE 2: Jika semua selesai, tutup challenge ---
                if allFinished {
                    challenge.Status = "finished"
                    config.DB.Save(&challenge)
                    utils.DetermineWinner(challenge.ID)
                    // Opsional: Logic penentuan pemenang bisa ditaruh disini
                    // utils.DetermineWinner(challenge.ID) 
                }
            }
        }
    }(uint(userID), history.QuizID, history.Score, history.TimeTaken)
	// ----------------------------
	// UPDATE QUESTION SNAPSHOT
	// ----------------------------
	go func(quizID uint, snapshotJSON []byte) {
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
						UpdateColumn("correct_count", gorm.Expr("correct_count + 1"))
				} else {
					config.DB.Model(&models.Question{}).Where("id = ?", qID).
						UpdateColumn("incorrect_count", gorm.Expr("incorrect_count + 1"))
				}
			}
		}
	}(history.QuizID, history.Snapshot)

	// ----------------------------
	// UPDATE XP USER
	// ----------------------------
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

			utils.SendNotification(
				user.ID,
				"⭐ Level Up! Kamu naik ke Level "+strconv.Itoa(newLevel),
				"/profile",
				"success",
			)
		}

		config.DB.Save(&user)
	}

	// ----------------------------
	// CEK ACHIEVEMENT SETELAH SELESAI
	// ----------------------------
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
