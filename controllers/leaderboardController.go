package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"

	"github.com/gofiber/fiber/v2"
)

type LeaderboardEntry struct {
	UserID        uint          `json:"user_id"`
	Name          string        `json:"name"`
	Username      string        `json:"username"`
	TotalPoints   int64         `json:"total_points"`
	EquippedItems []models.Item `json:"equipped_items" gorm:"-"`
}

func GetLeaderboardByTopic(c *fiber.Ctx) error {
	slug := c.Params("slug")
	userID := c.Locals("user_id").(float64)
	var topic models.Topic
	if err := config.DB.Where("slug = ?", slug).First(&topic).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Topic not found", nil)
	}

	var results []LeaderboardEntry

	err := config.DB.Table("histories").
		Select("users.id as user_id, users.name, users.username, SUM(histories.score) as total_points").
		Joins("JOIN users ON users.id = histories.user_id").
		Joins("JOIN quizzes ON quizzes.id = histories.quiz_id").
		Where("quizzes.topic_id = ?", topic.ID).
		Group("users.id, users.name, users.username").
		Order("total_points DESC").
		Limit(10).
		Scan(&results).Error

	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed calculate leaderboard", err.Error())
	}

	for i := range results {
		var items []models.Item

		config.DB.Table("items").
			Joins("JOIN user_items ON user_items.item_id = items.id").
			Where("user_items.user_id = ? AND user_items.is_equipped = ?", results[i].UserID, true).
			Find(&items)

		results[i].EquippedItems = items
	}
	utils.CheckDailyMissions(uint(userID), "social", 1, "view")
	return utils.SuccessResponse(c, fiber.StatusOK, "Leaderboard retrieved", results)
}
