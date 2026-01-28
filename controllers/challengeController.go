package controllers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
)

type CreateChallengeInput struct {
	OpponentUsernames []string `json:"opponent_usernames"`
	QuizID            uint     `json:"quiz_id"`
	Mode              string   `json:"mode"`
	TimeLimit         int      `json:"time_limit"`
	IsRealtime        bool     `json:"is_realtime"`
	WagerAmount       int      `json:"wager_amount"`
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

	// Validasi Survival (Optional)
	if input.Mode == "survival" {
		// Survival bisa 1v1 atau banyak. QuizID diabaikan (Random Global).
		if input.QuizID == 0 {
			// Opsional: Set placeholder ID jika perlu
		}
	}

	// --- [LOGIC BARU] Cek Saldo & Potong Taruhan Creator ---
	if input.WagerAmount > 0 {
		var creator models.User
		if err := config.DB.First(&creator, uint(creatorID)).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
		}

		if creator.Coins < input.WagerAmount {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Koin tidak cukup untuk taruhan!", nil)
		}

		// Potong koin creator sekarang
		creator.Coins -= input.WagerAmount
		if err := config.DB.Save(&creator).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memproses koin", err.Error())
		}
	}
	// --------------------------------------------------------

	// Prepare QuizID pointer (handle nullable)
	var quizIDPtr *uint
	if input.QuizID != 0 {
		quizIDPtr = &input.QuizID
	}

	// 1. Buat Header Challenge
	challenge := models.Challenge{
		CreatorID:   uint(creatorID),
		QuizID:      quizIDPtr,
		Mode:        input.Mode,
		TimeLimit:   input.TimeLimit,
		IsRealtime:  input.IsRealtime,
		Status:      "pending",
		WagerAmount: input.WagerAmount, // <-- Simpan nilai taruhan
		RoomCode:    generateRoomCode(),
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
		Team:        creatorTeam,
	}
	config.DB.Create(&creatorPart)

	// 3. Masukkan Lawan/Teman
	if len(input.OpponentUsernames) > 0 {
		var opponents []models.User
		config.DB.Where("username IN ?", input.OpponentUsernames).Find(&opponents)

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
					team = "A"
				} else {
					team = "B"
				}
			}

			part := models.ChallengeParticipant{
				ChallengeID: challenge.ID,
				UserID:      opp.ID,
				Status:      "pending",
				Team:        team,
			}
			config.DB.Create(&part)

			msg := "⚔️ Kamu ditantang main " + input.Mode + "!"
			if input.WagerAmount > 0 {
				msg += fmt.Sprintf(" (Taruhan: %d Koin)", input.WagerAmount)
			}
			if input.Mode == "2v2" && team == "A" {
				msg = "🛡️ Kamu diajak setim main 2v2!"
			}

			utils.SendNotification(opp.ID, "warning", "Tantangan Masuk!", msg, "/challenges")
		}
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Challenge created", challenge)
}

// controllers/socialController.go

func GetMyChallenges(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	var challenges []models.Challenge

	// 1. QUERY UTAMA (Fix Table Name: participants -> challenge_participants)
	err := config.DB.
		Preload("Creator").
		Preload("Quiz").
		Preload("Participants.User").
		// Ganti 'participants' dengan 'challenge_participants'
		Joins("JOIN challenge_participants ON challenge_participants.challenge_id = challenges.id").
		Where("challenge_participants.user_id = ? AND challenge_participants.deleted_at IS NULL", userID).
		Order("challenges.updated_at DESC").
		Limit(limit).
		Offset(offset).
		Distinct("challenges.id", "challenges.*").
		Find(&challenges).Error

	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch challenges", err.Error())
	}

	stats := fiber.Map{
		"total":    0,
		"wins":     0,
		"win_rate": 0,
	}

	// 2. HITUNG STATISTIK (Fix Table Name disini juga)
	if page == 1 {
		var totalPlayed int64
		var totalWins int64

		// A. Total Main
		config.DB.Table("challenges").
			Joins("JOIN challenge_participants ON challenge_participants.challenge_id = challenges.id").
			Where("challenge_participants.user_id = ? AND challenges.status = ? AND challenge_participants.deleted_at IS NULL", userID, "finished").
			Count(&totalPlayed)

		// B. Total Menang
		config.DB.Table("challenges").
			Joins("JOIN challenge_participants ON challenge_participants.challenge_id = challenges.id").
			Where("challenge_participants.user_id = ? AND challenges.status = ? AND challenge_participants.deleted_at IS NULL", userID, "finished").
			Where("(challenges.winner_id = ? OR (challenges.mode = '2v2' AND challenges.winning_team = challenge_participants.team))", userID).
			Count(&totalWins)

		winRate := 0.0
		if totalPlayed > 0 {
			winRate = float64(totalWins) / float64(totalPlayed) * 100
		}

		stats["total"] = totalPlayed
		stats["wins"] = totalWins
		stats["win_rate"] = int(winRate)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Challenges retrieved", fiber.Map{
		"list":     challenges,
		"stats":    stats,
		"has_more": len(challenges) == limit,
	})
}

