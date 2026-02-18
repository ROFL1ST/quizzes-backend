package utils

import (
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
	var payload interface{}
	if os.Getenv("APP_DEBUG") == "true" {
		payload = err
	}

	return c.Status(statusCode).JSON(ApiResponse{
		Status:  "error",
		Message: message,
		Error:   payload,
	})
}
