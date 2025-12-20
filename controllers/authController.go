package controllers

import (
	"fmt"
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strconv"
	"time"
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
	if err := config.DB.Preload("UserItems", "is_equipped = ?", true).Preload("UserItems.Item").Where("username = ?", input.Username).First(&user).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}
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
	streakMessage := ""

	if user.LastActivityDate != nil {

		last := *user.LastActivityDate
		y1, m1, d1 := last.Date()
		y2, m2, d2 := now.Date()

		dateLast := time.Date(y1, m1, d1, 0, 0, 0, 0, time.Local)
		dateNow := time.Date(y2, m2, d2, 0, 0, 0, 0, time.Local)

		daysDiff := int(dateNow.Sub(dateLast).Hours() / 24)

		if daysDiff == 1 {

			newStreak := user.StreakCount + 1
			streakMessage = fmt.Sprintf("🔥 Streak Lanjut! Hari ke-%d.", newStreak)
		} else if daysDiff > 1 {
			streakMessage = "😢 Streak terputus. Mulai lagi dari hari ke-1."
		} else {
			streakMessage = fmt.Sprintf("🔥 Streak Hari ke-%d aman.", user.StreakCount)
		}
	} else {
		streakMessage = "👋 Selamat datang! Streak hari ke-1 dimulai."
	}

	utils.UpdateStreak(&user)

	config.DB.Omit("UserItems").Save(&user)

	y, m, d := now.Date()
	todayStripped := time.Date(y, m, d, 0, 0, 0, 0, time.Local)

	var claimedCount int64
	config.DB.Model(&models.DailyClaim{}).
		Where("user_id = ? AND reward_type = ? AND claimed_date = ?", user.ID, "login", todayStripped).
		Count(&claimedCount)

	if claimedCount == 0 {
		streakMessage += " Jangan lupa klaim Koin di menu Daily Reward!"
	} else {
		streakMessage += " Koin hari ini sudah diklaim."
	}
	config.DB.Save(&user)
	currentHour := utils.GetJakartaTime().Hour()
	utils.CheckDailyMissions(user.ID, "login", 0, strconv.Itoa(currentHour))
	return utils.SuccessResponse(c, fiber.StatusOK, "Login success", fiber.Map{
		"token":          t,
		"user":           user,
		"streak_message": streakMessage,
	})
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