func AcceptChallenge(c *fiber.Ctx) error {
	id := c.Params("id") // Challenge ID
	userID := c.Locals("user_id").(float64)

	// 1. Ambil data partisipan
	var participant models.ChallengeParticipant
	if err := config.DB.Where("challenge_id = ? AND user_id = ?", id, userID).First(&participant).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "You are not in this challenge", nil)
	}

	if participant.Status != "pending" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Already responded", nil)
	}

	// 2. Ambil Data Challenge untuk cek WagerAmount
	var challenge models.Challenge
	if err := config.DB.First(&challenge, participant.ChallengeID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Challenge data missing", nil)
	}

	// --- [LOGIC BARU] Cek Saldo Penantang ---
	if challenge.WagerAmount > 0 {
		var user models.User
		if err := config.DB.First(&user, uint(userID)).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "User data error", nil)
		}

		if user.Coins < challenge.WagerAmount {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Koin kamu kurang untuk menerima taruhan ini!", nil)
		}

		// Potong koin penantang
		user.Coins -= challenge.WagerAmount
		if err := config.DB.Save(&user).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memproses pembayaran", err.Error())
		}
	}
	// ----------------------------------------

	// 3. Update Status jadi Accepted
	participant.Status = "accepted"
	config.DB.Save(&participant)

	// 4. Load ulang challenge dengan preload peserta (untuk broadcast)
	config.DB.Preload("Participants.User").First(&challenge, participant.ChallengeID)

	// A. Jika REALTIME
	if challenge.IsRealtime {
		utils.BroadcastLobby(challenge.ID, "player_update", fiber.Map{
			"players":    formatParticipants(challenge.Participants),
			"creator_id": challenge.CreatorID,
		})
	} else {
		// B. Jika ASYNC
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
				"players":    formatParticipants(challenge.Participants),
				"creator_id": challenge.CreatorID,
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
				"players":    formatParticipants(challenge.Participants),
				"creator_id": challenge.CreatorID,
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

	// Preload Participants to get their IDs
	var challenge models.Challenge
	if err := config.DB.Preload("Participants").First(&challenge, id).Error; err != nil {
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

	// --- LOGIC BARU: Non-Realtime / Santai ---
	if !challenge.IsRealtime {
		// Broadcast Update Status ke Lobby (supaya yang lagi nunggu tau)
		utils.BroadcastLobby(challenge.ID, "player_update", fiber.Map{
			"players":    formatParticipants(challenge.Participants),
			"creator_id": challenge.CreatorID,
			"status":     "active",
		})

		return utils.SuccessResponse(c, fiber.StatusOK, "Game active", fiber.Map{"status": "active"})
	}

	// Generate SEED jika Survival Mode
	seed := ""
	if challenge.Mode == "survival" {
		seed = fmt.Sprintf("%d-%d", time.Now().UnixNano(), challenge.ID)
	}

	// Safely dereference QuizID
	var quizIDVal uint
	if challenge.QuizID != nil {
		quizIDVal = *challenge.QuizID
	}

	// 2. Broadcast Countdown (3 Detik)
	utils.BroadcastLobby(challenge.ID, "start_countdown", fiber.Map{
		"seconds": 3,
		"mode":    challenge.Mode, // Biar frontend tau
		"seed":    seed,           // Kirim seed di awal (opsional) atau pas start
	})

	// Calculate Min Adaptive Rating
	initialDiff := 0.5
	if challenge.Mode == "survival" || challenge.Mode == "1v1" {
		minR := 1.0
		found := false
		for _, p := range challenge.Participants {
			if p.Status != "accepted" {
				continue
			}
			var ua models.UserAdaptivity
			if err := config.DB.First(&ua, p.UserID).Error; err == nil {
				if ua.AdaptiveRating < minR {
					minR = ua.AdaptiveRating
				}
				found = true
			}
		}
		if found {
			initialDiff = minR
		}
	}

	// 3. Goroutine untuk kirim sinyal 'GO' setelah 3 detik
	go func(chID uint, quizID uint, chMode string, chSeed string, initDiff float64) {
		time.Sleep(3 * time.Second)
		utils.BroadcastLobby(chID, "game_start", fiber.Map{
			"quiz_id":            quizID,
			"message":            "Game Started!",
			"seed":               chSeed,
			"mode":               chMode,
			"initial_difficulty": initDiff, // Send min rating
		})
	}(challenge.ID, quizIDVal, challenge.Mode, seed, initialDiff)

	return utils.SuccessResponse(c, fiber.StatusOK, "Countdown started", nil)
}

func GenerateRoomCode(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := uint(c.Locals("user_id").(float64))

	var challenge models.Challenge
	if err := config.DB.First(&challenge, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Challenge not found", nil)
	}

	if challenge.CreatorID != userID {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Only host can generate code", nil)
	}

	if challenge.RoomCode != "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Room code already exists", fiber.Map{"room_code": challenge.RoomCode})
	}

	challenge.RoomCode = generateRoomCode()
	config.DB.Save(&challenge)

	// Broadcast update settings (include room code)
	utils.BroadcastLobby(challenge.ID, "settings_update", fiber.Map{
		"mode":         challenge.Mode,
		"time_limit":   challenge.TimeLimit,
		"quiz_id":      challenge.QuizID,
		"is_realtime":  challenge.IsRealtime,
		"wager_amount": challenge.WagerAmount,
		"room_code":    challenge.RoomCode,
	})

	return utils.SuccessResponse(c, fiber.StatusOK, "Room code generated", fiber.Map{"room_code": challenge.RoomCode})
}

