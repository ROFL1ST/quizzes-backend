package utils

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
)

type ApiResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func SuccessResponse(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(ApiResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *fiber.Ctx, statusCode int, message string, err interface{}) error {
	// In production, do not leak internal error details (unless it's a specific logical error we want to show)
	// assuming "err" is often a raw error string or error object.
	if os.Getenv("GO_ENV") == "production" || os.Getenv("VERCEL_ENV") == "production" {
		// Log the actual error internally
		fmt.Printf("INTERNAL ERROR [%s]: %v\n", message, err)

		// If status is 500, mask the error
		if statusCode == fiber.StatusInternalServerError {
			return c.Status(statusCode).JSON(ApiResponse{
				Status:  "error",
				Message: "Internal Server Error",
				Error:   nil, // Don't send raw error
			})
		}
	}

	return c.Status(statusCode).JSON(ApiResponse{
		Status:  "error",
		Message: message,
		Error:   err,
	})
}
