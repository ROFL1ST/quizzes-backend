package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
)

// CreateReport handles the creation of a new report
func CreateReport(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	var input struct {
		TargetID   uint   `json:"target_id"`
		TargetType string `json:"target_type"`
		Reason     string `json:"reason"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	report := models.Report{
		ReporterID: user.ID,
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

	return utils.SuccessResponse(c, fiber.StatusOK, "Reports retrieved", reports)
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
