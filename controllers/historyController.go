package controllers

import (
	"encoding/json"
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"strconv"
	"sync"
	"time"
)

type CreateHistoryInput struct {
	QuizID      uint            `json:"quiz_id" validate:"required"`
	QuizTitle   string          `json:"quiz_title"`
	Score       int             `json:"score"`
	TotalSoal   int             `json:"total_soal"`
	Snapshot    json.RawMessage `json:"snapshot"`
	TimeTaken   int             `json:"time_taken"`
	ChallengeID uint            `json:"challenge_id"`
}

func SaveHistory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)

	// Gunakan struct input baru, bukan langsung models.History
	var input CreateHistoryInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid data", err.Error())
	}

	// Mapping ke Model History
	history := models.History{
		UserID:    uint(userID),
		QuizID:    input.QuizID,
		QuizTitle: input.QuizTitle,
		Score:     input.Score,
		Snapshot:  datatypes.JSON(input.Snapshot),
		TimeTaken: input.TimeTaken,
		TotalSoal: input.TotalSoal,
	}

	// Simpan history ke Database
	if err := config.DB.Create(&history).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed save history", err.Error())
	}

	go func(uid uint, score int) {
	today := utils.StripTime(time.Now())

	// 1. Ambil 5 Misi yang Aktif Hari Ini
	var activeMissions []models.UserMission
	config.DB.Preload("Mission").
		Where("user_id = ? AND reset_date = ?", uid, today).
		Find(&activeMissions)

	for _, um := range activeMissions {
		if um.IsClaimed { continue } // Skip jika sudah klaim

		key := um.Mission.Key
		shouldSave := false

		// 2. Logic Update Progress
		if key == "play_quiz_1" || key == "play_quiz_3" || key == "play_quiz_5" {
			um.Progress++
			shouldSave = true
		} else if key == "score_100" && score == 100 {
			um.Progress++
			shouldSave = true
		} else if key == "total_score_500" {
			um.Progress += score
			shouldSave = true
		}

		if shouldSave {
			// Cap progress agar tidak melebihi target (opsional)
			// if um.Progress > um.Mission.Target { um.Progress = um.Mission.Target }
			config.DB.Save(&um)
		}
	}
}(uint(userID), input.Score)

	var currentUser models.User
	if err := config.DB.First(&currentUser, uint(userID)).Error; err == nil {
		// Jika ChallengeID ada (dikirim dari frontend), kirim sinyal ke lobby
		if input.ChallengeID != 0 {
			utils.BroadcastLobby(input.ChallengeID, "player_finished", fiber.Map{
				"user_id":  userID,
				"username": currentUser.Name, // Atau currentUser.Username
				"score":    input.Score,
				"status":   "finished",
			})
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func(uid uint, qID uint, score int, timeTaken int, challengeID uint) {
		defer wg.Done()

		var participant models.ChallengeParticipant
		var err error

		// Optimasi: Jika challengeID dikirim, cari langsung by ID
		if challengeID != 0 {
			err = config.DB.Where("challenge_id = ? AND user_id = ?", challengeID, uid).First(&participant).Error
		} else {
			// Fallback ke logic lama (cari berdasarkan quiz_id aktif terakhir)
			err = config.DB.Table("challenge_participants").
				Joins("JOIN challenges ON challenges.id = challenge_participants.challenge_id").
				Where("challenge_participants.user_id = ? AND challenges.quiz_id = ? AND challenges.status = 'active'", uid, qID).
				Order("challenges.created_at DESC").
				Select("challenge_participants.*").
				First(&participant).Error
		}

		if err == nil {
			// Update Status Partisipan
			participant.Score = score
			participant.TimeTaken = timeTaken
			participant.IsFinished = true

			config.DB.Save(&participant)

			// Cek apakah Challenge Selesai (Semua peserta 'accepted' sudah 'finished')
			var challenge models.Challenge
			if err := config.DB.Preload("Participants").First(&challenge, participant.ChallengeID).Error; err == nil {

				allFinished := true
				for _, p := range challenge.Participants {
					if p.Status == "accepted" {
						if !p.IsFinished {
							allFinished = false
							break
						}
					}
				}

				// Jika semua selesai, tutup challenge
				if allFinished {
					challenge.Status = "finished"
					config.DB.Save(&challenge)
					utils.DetermineWinner(challenge.ID)
				}
			}
		}
	}(uint(userID), history.QuizID, history.Score, history.TimeTaken, input.ChallengeID)

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

	if currentUser.ID != 0 {
		xpGained := history.Score
		currentUser.XP += int64(xpGained)
		newLevel := utils.CalculateLevel(currentUser.XP)

		if newLevel > currentUser.Level {
			currentUser.Level = newLevel

			activity := models.Activity{
				UserID:      currentUser.ID,
				Type:        "level_up",
				Description: "Naik ke Level " + strconv.Itoa(newLevel),
			}
			config.DB.Create(&activity)

			utils.SendNotification(
				currentUser.ID,
				"success",
				"Naik Level!",
				"⭐ Level Up! Kamu naik ke Level "+strconv.Itoa(newLevel),
				"/profile",
			)
		}

		config.DB.Save(&currentUser)
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
