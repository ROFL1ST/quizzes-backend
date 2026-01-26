package controllers

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"

	"github.com/gofiber/fiber/v2"
)

func CreateQuestion(c *fiber.Ctx) error {
	var q models.Question
	if err := c.BodyParser(&q); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	var count int64
	config.DB.Model(&models.Quiz{}).Where("id = ?", q.QuizID).Count(&count)
	if count == 0 {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Quiz not found", nil)
	}

	if err := config.DB.Create(&q).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed create question", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Question created", q)
}

func GetQuestionsByQuizID(c *fiber.Ctx) error {
	id := c.Params("id")
	var quiz models.Quiz
	if err := config.DB.First(&quiz, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Quiz not found", nil)
	}

	var questions []models.Question
	config.DB.Where("quiz_id = ?", quiz.ID).Order("RANDOM()").Find(&questions)

	return utils.SuccessResponse(c, fiber.StatusOK, "Questions retrieved", questions)
}

// Request structure for adaptive question
type AdaptiveRequest struct {
	QuizID  uint `json:"quiz_id"`
	Answers []struct {
		QuestionID uint   `json:"question_id"`
		UserAnswer string `json:"user_answer"`
	} `json:"answers"`
}

func GetNextAdaptiveQuestion(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64) // Assuming auth middleware

	var req AdaptiveRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid Input", err.Error())
	}

	// 1. Reconstruct History and Calculate Accuracy/Difficulty so far
	history := make([]utils.UserHistoryItem, 0)
	answeredIDs := make(map[uint]bool)

	for _, ans := range req.Answers {
		answeredIDs[ans.QuestionID] = true

		// Fetch question to check correctness and difficulty
		var q models.Question
		// Optimization: Could cache this or fetch all in one query
		if err := config.DB.First(&q, ans.QuestionID).Error; err == nil {
			// Robust Grading Logic
			isCorrect := false
			switch q.Type {
			case "short_answer":
				// Case Insensitive & Trim Space
				if strings.TrimSpace(strings.ToLower(ans.UserAnswer)) == strings.TrimSpace(strings.ToLower(q.CorrectAnswer)) {
					isCorrect = true
				}
			case "multi_select":
				var userAns []string
				var correctAns []string
				err1 := json.Unmarshal([]byte(ans.UserAnswer), &userAns)
				err2 := json.Unmarshal([]byte(q.CorrectAnswer), &correctAns)
				if err1 == nil && err2 == nil {
					if len(userAns) == len(correctAns) {
						sort.Strings(userAns)
						sort.Strings(correctAns)
						match := true
						for i := range userAns {
							if userAns[i] != correctAns[i] {
								match = false
								break
							}
						}
						if match {
							isCorrect = true
						}
					}
				}
			default:
				// mcq, boolean
				if ans.UserAnswer == q.CorrectAnswer {
					isCorrect = true
				}
			}

			history = append(history, utils.UserHistoryItem{
				QuestionID: int(q.ID),
				IsCorrect:  isCorrect,
				Difficulty: q.Difficulty,
			})
			fmt.Printf("Grading QID: %d | UserAns: %s | CorrectAns: %s | isCorrect: %v\n", q.ID, ans.UserAnswer, q.CorrectAnswer, isCorrect)
		}
	}

	// 2. Call ML Service
	mlURL := os.Getenv("ML_SERVICE_URL")
	if mlURL == "" {
		mlURL = "http://localhost:5002"
	}
	mlClient := utils.NewMLClient(mlURL)

	// Helper to get Prior Ability (Adaptive Rating)
	var priorAbility *float64
	var userAdaptivity models.UserAdaptivity
	if err := config.DB.First(&userAdaptivity, uint(userID)).Error; err == nil {
		priorAbility = &userAdaptivity.AdaptiveRating
	}

	recommendation, err := mlClient.GetRecommendation(uint(userID), req.QuizID, history, priorAbility)

	targetDiff := 0.5
	if err != nil {
		// Fallback if ML service is down
		fmt.Println("ML Service Error:", err)
		// Keep default 0.5 or logic based on accuracy
	} else {
		targetDiff = recommendation.TargetDifficulty
	}

	// 3. Find next best question
	// Fetch all questions for this quiz NOT in answeredIDs
	var candidates []models.Question
	config.DB.Where("quiz_id = ?", req.QuizID).Find(&candidates)

	var bestQuestion *models.Question
	minDiff := 100.0

	for i := range candidates {
		q := candidates[i]
		if answeredIDs[q.ID] {
			continue
		}

		diff := math.Abs(q.Difficulty - targetDiff)
		if diff < minDiff {
			minDiff = diff
			bestQuestion = &q
		}
	}

	if bestQuestion == nil {
		return utils.SuccessResponse(c, fiber.StatusOK, "Quiz Completed", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Next Question Retrieved", fiber.Map{
		"question":          bestQuestion,
		"target_difficulty": targetDiff,
		"last_is_correct": func() *bool {
			if len(history) > 0 {
				val := history[len(history)-1].IsCorrect
				return &val
			}
			return nil
		}(),
		"prediction_msg": func() string {
			if recommendation != nil {
				return recommendation.Message
			} else {
				return "Fallback"
			}
		}(),
	})
}

// Setup for dev/testing: Randomize difficulty for all questions
func RandomizeDifficulty(c *fiber.Ctx) error {
	var questions []models.Question
	if err := config.DB.Find(&questions).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch questions", err.Error())
	}

	diffs := []float64{0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9}
	for i, q := range questions {
		q.Difficulty = diffs[i%len(diffs)]
		config.DB.Save(&q)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "All question difficulties homogenized (randomized)", nil)
}

// RecalculateDifficulty updates question difficulty based on historical performance (correct/incorrect counts)
func RecalculateDifficulty(c *fiber.Ctx) error {
	var questions []models.Question
	if err := config.DB.Find(&questions).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch questions", err.Error())
	}

	updatedCount := 0
	for _, q := range questions {
		total := q.CorrectCount + q.IncorrectCount
		if total < 1 {
			// Not enough data to be statistically significant, skip or keep default
			continue
		}

		// Calculate Accuracy Rate (0.0 - 1.0)
		accuracy := float64(q.CorrectCount) / float64(total)

		// Difficulty is inverse of accuracy (High Accuracy = Low Difficulty)
		// e.g. 90% Correct (0.9) -> Difficulty 0.1 (Easy)
		// e.g. 20% Correct (0.2) -> Difficulty 0.8 (Hard)
		newDiff := 1.0 - accuracy

		// Clamp logic (optional, keeping it raw for now or standard rounding)
		newDiff = math.Round(newDiff*100) / 100

		// Update if changed significantly (> 0.05 diff)
		if math.Abs(q.Difficulty-newDiff) > 0.01 {
			q.Difficulty = newDiff
			config.DB.Save(&q)
			updatedCount++
		}
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fmt.Sprintf("Recalculated difficulty for %d questions based on analysis", updatedCount), nil)
}
