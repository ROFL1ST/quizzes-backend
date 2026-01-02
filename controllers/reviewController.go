package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
)

// AddReview allows a user to review a quiz
func AddReview(c *fiber.Ctx) error {
	userId := uint(c.Locals("user_id").(float64))
	quizID := c.Params("id")

	var input struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	if input.Rating < 1 || input.Rating > 5 {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Rating must be between 1 and 5", nil)
	}

	var quiz models.Quiz
	if err := config.DB.First(&quiz, quizID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Quiz not found", nil)
	}

	// Check if user already reviewed
	var existing models.QuizReview
	if err := config.DB.Where("user_id = ? AND quiz_id = ?", userId, quizID).First(&existing).Error; err == nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "You have already reviewed this quiz", nil)
	}

	review := models.QuizReview{
		UserID:  userId,
		QuizID:  quiz.ID,
		Rating:  input.Rating,
		Comment: input.Comment,
	}

	if err := config.DB.Create(&review).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to add review", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Review added successfully", review)
}

// GetReviews retrieves reviews for a specific quiz
func GetReviews(c *fiber.Ctx) error {
	quizID := c.Params("id")

	var reviews []models.QuizReview
	if err := config.DB.Preload("User").Where("quiz_id = ?", quizID).Find(&reviews).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch reviews", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Reviews retrieved", reviews)
}

// GetAllReviews retrieves all reviews (Admin only)
func GetAllReviews(c *fiber.Ctx) error {
	var reviews []models.QuizReview
	// Preload Quiz and User
	if err := config.DB.Preload("User").Preload("Quiz").Order("created_at desc").Find(&reviews).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch reviews", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "All reviews retrieved", reviews)
}

// DeleteReview removes a review (Admin only)
func DeleteReview(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := config.DB.Delete(&models.QuizReview{}, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete review", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Review deleted successfully", nil)
}