func AuthMe(c *fiber.Ctx) error {
	userIDLocals := c.Locals("user_id")
	role := c.Locals("role").(string)
	userID := uint(userIDLocals.(float64))
	if role == "user" {

		var user models.User
		if err := config.DB.Preload("UserItems", "is_equipped = ?", true).Preload("UserItems.Item").First(&user, userID).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
		}

		now := time.Now()
		user.LastActivityDate = &now
		config.DB.Save(&user)

		claims := jwt.MapClaims{
			"user_id": user.ID,
			"role":    "user",
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		t, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

		var equippedItems []models.Item
		config.DB.Table("items").
			Joins("JOIN user_items ON user_items.item_id = items.id").
			Where("user_items.user_id = ? AND user_items.is_equipped = ?", user.ID, true).
			Find(&equippedItems)

		streakMessage := ""

		if user.LastActivityDate != nil {

			last := *user.LastActivityDate
			y1, m1, d1 := last.Date()
			y2, m2, d2 := now.Date()

			dateLast := time.Date(y1, m1, d1, 0, 0, 0, 0, time.Local)
			dateNow := time.Date(y2, m2, d2, 0, 0, 0, 0, time.Local)

			daysDiff := int(dateNow.Sub(dateLast).Hours() / 24)

			if daysDiff == 1 {

				newStreak := user.StreakCount + 1
				streakMessage = fmt.Sprintf("🔥 Streak Lanjut! Hari ke-%d.", newStreak)
			} else if daysDiff > 1 {
				streakMessage = "😢 Streak terputus. Mulai lagi dari hari ke-1."
			} else {
				streakMessage = fmt.Sprintf("🔥 Streak Hari ke-%d aman.", user.StreakCount)
			}
		} else {
			streakMessage = "👋 Selamat datang! Streak hari ke-1 dimulai."
		}

		utils.UpdateStreak(&user)

		config.DB.Omit("UserItems").Save(&user)

		y, m, d := now.Date()
		todayStripped := time.Date(y, m, d, 0, 0, 0, 0, time.Local)

		var claimedCount int64
		config.DB.Model(&models.DailyClaim{}).
			Where("user_id = ? AND reward_type = ? AND claimed_date = ?", user.ID, "login", todayStripped).
			Count(&claimedCount)

		if claimedCount == 0 {
			streakMessage += " Jangan lupa klaim Koin di menu Daily Reward!"
		} else {
			streakMessage += " Koin hari ini sudah diklaim."
		}

		currentHour := utils.GetJakartaTime().Hour()
		utils.CheckDailyMissions(user.ID, "login", 0, strconv.Itoa(currentHour))
		return utils.SuccessResponse(c, fiber.StatusOK, "User session refreshed", fiber.Map{
			"token":          t,
			"user":           user,
			"role":           "user",
			"equipped_items": equippedItems,
			"streak_message": streakMessage,
		})

	} else {

		var admin models.Admin
		// Preload Role untuk memastikan data role admin terbaru
		if err := config.DB.Preload("Role").First(&admin, userID).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Admin not found", nil)
		}

		if admin.Role.Name != role {

			role = admin.Role.Name
		}

		claims := jwt.MapClaims{
			"user_id": admin.ID,
			"role":    role,
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		t, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

		return utils.SuccessResponse(c, fiber.StatusOK, "Admin session refreshed", fiber.Map{
			"token": t,
			"user":  admin,
			"role":  role,
		})
	}
}

func ForgotPassword(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", nil)
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		// Return 404 jika email tidak ditemukan
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Email not found", nil)
	}

	// Cek apakah email sudah diverifikasi
	if !user.IsEmailVerified {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Email not verified. Please verify via settings first.", nil)
	}

	// 1. Generate Token menggunakan Utility
	token := utils.GenerateToken()

	// 2. Simpan ke tabel PasswordReset
	// Hapus token lama jika ada
	config.DB.Where("email = ?", input.Email).Delete(&models.PasswordReset{})

	resetEntry := models.PasswordReset{
		Email:     input.Email,
		Token:     token,
		ExpiredAt: time.Now().Add(1 * time.Hour), // Token berlaku 1 jam
	}
	config.DB.Create(&resetEntry)

	// 3. Kirim Email Reset (Goroutine)
	go func(emailAddr, tokenStr string) {
		err := utils.SendResetPasswordEmail(emailAddr, tokenStr)
		if err != nil {
			fmt.Println("Error sending reset email:", err)
		} else {
			fmt.Println("Reset email sent to:", emailAddr)
		}
	}(input.Email, token)

	return utils.SuccessResponse(c, fiber.StatusOK, "Password reset link sent to your email", nil)
}

// ResetPassword menangani perubahan password dengan token
func ResetPassword(c *fiber.Ctx) error {
	var input struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", nil)
	}

	// 1. Validasi Token
	var resetData models.PasswordReset
	if err := config.DB.Where("token = ? AND expired_at > ?", input.Token, time.Now()).First(&resetData).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid or expired token", nil)
	}

	// 2. Hash Password Baru
	hashed, _ := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 10)

	// 3. Update Password User
	// Kita cari user berdasarkan email yang tersimpan di tabel reset
	if err := config.DB.Model(&models.User{}).Where("email = ?", resetData.Email).Update("password", string(hashed)).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update password", nil)
	}

	// 4. Hapus token agar tidak bisa dipakai lagi
	config.DB.Delete(&resetData)

	return utils.SuccessResponse(c, fiber.StatusOK, "Password updated successfully. Please login.", nil)
}
