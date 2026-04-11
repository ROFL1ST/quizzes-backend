package controllers

import (
	"encoding/json"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CreateHistoryInput struct {
	QuizID              uint            `json:"quiz_id"`
	QuizTitle           string          `json:"quiz_title"`
	Score               int             `json:"score"`
	TotalSoal           int             `json:"total_soal"`
	Snapshot            json.RawMessage `json:"snapshot"`
	TimeTaken           int             `json:"time_taken"`
	ChallengeID         uint            `json:"challenge_id"`
	QuestionIDs         []uint          `json:"question_ids"`
	AssignmentID        *uint           `json:"assignment_id"` // New
	ClassroomID         *uint           `json:"classroom_id"`  // New
	FinalAdaptiveRating *float64        `json:"final_adaptive_rating"`
}

func SaveHistory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)

	var input CreateHistoryInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid data", err.Error())
	}

	// Validasi Deadline Assignment — Tolak jika sudah lewat
	if input.AssignmentID != nil && *input.AssignmentID != 0 {
		var assignment models.Assignment
		if err := config.DB.First(&assignment, *input.AssignmentID).Error; err == nil {
			if assignment.Deadline != "" {
				// Parse deadline (format: "YYYY-MM-DD HH:mm:ss" atau "YYYY-MM-DD")
				layouts := []string{"2006-01-02 15:04:05", "2006-01-02T15:04:05Z", "2006-01-02"}
				var deadline time.Time
				var parsed bool
				for _, layout := range layouts {
					if t, err := time.Parse(layout, assignment.Deadline); err == nil {
						deadline = t
						parsed = true
						break
					}
				}
				if parsed && time.Now().After(deadline) {
					return utils.ErrorResponse(c, fiber.StatusForbidden, "Deadline has passed. This assignment can no longer be submitted.", nil)
				}
			}
		}
	}

	// Init ML Client for AI Grading
	mlURL := os.Getenv("ML_SERVICE_URL")
	if mlURL == "" {
		mlURL = "http://localhost:5002"
	}
	mlClient := utils.NewMLClient(mlURL)
	var essaySubmissions []models.EssaySubmission

	var questions []models.Question
	if input.QuizID != 0 {
		// Kuis Normal
		if err := config.DB.Where("quiz_id = ?", input.QuizID).Find(&questions).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch questions", err.Error())
		}
	} else if len(input.QuestionIDs) > 0 {
		// Remedial (Ambil dari list ID)
		if err := config.DB.Where("id IN ?", input.QuestionIDs).Find(&questions).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch remedial questions", err.Error())
		}
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
	totalPointsEarned := 0.0
	totalQuestions := len(questions)

	// Hitung Benar/Salah
	for qIDStr, answer := range userAnswers {
		qID, _ := strconv.Atoi(qIDStr)
		if q, exists := questionMap[uint(qID)]; exists {
			isCorrect := false

			switch q.Type {
			case "essay":
				// AI Grading (Research Feature)
				aiResp, err := mlClient.GradeEssay(q.QuestionText, q.CorrectAnswer, answer)
				score := 0.0
				feedback := "Error connecting to AI"

				if err == nil && aiResp != nil {
					score = aiResp.ScoreFinal
					feedback = aiResp.Feedback
				} else {
					// Fallback if AI fails: check non-empty
					if len(answer) > 5 {
						score = 50.0 // Give half points for effort if AI down
						feedback = "AI Service Unavailable (Fallback)"
					}
				}

				// Threshold for "Correct" in boolean strict sense (e.g. for progression)
				if score >= 70.0 {
					isCorrect = true
				}

				totalPointsEarned += score

				// Queue for saving later
				essaySubmissions = append(essaySubmissions, models.EssaySubmission{
					QuestionID: q.ID,
					UserAnswer: answer,
					TeacherKey: q.CorrectAnswer,
					AIScore:    score,
					AIFeedback: feedback,
					IsGraded:   true,
				})

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

			if q.Type != "essay" && isCorrect {
				totalPointsEarned += 100.0
			}
		}
	}

	// Hitung Final Score (0-100)
	finalScore := 0
	// Survival Mode (QuizID == 0): Trust Client Score (Streak)
	if input.QuizID == 0 {
		finalScore = input.Score
		totalQuestions = input.TotalSoal
	} else {
		if totalQuestions > 0 {
			finalScore = int(math.Round(totalPointsEarned / float64(totalQuestions)))
		}
	}

	history := models.History{
		UserID:       uint(userID),
		QuizID:       input.QuizID,
		QuizTitle:    input.QuizTitle,
		Score:        finalScore,
		Snapshot:     datatypes.JSON(input.Snapshot),
		TimeTaken:    input.TimeTaken,
		TotalSoal:    totalQuestions,
		AssignmentID: input.AssignmentID, // Save field
		ClassroomID:  input.ClassroomID,  // Save field
	}

	if err := config.DB.Create(&history).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed save history", err.Error())
	}

	// Save Buffered Essay Submissions
	if len(essaySubmissions) > 0 {
		for i := range essaySubmissions {
			essaySubmissions[i].HistoryID = history.ID
			config.DB.Create(&essaySubmissions[i])
		}
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

		// utils.UpdateQuizStreak(&currentUser)
		config.DB.Save(&currentUser)
	}
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

			utils.SendNotification(currentUser.ID, "success", "Naik Level!", "⭐ Level Up! Kamu naik ke Level "+strconv.Itoa(newLevel), "/@"+currentUser.Username)
			utils.CheckDailyMissions(currentUser.ID, "level", 0, "levelup")
		}
		// utils.UpdateQuizStreak(&currentUser)
		config.DB.Save(&currentUser)
	}

	go func() {
		wg.Wait()
		utils.CheckQuizAchievements(history.UserID, finalScore)
	}()
	utils.RecordActivity(uint(userID))
	utils.CheckDailyMissions(currentUser.ID, "quiz", finalScore, history.QuizTitle)
	utils.CheckDailyMissions(currentUser.ID, "level", finalScore, "xp_gain")

	// E. Update User Adaptivity (Sync with DB)
	go func(uid uint, inputRating *float64, score int) {
		var userAdaptivity models.UserAdaptivity
		// Cari atau Buat baru (Default 0.5)
		if err := config.DB.FirstOrCreate(&userAdaptivity, models.UserAdaptivity{UserID: uid}).Error; err != nil {
			return
		}

		// 1. Jika dari Adaptive Mode (Client kirim nilai final)
		if inputRating != nil {
			userAdaptivity.AdaptiveRating = *inputRating
			userAdaptivity.LastDiff = *inputRating
			userAdaptivity.Confidence += 0.05 // Increment confidence
		} else {
			// 2. Jika dari Classic Mode (Reward/Punishment)
			// Hanya update jika confidence masih rendah atau sebagai penyesuaian kecil
			change := 0.0
			if score >= 85 {
				change = 0.02 // Naik dikit
			} else if score <= 40 {
				change = -0.01 // Turun dikit
			}

			newRating := userAdaptivity.AdaptiveRating + change
			if newRating > 1.0 {
				newRating = 1.0
			}
			if newRating < 0.1 {
				newRating = 0.1
			}
			userAdaptivity.AdaptiveRating = newRating
		}
		userAdaptivity.UpdatedAt = time.Now()
		config.DB.Save(&userAdaptivity)
	}(uint(userID), input.FinalAdaptiveRating, finalScore)

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
	if history.QuizID != 0 {
		config.DB.Where("quiz_id = ?", history.QuizID).Find(&questions)
	} else {
		var userAnswers map[string]string
		// Try to unmarshal directly
		if err := json.Unmarshal(history.Snapshot, &userAnswers); err != nil {
			// Failed, maybe it's a JSON string (double encoded)?
			var jsonString string
			if errString := json.Unmarshal(history.Snapshot, &jsonString); errString == nil {
				// It was a string, now unmarshal string content to map
				json.Unmarshal([]byte(jsonString), &userAnswers)
			}
		}

		if len(userAnswers) > 0 {
			var qIDs []uint
			for k := range userAnswers {
				if idInt, err := strconv.Atoi(k); err == nil {
					qIDs = append(qIDs, uint(idInt))
				}
			}
			if len(qIDs) > 0 {
				config.DB.Where("id IN ?", qIDs).Find(&questions)
			}
		}
	}

	// Fetch Essay Submissions (AI Feedback) if any
	var essaySubmissions []models.EssaySubmission
	if err := config.DB.Where("history_id = ?", history.ID).Find(&essaySubmissions).Error; err != nil {
		// Log error but don't fail request
		// fmt.Println("Error fetching essay submissions:", err)
	}

	response := fiber.Map{
		"id":                history.ID,
		"quiz_title":        history.QuizTitle,
		"score":             history.Score,
		"snapshot":          history.Snapshot,
		"time_taken":        history.TimeTaken,
		"questions":         questions,
		"essay_submissions": essaySubmissions, // New field
		"created_at":        history.CreatedAt,
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "History retrieved", response)
}
