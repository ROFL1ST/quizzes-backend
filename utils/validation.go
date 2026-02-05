package utils

import "github.com/gofiber/fiber/v2"

// ValidateCredentials validates username and password for registration
// Returns an error response if validation fails, nil otherwise
func ValidateCredentials(c *fiber.Ctx, username, password string) error {
	if len(username) < 3 {
		return ErrorResponse(c, fiber.StatusBadRequest, "Username must be at least 3 characters", nil)
	}
	if len(password) < 6 {
		return ErrorResponse(c, fiber.StatusBadRequest, "Password must be at least 6 characters", nil)
	}
	return nil
}

// ValidatePassword validates password length
// Returns an error response if validation fails, nil otherwise
func ValidatePassword(c *fiber.Ctx, password string) error {
	if len(password) < 6 {
		return ErrorResponse(c, fiber.StatusBadRequest, "Password must be at least 6 characters", nil)
	}
	return nil
}
