package controllers

import (
	"os"
	"time"

	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(c *fiber.Ctx) error {
	// Gunakan struct sementara, JANGAN models.User
	var input struct {
		Name     string `json:"name"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input data", err.Error())
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 10)

	newUser := models.User{
		Name:     input.Name,
		Username: input.Username,
		Password: string(hashed),
	}

	if err := config.DB.Create(&newUser).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Username already exists", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "User registered successfully", newUser)
}

func LoginUser(c *fiber.Ctx) error {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	var user models.User
	config.DB.Where("username = ?", input.Username).First(&user)
	if user.ID == 0 {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Incorrect password", nil)
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    "user",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	now := time.Now()
	if user.LastActivityDate != nil {
		diff := now.Sub(*user.LastActivityDate).Hours()

		if diff >= 24 && diff < 48 {
			user.StreakCount++
		} else if diff >= 48 {
			user.StreakCount = 1
		}
	} else {
		user.StreakCount = 1
	}

	user.LastActivityDate = &now
	config.DB.Save(&user)
	return utils.SuccessResponse(c, fiber.StatusOK, "Login success", fiber.Map{"token": t, "user": user})
}

func RegisterAdmin(c *fiber.Ctx) error {
	var input struct {
		Name     string `json:"name"`
		Username string `json:"username"`
		Password string `json:"password"`
		RoleID   uint   `json:"role_id"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	var role models.Role
	if err := config.DB.First(&role, input.RoleID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid Role ID", nil)
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	admin := models.Admin{Name: input.Name, Username: input.Username, Password: string(hashed), RoleID: input.RoleID}

	if err := config.DB.Create(&admin).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Username exists", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Admin registered", admin)
}

func LoginAdmin(c *fiber.Ctx) error {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	var admin models.Admin
	if err := config.DB.Preload("Role").Where("username = ?", input.Username).First(&admin).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Admin not found", nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(input.Password)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Incorrect password", nil)
	}

	claims := jwt.MapClaims{
		"user_id": admin.ID,
		"role":    admin.Role.Name,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	return utils.SuccessResponse(c, fiber.StatusOK, "Login success", fiber.Map{"token": t, "user": admin, "role": admin.Role.Name})
}
