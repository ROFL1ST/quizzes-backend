package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"

	"github.com/gofiber/fiber/v2"
)

func CreateTopic(c *fiber.Ctx) error {
	var topic models.Topic
	if err := c.BodyParser(&topic); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	if err := config.DB.Create(&topic).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed create topic", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Topic created", topic)
}

func GetAllTopics(c *fiber.Ctx) error {
	var topics []models.Topic
	config.DB.Find(&topics)
	return utils.SuccessResponse(c, fiber.StatusOK, "Topics retrieved", topics)
}