package controllers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
	"strconv"
	"time"
)

type CreateChallengeInput struct {
	OpponentUsernames []string `json:"opponent_usernames"`
	QuizID            uint     `json:"quiz_id"`
	Mode              string   `json:"mode"`
	TimeLimit         int      `json:"time_limit"`
	IsRealtime        bool     `json:"is_realtime"`
}

func CreateChallenge(c *fiber.Ctx) error {
	creatorID := c.Locals("user_id").(float64)
	var input CreateChallengeInput

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	// Validasi input khusus 2v2
	if input.Mode == "2v2" && len(input.OpponentUsernames) != 3 {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Mode 2v2 butuh 3 orang (1 teman, 2 lawan)", nil)
	}

	// 1. Buat Header Challenge
	challenge := models.Challenge{
		CreatorID:  uint(creatorID),
		QuizID:     input.QuizID,
		Mode:       input.Mode,
		TimeLimit:  input.TimeLimit,
		IsRealtime: input.IsRealtime,
		Status:     "pending",
	}

	if err := config.DB.Create(&challenge).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed create challenge", err.Error())
	}

	// 2. Masukkan Creator sebagai Peserta (Creator selalu Tim A)
	creatorTeam := "solo"
	if input.Mode == "2v2" {
		creatorTeam = "A"
	}

	creatorPart := models.ChallengeParticipant{
		ChallengeID: challenge.ID,
		UserID:      uint(creatorID),
		Status:      "accepted",
		Team:        creatorTeam, // Set Tim
	}
	config.DB.Create(&creatorPart)

	// 3. Masukkan Lawan/Teman
	if len(input.OpponentUsernames) > 0 {
		var opponents []models.User
		// Pastikan urutan query sesuai urutan input array (workaround sederhana)
		config.DB.Where("username IN ?", input.OpponentUsernames).Find(&opponents)

		// Map user untuk memastikan urutan assign tim sesuai input user
		userMap := make(map[string]models.User)
		for _, u := range opponents {
			userMap[u.Username] = u
		}

		for i, username := range input.OpponentUsernames {
			opp, exists := userMap[username]
			if !exists || opp.ID == uint(creatorID) {
				continue
			}

			team := "solo"
			if input.Mode == "2v2" {
				if i == 0 {
					team = "A" // Teman si Creator
				} else {
					team = "B" // Musuh
				}
			}

			part := models.ChallengeParticipant{
				ChallengeID: challenge.ID,
				UserID:      opp.ID,
				Status:      "pending",
				Team:        team, // Set Tim
			}
			config.DB.Create(&part)

			msg := "⚔️ Kamu ditantang main " + input.Mode + "!"
			if input.Mode == "2v2" && team == "A" {
				msg = "🛡️ Kamu diajak setim main 2v2!"
			}

			// [UPDATED] Send Notification dengan Title
			utils.SendNotification(opp.ID, "warning", "Tantangan Masuk!", msg, "/challenges")
		}
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Challenge created", challenge)
}

func GetMyChallenges(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)

	var participants []models.ChallengeParticipant
	// Ambil ID challenge dimana user terlibat
	config.DB.Where("user_id = ?", userID).Find(&participants)

	var challengeIDs []uint
	for _, p := range participants {
		challengeIDs = append(challengeIDs, p.ChallengeID)
	}

	var challenges []models.Challenge
	if len(challengeIDs) > 0 {
		config.DB.
			Preload("Quiz").
			Preload("Creator").
			Preload("Participants.User"). // Load user detail tiap peserta
			Where("id IN ?", challengeIDs).
			Order("created_at DESC").
			Find(&challenges)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Challenges retrieved", challenges)
}

func AcceptChallenge(c *fiber.Ctx) error {
	id := c.Params("id") // Challenge ID
	userID := c.Locals("user_id").(float64)

	var participant models.ChallengeParticipant
	if err := config.DB.Where("challenge_id = ? AND user_id = ?", id, userID).First(&participant).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "You are not in this challenge", nil)
	}

	if participant.Status != "pending" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Already responded", nil)
	}

	participant.Status = "accepted"
	config.DB.Save(&participant)

	var challenge models.Challenge
	config.DB.Preload("Participants.User").First(&challenge, participant.ChallengeID)

	// A. Jika REALTIME: Jangan auto-start! Trigger SSE update saja.
	if challenge.IsRealtime {
		// Broadcast list player terbaru ke Lobby
		utils.BroadcastLobby(challenge.ID, "player_update", fiber.Map{
			"players": formatParticipants(challenge.Participants),
		})
	} else {
		// B. Jika ASYNC (Battle Royale Biasa): Logic lama (Auto Active)
		if challenge.Status == "pending" {
			challenge.Status = "active"
			config.DB.Save(&challenge)
		}
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Challenge accepted!", nil)
}

