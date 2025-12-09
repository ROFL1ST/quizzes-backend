package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"

	"github.com/gofiber/fiber/v2"
)

func CreateQuestion(c *fiber.Ctx) error {
	var q models.Question
	if err := c.BodyParser(&q); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	var count int64
	config.DB.Model(&models.Quiz{}).Where("id = ?", q.QuizID).Count(&count)
	if count == 0 {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Quiz not found", nil)
	}

	if err := config.DB.Create(&q).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed create question", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Question created", q)
}

func GetQuestionsByQuizID(c *fiber.Ctx) error {
	id := c.Params("id")
	var quiz models.Quiz
	if err := config.DB.First(&quiz, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Quiz not found", nil)
	}

	var questions []models.Question
	config.DB.Where("quiz_id = ?", quiz.ID).Order("RANDOM()").Find(&questions)

	return utils.SuccessResponse(c, fiber.StatusOK, "Questions retrieved", questions)
}