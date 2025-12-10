package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"

	"github.com/gofiber/fiber/v2"
)

func CreateQuiz(c *fiber.Ctx) error {
	var quiz models.Quiz
	if err := c.BodyParser(&quiz); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	var count int64
	config.DB.Model(&models.Topic{}).Where("id = ?", quiz.TopicID).Count(&count)
	if count == 0 {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Topic not found", nil)
	}

	if err := config.DB.Create(&quiz).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed create quiz", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Quiz created", quiz)
}

func GetQuizzesByTopicSlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var topic models.Topic
	if err := config.DB.Where("slug = ?", slug).First(&topic).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Topic not found", nil)
	}

	var quizzes []models.Quiz
	config.DB.Where("(topic_id = ?) AND active=TRUE", topic.ID).Find(&quizzes)

	return utils.SuccessResponse(c, fiber.StatusOK, "Quizzes retrieved", quizzes)
}