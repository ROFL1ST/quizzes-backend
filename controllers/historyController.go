package controllers

import (
	"encoding/json"
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CreateHistoryInput struct {
	QuizID      uint            `json:"quiz_id" validate:"required"`
	QuizTitle   string          `json:"quiz_title"`
	Score       int             `json:"score"`
	TotalSoal   int             `json:"total_soal"`
	Snapshot    json.RawMessage `json:"snapshot"`
	TimeTaken   int             `json:"time_taken"`
	ChallengeID uint            `json:"challenge_id"`
}

func SaveHistory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)

	var input CreateHistoryInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid data", err.Error())
	}

	// =================================================================
	// 1. LOGIKA PENILAIAN (GRADING)
	// =================================================================
	var questions []models.Question
	if err := config.DB.Where("quiz_id = ?", input.QuizID).Find(&questions).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch questions", err.Error())
	}

	questionMap := make(map[uint]models.Question)
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	var userAnswers map[string]string
	if err := json.Unmarshal(input.Snapshot, &userAnswers); err != nil {
		userAnswers = make(map[string]string)
	}

	correctCount := 0
	totalQuestions := len(questions)

	// Hitung Benar/Salah
	for qIDStr, answer := range userAnswers {
		qID, _ := strconv.Atoi(qIDStr)
		if q, exists := questionMap[uint(qID)]; exists {
			isCorrect := false

			switch q.Type {
			case "short_answer":
				// Case Insensitive & Trim Space
				userAns := strings.ToLower(strings.TrimSpace(answer))
				correctAns := strings.ToLower(strings.TrimSpace(q.CorrectAnswer))
				if userAns == correctAns {
					isCorrect = true
				}

			case "multi_select":
				var userAns []string
				var correctAns []string

				err1 := json.Unmarshal([]byte(answer), &userAns)
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

			case "boolean", "mcq":
				if answer == q.CorrectAnswer {
					isCorrect = true
				}

			default:
				if answer == q.CorrectAnswer {
					isCorrect = true
				}
			}

			if isCorrect {
				correctCount++
			}
		}
	}

	// Hitung Final Score (0-100)
	finalScore := 0
	if totalQuestions > 0 {
		finalScore = int(math.Round(float64(correctCount) / float64(totalQuestions) * 100))
	}

	history := models.History{
		UserID:    uint(userID),
		QuizID:    input.QuizID,
		QuizTitle: input.QuizTitle,
		Score:     finalScore,
		Snapshot:  datatypes.JSON(input.Snapshot),
		TimeTaken: input.TimeTaken,
		TotalSoal: totalQuestions,
	}

	if err := config.DB.Create(&history).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed save history", err.Error())
	}

	// A. Update Misi Harian
	go func(uid uint, score int) {
		today := utils.StripTime(time.Now())
		var activeMissions []models.UserMission
		config.DB.Preload("Mission").
			Where("user_id = ? AND reset_date = ?", uid, today).
			Find(&activeMissions)

		for _, um := range activeMissions {
			if um.IsClaimed {
				continue
			}

			key := um.Mission.Key
			shouldSave := false

			if key == "play_quiz_1" || key == "play_quiz_3" || key == "play_quiz_5" {
				um.Progress++
				shouldSave = true
			} else if key == "score_100" && score == 100 {
				um.Progress++
				shouldSave = true
			} else if key == "total_score_500" {
				um.Progress += score
				shouldSave = true
			}

			if shouldSave {
				config.DB.Save(&um)
			}
		}
	}(uint(userID), finalScore)

	var currentUser models.User
	if err := config.DB.First(&currentUser, uint(userID)).Error; err == nil {
		// Broadcast Lobby (Realtime) - Memberitahu pemain lain bahwa user ini selesai
		if input.ChallengeID != 0 {
			utils.BroadcastLobby(input.ChallengeID, "player_finished", fiber.Map{
				"user_id":  userID,
				"username": currentUser.Name,
				"score":    finalScore,
				"status":   "finished",
			})
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// B. Update Challenge Status (Hanya jika ChallengeID Valid)
	go func(uid uint, score int, timeTaken int, challengeID uint) {
		defer wg.Done()

		if challengeID == 0 {
			return
		}

		var participant models.ChallengeParticipant

		// Cari partisipan spesifik untuk challenge ini
		if err := config.DB.Where("challenge_id = ? AND user_id = ?", challengeID, uid).First(&participant).Error; err != nil {
			return // Data partisipan tidak ditemukan, abaikan
		}

		// Update Score & Status Selesai User Ini
		participant.Score = score
		participant.TimeTaken = timeTaken
		participant.IsFinished = true
		config.DB.Save(&participant)

		// Cek apakah SEMUA peserta (accepted) sudah selesai?
		var challenge models.Challenge
		if err := config.DB.Preload("Participants").First(&challenge, challengeID).Error; err == nil {
			allFinished := true

			// Loop semua peserta
			for _, p := range challenge.Participants {
				// Hanya cek yang sudah ACCEPT challenge (yang pending/reject ga dihitung)
				if p.Status == "accepted" {
					if !p.IsFinished {
						allFinished = false
						break
					}
				}
			}

			// Jika semua sudah selesai, tutup challenge & tentukan pemenang
			if allFinished {
				challenge.Status = "finished"
				config.DB.Save(&challenge)
				utils.DetermineWinner(challenge.ID)
			}
		}
	}(uint(userID), finalScore, history.TimeTaken, input.ChallengeID)

	// C. Update Statistik Soal
	go func(qMap map[uint]models.Question, uAns map[string]string) {
		for qIDStr, answer := range uAns {
			qID, _ := strconv.Atoi(qIDStr)
			if q, exists := qMap[uint(qID)]; exists {
				isCorrect := false
				switch q.Type {
				case "short_answer":
					if strings.EqualFold(strings.TrimSpace(answer), strings.TrimSpace(q.CorrectAnswer)) {
						isCorrect = true
					}
				case "multi_select":
					var ua, ca []string
					json.Unmarshal([]byte(answer), &ua)
					json.Unmarshal([]byte(q.CorrectAnswer), &ca)
					if len(ua) == len(ca) {
						sort.Strings(ua)
						sort.Strings(ca)
						match := true
						for i := range ua {
							if ua[i] != ca[i] {
								match = false
								break
							}
						}
						if match {
							isCorrect = true
						}
					}
				default:
					if answer == q.CorrectAnswer {
						isCorrect = true
					}
				}

				if isCorrect {
					config.DB.Model(&models.Question{}).Where("id = ?", qID).UpdateColumn("correct_count", gorm.Expr("correct_count + 1"))
				} else {
					config.DB.Model(&models.Question{}).Where("id = ?", qID).UpdateColumn("incorrect_count", gorm.Expr("incorrect_count + 1"))
				}
			}
		}
	}(questionMap, userAnswers)

	// D. Level Up & Notification
	if currentUser.ID != 0 {
		xpGained := finalScore
		currentUser.XP += int64(xpGained)
		newLevel := utils.CalculateLevel(currentUser.XP)

		if newLevel > currentUser.Level {
			currentUser.Level = newLevel
			activity := models.Activity{UserID: currentUser.ID, Type: "level_up", Description: "Naik ke Level " + strconv.Itoa(newLevel)}
			config.DB.Create(&activity)

			utils.SendNotification(currentUser.ID, "success", "Naik Level!", "⭐ Level Up! Kamu naik ke Level "+strconv.Itoa(newLevel), "/profile")
			utils.CheckDailyMissions(currentUser.ID, "level", 0, "levelup")
		}
		config.DB.Save(&currentUser)
	}

	go func() {
		wg.Wait()
		utils.CheckQuizAchievements(history.UserID, finalScore)
	}()

	utils.CheckDailyMissions(currentUser.ID, "quiz", finalScore, history.QuizTitle)
	utils.CheckDailyMissions(currentUser.ID, "level", finalScore, "xp_gain")

	return utils.SuccessResponse(c, fiber.StatusCreated, "History saved", history)
}
func GetMyHistory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)

	params := utils.GetPaginationParams(c)

	var histories []models.History
	var total int64

	query := config.DB.Model(&models.History{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to count history", err.Error())
	}

	var avgScore float64

	if err := config.DB.Model(&models.History{}).
		Where("user_id = ?", userID).
		Select("COALESCE(AVG(score), 0)").
		Scan(&avgScore).Error; err != nil {
		avgScore = 0
	}

	if err := query.Preload("User").
		Order("created_at desc").
		Offset(params.Offset).
		Limit(params.PageSize).
		Find(&histories).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch history", err.Error())
	}

	responseData := fiber.Map{
		"list": histories,
		"stats": fiber.Map{
			"total_quiz":    total,
			"average_score": int(math.Round(avgScore)),
		},
	}

	return utils.PaginatedSuccessResponse(c, fiber.StatusOK, "History retrieved", responseData, total, params)
}

func GetHistoryByID(c *fiber.Ctx) error {
	id := c.Params("id")

	var history models.History
	if err := config.DB.First(&history, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "History not found", nil)
	}

	var questions []models.Question
	if err := config.DB.Where("quiz_id = ?", history.QuizID).Find(&questions).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve questions", nil)
	}
	response := fiber.Map{
		"id":         history.ID,
		"quiz_title": history.QuizTitle,
		"score":      history.Score,
		"snapshot":   history.Snapshot,
		"time_taken": history.TimeTaken,
		"questions":  questions,
		"created_at": history.CreatedAt,
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "History retrieved", response)
}