// Helper kecil untuk format data peserta agar rapi di JSON
func formatParticipants(parts []models.ChallengeParticipant) []map[string]interface{} {
	var result []map[string]interface{}
	for _, p := range parts {
		result = append(result, map[string]interface{}{
			"user_id":     p.UserID,
			"name":        p.User.Name,
			"status":      p.Status,
			"team":        p.Team,
			"score":       p.Score,      // NEW: Send score
			"is_finished": p.IsFinished, // NEW: Send finish status
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

func LeaveLobby(c *fiber.Ctx) error {
	id := c.Params("id")
	challengeIDData, _ := strconv.Atoi(id)
	challengeID := uint(challengeIDData)
	userID := c.Locals("user_id").(float64)

	// 1. Ambil Partisipan
	var participant models.ChallengeParticipant
	if err := config.DB.Where("challenge_id = ? AND user_id = ?", challengeID, userID).First(&participant).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "You are not in this challenge", nil)
	}

	// 2. Hapus dari lobby (Soft Delete)
	// Ini memungkinkan user untuk join kembali jika mereka mau (karena record dianggap hilang)
	if err := config.DB.Delete(&participant).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to leave lobby", err.Error())
	}

	// 3. HOST MIGRATION LOGIC
	var challenge models.Challenge
	if err := config.DB.Preload("Participants.User").First(&challenge, challengeID).Error; err == nil {

		// Jika yang keluar adalah Host
		if challenge.CreatorID == uint(userID) {
			var newHost *models.ChallengeParticipant

			// Cari kandidat host baru (Oldest joint time, status accepted/pending)
			for _, p := range challenge.Participants {
				if p.UserID == uint(userID) || p.Status == "rejected" {
					continue
				}
				// Karena slice participants biasanya preloaded default order (atau kita bisa assume ID created at)
				// Kita ambil yang pertama ketemu yang valid
				newHost = &p
				break
			}

			if newHost != nil {
				// Transfer Host Authority
				challenge.CreatorID = newHost.UserID
				config.DB.Save(&challenge)

				// Broadcast Notif Host Migration
				utils.BroadcastLobby(challenge.ID, "host_migration", fiber.Map{
					"new_host_id": newHost.UserID,
					"message":     "Host has left. New host is " + newHost.User.Name,
				})
			} else {
				// Lobby kosong melompong -> Cancel Challenge
				challenge.Status = "cancelled"
				config.DB.Save(&challenge)
			}
		}

		// Broadcast ke lobby bahwa ada update player
		// Reload challenge untuk memastikan data participants terbaru (jika perlu)
		config.DB.Preload("Participants.User").First(&challenge, challengeID)
		utils.BroadcastLobby(challenge.ID, "player_update", fiber.Map{
			"players":    formatParticipants(challenge.Participants),
			"creator_id": challenge.CreatorID,
			"status":     challenge.Status,
		})
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Left lobby successfully", nil)
}

func generateRoomCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// Handler baru: Join by Room Code
type JoinByCodeInput struct {
	RoomCode string `json:"room_code"`
}

func JoinChallengeByCode(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	var input JoinByCodeInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	var challenge models.Challenge
	if err := config.DB.Preload("Participants").Where("room_code = ?", input.RoomCode).First(&challenge).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Room not found", nil)
	}

	if challenge.Status != "pending" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Game already started or finished", nil)
	}

	// Cek apakah user pernah join (termasuk soft deleted)
	var existingPart models.ChallengeParticipant
	if err := config.DB.Unscoped().Where("challenge_id = ? AND user_id = ?", challenge.ID, userID).First(&existingPart).Error; err == nil {
		// Jika user sudah ada (tapi mungkin soft deleted atau left)
		if existingPart.DeletedAt.Valid {
			// Revive & Reset Status
			config.DB.Unscoped().Model(&existingPart).Updates(map[string]interface{}{
				"deleted_at": nil,
				"status":     "accepted",
				"team":       "solo", // Reset team if needed
			})
		} else {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "You are already in this lobby", nil)
		}
	} else {
		// New Joiner
		part := models.ChallengeParticipant{
			ChallengeID: challenge.ID,
			UserID:      uint(userID),
			Status:      "accepted",
			Team:        "solo",
		}
		if challenge.Mode == "2v2" {
			part.Team = "B"
		}
		if err := config.DB.Create(&part).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to join", err.Error())
		}
	}

	// Broadcast ke Lobby
	config.DB.Preload("Participants.User").First(&challenge, challenge.ID)
	utils.BroadcastLobby(challenge.ID, "player_update", fiber.Map{
		"players":    formatParticipants(challenge.Participants),
		"creator_id": challenge.CreatorID,
	})

	return utils.SuccessResponse(c, fiber.StatusOK, "Joined lobby", challenge)
}

