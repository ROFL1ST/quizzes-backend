package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"

	"github.com/gofiber/fiber/v2"
)

func AddFriend(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	var input struct {
		Username string `json:"username"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	var friend models.User
	if err := config.DB.Where("username = ?", input.Username).First(&friend).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	if friend.ID == uint(userID) {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Cannot add self", nil)
	}

	var check models.Friendship
	if config.DB.Where("user_id = ? AND friend_id = ?", userID, friend.ID).First(&check).RowsAffected > 0 {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Already friends", nil)
	}

	friendship := models.Friendship{UserID: uint(userID), FriendID: friend.ID}
	config.DB.Create(&friendship)

	return utils.SuccessResponse(c, fiber.StatusCreated, "Friend added", nil)
}

func GetMyFriends(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	var friends []models.Friendship
	config.DB.Preload("Friend").Where("user_id = ?", userID).Find(&friends)
	return utils.SuccessResponse(c, fiber.StatusOK, "Friends retrieved", friends)
}

func RemoveFriend(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	friendID := c.Params("id")
	config.DB.Where("user_id = ? AND friend_id = ?", userID, friendID).Delete(&models.Friendship{})
	return utils.SuccessResponse(c, fiber.StatusOK, "Friend removed", nil)
}