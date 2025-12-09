package utils

import "github.com/gofiber/fiber/v2"

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
	return c.Status(statusCode).JSON(ApiResponse{
		Status:  "error",
		Message: message,
		Error:   err,
	})
}