type UpdateSettingsInput struct {
	Mode        string `json:"mode"`
	TimeLimit   *int   `json:"time_limit"` // Pointer agar bisa deteksi 0
	QuizID      uint   `json:"quiz_id"`
	IsRealtime  *bool  `json:"is_realtime"`
	WagerAmount *int   `json:"wager_amount"` // Pointer for 0 or value
}

func UpdateLobbySettings(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := uint(c.Locals("user_id").(float64))

	var input UpdateSettingsInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	var challenge models.Challenge
	if err := config.DB.First(&challenge, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Challenge not found", nil)
	}

	if challenge.CreatorID != userID {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Only host can update settings", nil)
	}

	// VALIDASI: Tidak boleh ubah setting kalau sudah active
	if challenge.Status != "pending" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Cannot update settings while game is active", nil)
	}

	// Update Fields
	if input.Mode != "" {
		challenge.Mode = input.Mode
	}
	if input.TimeLimit != nil {
		challenge.TimeLimit = *input.TimeLimit
	}
	if input.QuizID != 0 {
		qID := input.QuizID
		challenge.QuizID = &qID
	}
	if input.IsRealtime != nil {
		challenge.IsRealtime = *input.IsRealtime
	}
	if input.WagerAmount != nil {
		challenge.WagerAmount = *input.WagerAmount
	}

	config.DB.Save(&challenge)

	// Broadcast update settings
	utils.BroadcastLobby(challenge.ID, "settings_update", fiber.Map{
		"mode":         challenge.Mode,
		"time_limit":   challenge.TimeLimit,
		"quiz_id":      challenge.QuizID,
		"is_realtime":  challenge.IsRealtime,
		"wager_amount": challenge.WagerAmount,
	})

	return utils.SuccessResponse(c, fiber.StatusOK, "Settings updated", nil)
}

