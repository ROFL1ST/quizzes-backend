package controllers

import (
	"encoding/csv"
	"fmt"
	"strconv"

	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
	"time"
)

type DashboardAnalytics struct {
	TotalUsers    int64   `json:"total_users"`
	TotalQuizzes  int64   `json:"total_quizzes"`
	TotalAttempts int64   `json:"total_attempts"` // Total riwayat pengerjaan
	AverageScore  float64 `json:"average_score"`
	ActiveUsers   int64   `json:"active_users"` // User aktif 7 hari terakhir
}

type QuestionAnalysis struct {
	ID             uint   `json:"id"`
	QuestionText   string `json:"question_text"`
	CorrectCount   int    `json:"correct_count"`
	IncorrectCount int    `json:"incorrect_count"`
	TotalAttempts  int    `json:"total_attempts"`
	Difficulty     string `json:"difficulty"`    // Mudah, Sedang, Sulit
	AccuracyRate   string `json:"accuracy_rate"` // Contoh: "85.5%"
}

func GetDashboardAnalytics(c *fiber.Ctx) error {
	var totalUsers, totalQuizzes, totalAttempts int64
	var avgScore float64

	if err := config.DB.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Error counting users", nil)
	}

	if err := config.DB.Model(&models.Quiz{}).Count(&totalQuizzes).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Error counting quizzes", nil)
	}

	if err := config.DB.Model(&models.History{}).Count(&totalAttempts).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Error counting history", nil)
	}

	if err := config.DB.Model(&models.History{}).Select("COALESCE(AVG(score), 0)").Scan(&avgScore).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Error calculating average", nil)
	}

	var activeUsers int64
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	if err := config.DB.Model(&models.History{}).
		Where("created_at >= ?", sevenDaysAgo).
		Distinct("user_id").
		Count(&activeUsers).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Error counting active users", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Dashboard analytics retrieved", DashboardAnalytics{
		TotalUsers:    totalUsers,
		TotalQuizzes:  totalQuizzes,
		TotalAttempts: totalAttempts,
		AverageScore:  avgScore,
		ActiveUsers:   activeUsers,
	})
}

func BulkUploadQuestions(c *fiber.Ctx) error {
	quizID, _ := strconv.Atoi(c.FormValue("quiz_id"))

	// Ambil file dari form-data
	file, err := c.FormFile("file")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "File required", nil)
	}

	f, _ := file.Open()
	defer f.Close()

	// Baca CSV
	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Failed parse CSV", nil)
	}

	var questions []models.Question
	for i, row := range records {
		if i == 0 {
			continue
		}

		if len(row) < 6 {
			continue
		}

		q := models.Question{
			QuizID:        uint(quizID),
			QuestionText:  row[0],
			Options:       pq.StringArray{row[1], row[2], row[3], row[4]},
			CorrectAnswer: row[5],
		}
		questions = append(questions, q)
	}

	if err := config.DB.Create(&questions).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed insert questions", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Bulk upload success", fiber.Map{
		"total_inserted": len(questions),
	})
}

func GetQuizAnalysis(c *fiber.Ctx) error {
	quizID := c.Params("id")

	var questions []models.Question
	if err := config.DB.Where("quiz_id = ?", quizID).Order("id asc").Find(&questions).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Questions not found", nil)
	}

	var analysis []QuestionAnalysis

	for _, q := range questions {
		total := q.CorrectCount + q.IncorrectCount
		accuracy := 0.0
		difficulty := "Belum ada data"

		if total > 0 {
			accuracy = (float64(q.CorrectCount) / float64(total)) * 100
			if accuracy > 80 {
				difficulty = "Mudah"
			} else if accuracy > 40 {
				difficulty = "Sedang"
			} else {
				difficulty = "Sulit"
			}
		}

		analysis = append(analysis, QuestionAnalysis{
			ID:             q.ID,
			QuestionText:   q.QuestionText,
			CorrectCount:   q.CorrectCount,
			IncorrectCount: q.IncorrectCount,
			TotalAttempts:  total,
			Difficulty:     difficulty,
			AccuracyRate:   fmt.Sprintf("%.1f%%", accuracy),
		})
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Question analysis retrieved", analysis)
}

func GetAllUsers(c *fiber.Ctx) error {
	var users []models.User
	// Ambil semua user, urutkan dari yang terbaru
	if err := config.DB.Order("created_at desc").Find(&users).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch users", nil)
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Users retrieved", users)
}

// GET ALL QUIZZES (Untuk Halaman Manajemen Kuis)
func GetAllQuizzesAdmin(c *fiber.Ctx) error {
	var quizzes []models.Quiz
	// Preload Topic agar nama topik muncul
	if err := config.DB.Preload("Topic").Order("created_at desc").Find(&quizzes).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch quizzes", nil)
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Quizzes retrieved", quizzes)
}
