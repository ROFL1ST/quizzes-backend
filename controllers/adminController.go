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
	TotalAttempts int64   `json:"total_attempts"`
	AverageScore  float64 `json:"average_score"`
	ActiveUsers   int64   `json:"active_users"`
}

type QuestionAnalysis struct {
	ID             uint   `json:"id"`
	QuestionText   string `json:"question_text"`
	CorrectCount   int    `json:"correct_count"`
	IncorrectCount int    `json:"incorrect_count"`
	TotalAttempts  int    `json:"total_attempts"`
	Difficulty     string `json:"difficulty"`
	AccuracyRate   string `json:"accuracy_rate"`
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

func GetAllUsers(c *fiber.Ctx) error {
	var users []models.User
	if err := config.DB.Order("created_at desc").Find(&users).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch users", nil)
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Users retrieved", users)
}

// ========== TOPICS WITH PAGINATION ==========
func GetAllTopicsAdmin(c *fiber.Ctx) error {
	params := utils.GetPaginationParams(c)
	
	var topics []models.Topic
	var total int64

	// Hitung total data
	if err := config.DB.Model(&models.Topic{}).Count(&total).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to count topics", nil)
	}

	// Ambil data dengan pagination
	if err := config.DB.
		Order("created_at desc").
		Limit(params.PageSize).
		Offset(params.Offset).
		Find(&topics).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch topics", nil)
	}

	return utils.PaginatedSuccessResponse(c, fiber.StatusOK, "Topics retrieved", topics, total, params)
}

func PostTopicAdmin(c *fiber.Ctx) error {
	var topic models.Topic
	if err := c.BodyParser(&topic); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	if err := config.DB.Create(&topic).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed create topic", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, "Topic created", topic)
}

func DeleteTopicAdmin(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := config.DB.Delete(&models.Topic{}, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed delete topic", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Topic deleted", nil)
}

func UpdateTopicAdmin(c *fiber.Ctx) error {
	id := c.Params("slug")
	var topic models.Topic
	if err := config.DB.First(&topic, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Topic not found", nil)
	}
	if err := c.BodyParser(&topic); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}
	if err := config.DB.Save(&topic).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed update topic", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Topic updated", topic)
}

// ========== QUIZZES WITH PAGINATION ==========
func GetAllQuizzesAdmin(c *fiber.Ctx) error {
	params := utils.GetPaginationParams(c)
	
	var quizzes []models.Quiz
	var total int64

	// Hitung total data
	if err := config.DB.Model(&models.Quiz{}).Count(&total).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to count quizzes", nil)
	}

	// Ambil data dengan pagination
	if err := config.DB.
		Preload("Topic").
		Order("created_at desc").
		Limit(params.PageSize).
		Offset(params.Offset).
		Find(&quizzes).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch quizzes", nil)
	}

	return utils.PaginatedSuccessResponse(c, fiber.StatusOK, "Quizzes retrieved", quizzes, total, params)
}

func PostQuizAdmin(c *fiber.Ctx) error {
	var quiz models.Quiz
	if err := c.BodyParser(&quiz); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}
	if err := config.DB.Create(&quiz).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed create quiz", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, "Quiz created", quiz)
}

func UpdateQuizAdmin(c *fiber.Ctx) error {
	id := c.Params("id")
	var quiz models.Quiz
	if err := config.DB.First(&quiz, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Quiz not found", nil)
	}
	if err := c.BodyParser(&quiz); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}
	if err := config.DB.Save(&quiz).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed update quiz", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Quiz updated", quiz)
}

func DeleteQuizAdmin(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := config.DB.Delete(&models.Quiz{}, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed delete quiz", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Quiz deleted", nil)
}

func GetQuizAnalysisAdminById(c *fiber.Ctx) error {
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

// ========== QUESTIONS WITH PAGINATION ==========
func GetAllQuestionsAdmin(c *fiber.Ctx) error {
	params := utils.GetPaginationParams(c)
	
	var questions []models.Question
	var total int64

	// Filter berdasarkan quiz_id jika ada
	query := config.DB.Model(&models.Question{})
	
	if quizID := c.Query("quiz_id"); quizID != "" {
		query = query.Where("quiz_id = ?", quizID)
	}

	// Hitung total data
	if err := query.Count(&total).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to count questions", nil)
	}

	// Ambil data dengan pagination
	if err := query.
		Preload("Quiz").
		Order("created_at desc").
		Limit(params.PageSize).
		Offset(params.Offset).
		Find(&questions).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch questions", nil)
	}

	return utils.PaginatedSuccessResponse(c, fiber.StatusOK, "Questions retrieved", questions, total, params)
}

func PostQuestionAdmin(c *fiber.Ctx) error {
	var question models.Question
	if err := c.BodyParser(&question); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}
	if err := config.DB.Create(&question).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed create question", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, "Question created", question)
}

func UpdateQuestionAdmin(c *fiber.Ctx) error {
	id := c.Params("id")
	var question models.Question
	if err := config.DB.First(&question, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Question not found", nil)
	}
	if err := c.BodyParser(&question); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}
	if err := config.DB.Save(&question).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed update question", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Question updated", question)
}

func DeleteQuestionAdmin(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := config.DB.Delete(&models.Question{}, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed delete question", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Question deleted", nil)
}

func BulkUploadQuestions(c *fiber.Ctx) error {
	quizID, _ := strconv.Atoi(c.FormValue("quiz_id"))

	file, err := c.FormFile("file")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "File required", nil)
	}

	f, _ := file.Open()
	defer f.Close()

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
			Hint: row[6],
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