func RejectChallenge(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(float64)

	// 1. Ambil data partisipan
	var participant models.ChallengeParticipant
	if err := config.DB.Where("challenge_id = ? AND user_id = ?", id, userID).First(&participant).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "You are not in this challenge", nil)
	}

	// 2. Set status partisipan jadi 'rejected'
	participant.Status = "rejected"
	config.DB.Save(&participant)

	// 3. LOGIC BARU: Cek apakah Challenge harus dibatalkan sepenuhnya?
	var challenge models.Challenge
	if err := config.DB.Preload("Participants").First(&challenge, participant.ChallengeID).Error; err == nil {

		allOpponentsRejected := true

		for _, p := range challenge.Participants {
			// Skip Creator (Host pasti accepted)
			if p.UserID == challenge.CreatorID {
				continue
			}

			if p.Status != "rejected" {
				allOpponentsRejected = false
				break
			}
		}

		if allOpponentsRejected {
			challenge.Status = "rejected"
			config.DB.Save(&challenge)
		}

		if challenge.IsRealtime {
			utils.BroadcastLobby(challenge.ID, "player_update", fiber.Map{
				"players": formatParticipants(challenge.Participants),
			})
		}
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Challenge rejected", nil)
}

func StreamChallengeLobby(c *fiber.Ctx) error {
	idStr := c.Params("id")
	challengeIDData, _ := strconv.Atoi(idStr)
	challengeID := uint(challengeIDData)

	userVal := c.Locals("user_id")
	userID := uint(userVal.(float64))

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	msgChan := utils.AddClientToLobby(challengeID, userID)

	// --- FIX: Kirim Data Awal dengan Format SSE yang Benar ---
	var challenge models.Challenge
	if err := config.DB.Preload("Participants.User").First(&challenge, challengeID).Error; err == nil {
		go func() {
			// Siapkan JSON data pemain
			playersJSON, _ := json.Marshal(fiber.Map{
				"players": formatParticipants(challenge.Participants),
			})

			// Format Manual: event: ... \n data: ... \n\n
			initMsg := fmt.Sprintf("event: player_update\ndata: %s\n\n", string(playersJSON))

			// Kirim ke channel
			msgChan <- initMsg
		}()
	}

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		defer utils.RemoveClientFromLobby(challengeID, userID)

		for {
			select {
			case msg, ok := <-msgChan:
				if !ok {
					return
				}
				// --- FIX: Tulis langsung msg (karena sudah diformat di BroadcastLobby/initMsg) ---
				// Jangan pakai fmt.Fprintf(w, "data: %s\n\n", msg) lagi!
				fmt.Fprint(w, msg)
				w.Flush()

			case <-ticker.C:
				// Keepalive event
				fmt.Fprintf(w, ":keepalive\n\n")
				w.Flush()
			}
		}
	})

	return nil
}

func StartGameRealtime(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := uint(c.Locals("user_id").(float64))

	var challenge models.Challenge
	if err := config.DB.First(&challenge, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Challenge not found", nil)
	}

	// Validasi: Hanya Creator yang boleh start
	if challenge.CreatorID != userID {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Only host can start the game", nil)
	}

	// Validasi: Jangan start kalau sudah active/finished
	if challenge.Status != "pending" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Game already started", nil)
	}

	// 1. Ubah Status DB jadi Active
	challenge.Status = "active"
	config.DB.Save(&challenge)

	// 2. Broadcast Countdown (3 Detik)
	utils.BroadcastLobby(challenge.ID, "start_countdown", fiber.Map{"seconds": 3})

	// 3. Goroutine untuk kirim sinyal 'GO' setelah 3 detik
	go func(chID uint, quizID uint) {
		time.Sleep(3 * time.Second)
		utils.BroadcastLobby(chID, "game_start", fiber.Map{
			"quiz_id": quizID,
			"message": "Game Started!",
		})
	}(challenge.ID, challenge.QuizID)

	return utils.SuccessResponse(c, fiber.StatusOK, "Countdown started", nil)
}

// Helper kecil untuk format data peserta agar rapi di JSON
func formatParticipants(parts []models.ChallengeParticipant) []map[string]interface{} {
	var result []map[string]interface{}
	for _, p := range parts {
		result = append(result, map[string]interface{}{
			"user_id": p.UserID,
			"name":    p.User.Name,
			"status":  p.Status,
		})
	}
	return result
}

type ProgressInput struct {
	CurrentIndex int `json:"current_index"`
	TotalSoal    int `json:"total_soal"`
}

// fungsi untuk mengupdate progress challenge realtime
func UpdateChallengeProgress(c *fiber.Ctx) error {
	id := c.Params("id") // Challenge ID
	challengeIDData, _ := strconv.Atoi(id)
	challengeID := uint(challengeIDData)

	userID := c.Locals("user_id").(float64)

	var user = &models.User{}
	if err := config.DB.First(&user, uint(userID)).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}
	var input ProgressInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", nil)
	}

	// Hitung persentase progress
	percentage := 0
	if input.TotalSoal > 0 {
		percentage = int((float64(input.CurrentIndex) / float64(input.TotalSoal)) * 100)
	}

	utils.BroadcastLobby(challengeID, "opponent_progress", fiber.Map{
		"user_id":  userID,
		"username": user.Username,
		"progress": percentage,
		"index":    input.CurrentIndex,
	})

	return utils.SuccessResponse(c, fiber.StatusOK, "Progress updated", nil)
}
