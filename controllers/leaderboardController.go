package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"

	"github.com/gofiber/fiber/v2"
)

type LeaderboardEntry struct {
	Name        string `json:"name"`
	Username    string `json:"username"`
	TotalPoints int64  `json:"total_points"`
}

func GetLeaderboardByTopic(c *fiber.Ctx) error {
	slug := c.Params("slug")

	var topic models.Topic
	if err := config.DB.Where("slug = ?", slug).First(&topic).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Topic not found", nil)
	}

	var results []LeaderboardEntry

	err := config.DB.Table("histories").
		Select("users.name, users.username, SUM(histories.score) as total_points").
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

	return utils.SuccessResponse(c, fiber.StatusOK, "Leaderboard retrieved", results)
}