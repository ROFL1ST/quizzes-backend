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
)

type QuestionAnalysis struct {
	ID             uint   `json:"id"`
	QuestionText   string `json:"question_text"`
	CorrectCount   int    `json:"correct_count"`
	IncorrectCount int    `json:"incorrect_count"`
	TotalAttempts  int    `json:"total_attempts"`
	Difficulty     string `json:"difficulty"`    // Mudah, Sedang, Sulit
	AccuracyRate   string `json:"accuracy_rate"` // Contoh: "85.5%"
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
