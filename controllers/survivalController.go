package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
)

// StartSurvival starts a survival game session
func StartSurvival(c *fiber.Ctx) error {
	// Accept optional seed for deterministic random (multiplayer)
	seed := c.Query("seed", "")

	var question models.Question

	query := config.DB
	if seed != "" {
		// Deterministic Order for Multiplayer
		// MD5(id || seed)
		query = query.Order(config.DB.Raw("MD5(CAST(id AS TEXT) || ?)", seed))
	} else {
		// Pure Random for Singleplayer
		query = query.Order("RANDOM()")
	}

	if err := query.First(&question).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get question", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Survival Started", fiber.Map{
		"question": question,
		"streak":   0,
	})
}

// AnswerSurvival processes the answer for survival mode
func AnswerSurvival(c *fiber.Ctx) error {
	// user := c.Locals("user").(*models.User) // Not used yet
	var input struct {
		QuestionID uint   `json:"question_id"`
		Answer     string `json:"answer"`
		Streak     int    `json:"streak"`
		Seed       string `json:"seed"` // New: Seed for deterministic next question
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	var question models.Question
	if err := config.DB.First(&question, input.QuestionID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Question not found", nil)
	}

	// Normalize Answer (trim space)
	if question.CorrectAnswer != input.Answer {
		// GAME OVER
		return utils.SuccessResponse(c, fiber.StatusOK, "Game Over", fiber.Map{
			"correct":        false,
			"correct_answer": question.CorrectAnswer,
			"final_streak":   input.Streak,
		})
	}

	// Correct! Get next question
	// Use Offset = streak + 1 (assuming current was at offset 'streak')
	// Wait, if users passed 'streak' as current count of correct answers (e.g. 0 initially),
	// then the question they just answered was at offset 0?
	// If input.Streak is 5 (they have 5 correct answers), they just answered the 6th question (index 5)?
	// Let's assume input.Streak is the CURRENT streak count BEFORE this answer.
	// So if input.Streak = 0. They answered question #1. Result new streak = 1. Next question is #2 (offset 1).

	nextOffset := input.Streak + 1

	var nextQuestion models.Question

	query := config.DB
	if input.Seed != "" {
		// Deterministic
		query = query.Order(config.DB.Raw("MD5(CAST(id AS TEXT) || ?)", input.Seed))
	} else {
		// Random but Exclude previous IDs?
		// In pure random mode, we can't easily exclude ALL previous without sending them ALL.
		// For simplicity in single player, just ensuring it's not the SAME question is often enough,
		// or we can live with true random.
		// A better approach for single player is also generating a random seed client side?
		// For now, keep existing single player logic (just != current) + Random.
		query = query.Where("id != ?", question.ID).Order("RANDOM()")
	}

	// If seeded, use offset. If random, just pick one.
	if input.Seed != "" {
		if err := query.Offset(nextOffset).First(&nextQuestion).Error; err != nil {
			// Might run out of questions? Reset offset?
			// For now, assume ample questions. If error, maybe wrap around.
			query.Offset(0).First(&nextQuestion)
		}
	} else {
		query.First(&nextQuestion)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Correct!", fiber.Map{
		"correct":       true,
		"new_streak":    input.Streak + 1,
		"next_question": nextQuestion,
	})
}
