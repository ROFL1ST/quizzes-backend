package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"

	"github.com/gofiber/fiber/v2"
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

	return utils.SuccessResponse(c, fiber.StatusCreated, "History saved", history)
}

func GetMyHistory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	var histories []models.History
	config.DB.Preload("User").Where("user_id = ?", userID).Order("created_at desc").Find(&histories)
	return utils.SuccessResponse(c, fiber.StatusOK, "History retrieved", histories)
}