package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
)

// CreateReport handles the creation of a new report
func CreateReport(c *fiber.Ctx) error {
	userId := uint(c.Locals("user_id").(float64))

	var input struct {
		TargetID   uint   `json:"target_id"`
		TargetType string `json:"target_type"`
		Reason     string `json:"reason"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	report := models.Report{
		ReporterID: userId,
		TargetID:   input.TargetID,
		TargetType: input.TargetType,
		Reason:     input.Reason,
		Status:     "pending",
	}

	if err := config.DB.Create(&report).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create report", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Report created successfully", report)
}

// GetAllReports retrieves all reports (Admin only)
func GetAllReports(c *fiber.Ctx) error {
	var reports []models.Report
	if err := config.DB.Preload("Reporter").Find(&reports).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch reports", err.Error())
	}

	// Enrich reports with Target Details
	var userIDs []uint
	var questionIDs []uint

	for _, r := range reports {
		if r.TargetType == "user" {
			userIDs = append(userIDs, r.TargetID)
		} else if r.TargetType == "question" {
			questionIDs = append(questionIDs, r.TargetID)
		}
	}

	usersMap := make(map[uint]string)
	questionsMap := make(map[uint]string)

	if len(userIDs) > 0 {
		var users []models.User
		config.DB.Where("id IN ?", userIDs).Find(&users)
		for _, u := range users {
			usersMap[u.ID] = u.Username
		}
	}

	if len(questionIDs) > 0 {
		var questions []models.Question
		config.DB.Where("id IN ?", questionIDs).Find(&questions)
		for _, q := range questions {
			questionsMap[q.ID] = q.QuestionText
		}
	}

	type ReportResponse struct {
		models.Report
		TargetDetail string `json:"target_detail"`
	}

	var response []ReportResponse
	for _, r := range reports {
		detail := "-"
		if r.TargetType == "user" {
			if val, ok := usersMap[r.TargetID]; ok {
				detail = val
			} else {
				detail = "Unknown User"
			}
		} else if r.TargetType == "question" {
			if val, ok := questionsMap[r.TargetID]; ok {
				// Truncate if too long
				if len(val) > 50 {
					detail = val[:47] + "..."
				} else {
					detail = val
				}
			} else {
				detail = "Unknown Question"
			}
		}
		response = append(response, ReportResponse{
			Report:       r,
			TargetDetail: detail,
		})
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Reports retrieved", response)
}

// ResolveReport updates the status of a report (Admin only)
func ResolveReport(c *fiber.Ctx) error {
	id := c.Params("id")
	var input struct {
		Status string `json:"status"` // "resolved" or "dismissed"
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	var report models.Report
	if err := config.DB.First(&report, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Report not found", nil)
	}

	report.Status = input.Status
	if err := config.DB.Save(&report).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update report", err.Error())
	}

	// Optional: Take action based on TargetType and Status (e.g., ban user, disable question)
	// For now, just update the status.

	return utils.SuccessResponse(c, fiber.StatusOK, "Report updated", report)
}
