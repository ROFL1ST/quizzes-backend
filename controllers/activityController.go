package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
)

func GetFriendActivity(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	var friendships []models.Friendship
	config.DB.Where("(user_id = ? OR friend_id = ?) AND status = 'accepted'", userID, userID).Find(&friendships)

	var friendIDs []uint
	for _, f := range friendships {
		if f.UserID == uint(userID) {
			friendIDs = append(friendIDs, f.FriendID)
		} else {
			friendIDs = append(friendIDs, f.UserID)
		}
	}
	var activities []models.Activity
	if len(friendIDs) > 0 {
		config.DB.Preload("User").Where("user_id IN ?", friendIDs).Order("created_at desc").Limit(20).Find(&activities)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Feed retrieved", activities)
}