// InviteToLobby sends a notification to a user to join the lobby
func InviteToLobby(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := uint(c.Locals("user_id").(float64))

	var input struct {
		Username string `json:"username"`
	}
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", nil)
	}

	var challenge models.Challenge
	if err := config.DB.Preload("Participants").First(&challenge, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Challenge not found", nil)
	}

	// Cek apakah pengundang adalah anggota lobby (semua member boleh invite)
	isMember := false
	for _, p := range challenge.Participants {
		if p.UserID == userID && p.Status != "rejected" {
			isMember = true
			break
		}
	}
	if !isMember {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "You are not in this lobby", nil)
	}

	var target models.User
	if err := config.DB.Where("username = ?", input.Username).First(&target).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	// Cek apakah target sudah ada di lobby
	for _, p := range challenge.Participants {
		if p.UserID == target.ID {
			if p.Status == "rejected" {
				// Boleh invite lagi yg udah keluar
			} else {
				return utils.ErrorResponse(c, fiber.StatusBadRequest, "User already in lobby", nil)
			}
		}
	}

	// Create Participant Entry (Pending) so they see it in list?
	// Actually, usually Invite just sends Notif. They join via Notification Click -> /challenges -> Accept.
	// So we just add them as "pending" participant if not exists?
	// Or just send Notification with Room Code?

	// Existing logic in CreateChallenge adds them as pending. Let's do that.
	// Check if exists (including soft deleted)
	var part models.ChallengeParticipant
	err := config.DB.Unscoped().Where("challenge_id = ? AND user_id = ?", challenge.ID, target.ID).First(&part).Error

	if err == nil {
		// Exists
		if part.Status == "rejected" || part.DeletedAt.Valid {
			// Revive / Reset
			config.DB.Unscoped().Model(&part).Updates(map[string]interface{}{
				"deleted_at": nil,
				"status":     "pending",
				"team":       "solo", // Default
			})
		}
	} else {
		// Create new
		part = models.ChallengeParticipant{
			ChallengeID: challenge.ID,
			UserID:      target.ID,
			Status:      "pending",
			Team:        "solo",
		}
		if challenge.Mode == "2v2" {
			part.Team = "solo" // Host can assign later or auto
		}
		config.DB.Create(&part)
	}

	// Send Notification
	msg := fmt.Sprintf("📩 Kamu diundang masuk Lobby %s (%s)!", challenge.Mode, challenge.RoomCode)
	utils.SendNotification(target.ID, "info", "Undangan Lobby", msg, "/challenges")

	// Broadcast Update to Lobby (so everyone sees the pending player)
	config.DB.Preload("Participants.User").First(&challenge, challenge.ID)
	utils.BroadcastLobby(challenge.ID, "player_update", fiber.Map{
		"players":    formatParticipants(challenge.Participants),
		"creator_id": challenge.CreatorID,
	})

	return utils.SuccessResponse(c, fiber.StatusOK, "Invitation sent", nil)
}
