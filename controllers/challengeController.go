package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
)

func CreateChallenge(c *fiber.Ctx) error {
	challengerID := c.Locals("user_id").(float64)
	var input struct {
		OpponentUsername string `json:"opponent_username"`
		QuizID           uint   `json:"quiz_id"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	var opponent models.User
	if err := config.DB.Where("username = ?", input.OpponentUsername).First(&opponent).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Opponent not found", nil)
	}

	challenge := models.Challenge{
		ChallengerID: uint(challengerID),
		OpponentID:   opponent.ID,
		QuizID:       input.QuizID,
		Status:       "pending",
	}

	config.DB.Create(&challenge)

	utils.SendNotification(opponent.ID, "⚔️ Kamu ditantang duel kuis!", "/challenges", "warning")
	return utils.SuccessResponse(c, fiber.StatusCreated, "Challenge sent", challenge)
}

func GetMyChallenges(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)

	var challenges []models.Challenge

	config.DB.
		Preload("Quiz").
		Preload("Challenger").
		Preload("Opponent").
		Where("(challenger_id = ? OR opponent_id = ?)", userID, userID).
		Order("created_at DESC").
		Find(&challenges)

	return utils.SuccessResponse(c, fiber.StatusOK, "Challenges retrieved", challenges)
}


func AcceptChallenge(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(float64)

	var challenge models.Challenge
	if err := config.DB.First(&challenge, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Challenge not found", nil)
	}

	if challenge.OpponentID != uint(userID) {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "You are not the opponent", nil)
	}

	if challenge.Status != "pending" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Challenge is not pending", nil)
	}

	challenge.Status = "active"
	config.DB.Save(&challenge)

	return utils.SuccessResponse(c, fiber.StatusOK, "Challenge accepted! Game on!", challenge)
}

func RejectChallenge(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(float64)
	var challenge models.Challenge
	if err := config.DB.First(&challenge, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Challenge not found", nil)
	}
	if challenge.OpponentID != uint(userID) {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "You are not the opponent", nil)
	}
	if challenge.Status != "pending" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Challenge is not pending", nil)
	}
	challenge.Status = "rejected"
	config.DB.Save(&challenge)
	return utils.SuccessResponse(c, fiber.StatusOK, "Challenge rejected", challenge)
}
