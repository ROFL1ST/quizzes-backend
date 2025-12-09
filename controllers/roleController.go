package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"

	"github.com/gofiber/fiber/v2"
)

func CreateRole(c *fiber.Ctx) error {
	var role models.Role
	if err := c.BodyParser(&role); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	if err := config.DB.Create(&role).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed create role", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Role created", role)
}

func GetAllRoles(c *fiber.Ctx) error {
	var roles []models.Role
	config.DB.Find(&roles)
	return utils.SuccessResponse(c, fiber.StatusOK, "Roles retrieved", roles)
}