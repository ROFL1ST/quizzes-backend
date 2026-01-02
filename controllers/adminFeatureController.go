package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
)

// BanUser bans a user
func BanUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	user.IsBanned = true
	if err := config.DB.Save(&user).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to ban user", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "User banned successfully", nil)
}

// UnbanUser unbans a user
func UnbanUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	user.IsBanned = false
	if err := config.DB.Save(&user).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to unban user", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "User unbanned successfully", nil)
}

// Broadcast creates a system-wide announcement
func Broadcast(c *fiber.Ctx) error {
	userId := uint(c.Locals("user_id").(float64))

	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Type    string `json:"type"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	announcement := models.Announcement{
		Title:     input.Title,
		Content:   input.Content,
		Type:      input.Type, // Save type
		CreatorID: userId,
		Active:    true,
	}

	if err := config.DB.Create(&announcement).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create broadcast", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Broadcast created", announcement)
}

// GetAnnouncements gets active announcements (Public/User)
func GetAnnouncements(c *fiber.Ctx) error {
	var announcements []models.Announcement
	if err := config.DB.Where("active = ?", true).Order("created_at desc").Limit(5).Find(&announcements).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch announcements", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Announcements retrieved", announcements)
}
