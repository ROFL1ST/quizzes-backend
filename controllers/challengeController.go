package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
)

type CreateChallengeInput struct {
	OpponentUsernames []string `json:"opponent_usernames"`
	QuizID            uint     `json:"quiz_id"`
	Mode              string   `json:"mode"`
	TimeLimit         int      `json:"time_limit"`
	IsRealtime        bool     `json:"is_realtime"`
}

func CreateChallenge(c *fiber.Ctx) error {
	creatorID := c.Locals("user_id").(float64)
	var input CreateChallengeInput

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	// 1. Buat Header Challenge
	challenge := models.Challenge{
		CreatorID:  uint(creatorID),
		QuizID:     input.QuizID,
		Mode:       input.Mode,
		TimeLimit:  input.TimeLimit,
		IsRealtime: input.IsRealtime,
		Status:     "pending",
	}

	if err := config.DB.Create(&challenge).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed create challenge", err.Error())
	}

	// 2. Masukkan Creator sebagai Peserta (Status Accepted)
	creatorPart := models.ChallengeParticipant{
		ChallengeID: challenge.ID,
		UserID:      uint(creatorID),
		Status:      "accepted", // Pembuat otomatis accept
	}
	config.DB.Create(&creatorPart)

	// 3. Masukkan Lawan sebagai Peserta (Status Pending)
	if len(input.OpponentUsernames) > 0 {
		var opponents []models.User
		config.DB.Where("username IN ?", input.OpponentUsernames).Find(&opponents)

		for _, opp := range opponents {
			if opp.ID == uint(creatorID) {
				continue
			} // Skip diri sendiri

			part := models.ChallengeParticipant{
				ChallengeID: challenge.ID,
				UserID:      opp.ID,
				Status:      "pending",
			}
			config.DB.Create(&part)

			// Kirim notifikasi ke lawan
			utils.SendNotification(opp.ID, "⚔️ Kamu ditantang main "+input.Mode+"!", "/challenges", "warning")
		}
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Challenge created", challenge)
}

func GetMyChallenges(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)

	var participants []models.ChallengeParticipant
	// Ambil ID challenge dimana user terlibat
	config.DB.Where("user_id = ?", userID).Find(&participants)

	var challengeIDs []uint
	for _, p := range participants {
		challengeIDs = append(challengeIDs, p.ChallengeID)
	}

	var challenges []models.Challenge
	if len(challengeIDs) > 0 {
		config.DB.
			Preload("Quiz").
			Preload("Creator").
			Preload("Participants.User"). // Load user detail tiap peserta
			Where("id IN ?", challengeIDs).
			Order("created_at DESC").
			Find(&challenges)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Challenges retrieved", challenges)
}

func AcceptChallenge(c *fiber.Ctx) error {
	id := c.Params("id") // Challenge ID
	userID := c.Locals("user_id").(float64)

	var participant models.ChallengeParticipant
	if err := config.DB.Where("challenge_id = ? AND user_id = ?", id, userID).First(&participant).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "You are not in this challenge", nil)
	}

	if participant.Status != "pending" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Already responded", nil)
	}

	participant.Status = "accepted"
	config.DB.Save(&participant)

	// Cek apakah Challenge bisa dimulai (misal jika realtime, tunggu semua accept. Jika tidak, langsung aktif)
	var challenge models.Challenge
	config.DB.First(&challenge, id)

	// Sederhana: Begitu ada 1 yang accept, status challenge jadi active (atau sesuaikan logic game)
	if challenge.Status == "pending" {
		challenge.Status = "active"
		config.DB.Save(&challenge)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Challenge accepted!", nil)
}

func RejectChallenge(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(float64)

	var participant models.ChallengeParticipant
	if err := config.DB.Where("challenge_id = ? AND user_id = ?", id, userID).First(&participant).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "You are not in this challenge", nil)
	}

	participant.Status = "rejected"
	config.DB.Save(&participant)

	return utils.SuccessResponse(c, fiber.StatusOK, "Challenge rejected", nil)
}
