package utils

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"math"
	"strconv"
	"time"
	"math/rand"
)

func StripTime(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

func AssignDailyMissions(userID uint) {
	today := StripTime(time.Now())

	// Cek apakah user sudah punya misi hari ini?
	var count int64
	config.DB.Model(&models.UserMission{}).
		Where("user_id = ? AND reset_date = ?", userID, today).
		Count(&count)

	if count > 0 { return } // Sudah ada, skip.

	// Ambil semua misi aktif
	var allMissions []models.Mission
	config.DB.Where("is_active = ?", true).Find(&allMissions)

	if len(allMissions) == 0 { return }

	// Acak (Shuffle)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(allMissions), func(i, j int) {
		allMissions[i], allMissions[j] = allMissions[j], allMissions[i]
	})

	// Ambil 5 Teratas
	limit := 5
	if len(allMissions) < 5 { limit = len(allMissions) }
	
	for i := 0; i < limit; i++ {
		um := models.UserMission{
			UserID: userID,
			MissionID: allMissions[i].ID,
			ResetDate: today,
			Progress: 0,
		}
		config.DB.Create(&um)
	}
}

func UnlockAchievement(userID uint, achievementID uint) {
	var count int64
	config.DB.Model(&models.UserAchievement{}).
		Where("user_id = ? AND achievement_id = ?", userID, achievementID).
		Count(&count)

	if count > 0 {
		return
	}

	ua := models.UserAchievement{
		UserID:        userID,
		AchievementID: achievementID,
		UnlockedAt:    time.Now(),
	}
	if err := config.DB.Create(&ua).Error; err != nil {
		return
	}

	var ach models.Achievement
	config.DB.First(&ach, achievementID)

	// 1. Ambil data User untuk mendapatkan Username
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return // Handle jika user tidak ditemukan (opsional)
	}

	activity := models.Activity{
		UserID:      userID,
		Type:        "achievement",
		Description: "Membuka Achievement: " + ach.Name,
	}
	config.DB.Create(&activity)

	// 2. [UPDATED] Send Notification dengan Title
	// Format: UserID, Type, Title, Message, Link
	SendNotification(userID, "success", "Achievement Unlocked!", "🏆 Selamat! Kamu membuka: "+ach.Name, "/@"+user.Username)
}

func CheckQuizAchievements(userID uint, score int) {
	UnlockAchievement(userID, 1)

	if score == 100 {
		UnlockAchievement(userID, 2)
	}

	if score < 50 {
		UnlockAchievement(userID, 11)
	}

	var totalKuis int64
	config.DB.Model(&models.History{}).Where("user_id = ?", userID).Count(&totalKuis)
	if totalKuis >= 10 {
		UnlockAchievement(userID, 3)
	}

	currentHour := time.Now().Hour()
	if currentHour >= 0 && currentHour < 5 {
		UnlockAchievement(userID, 12)
	}

	var user models.User
	config.DB.First(&user, userID)

	if user.Level >= 5 {
		UnlockAchievement(userID, 5)
	}

	if user.Level >= 10 {
		UnlockAchievement(userID, 13)
	}

	if user.XP >= 5000 {
		UnlockAchievement(userID, 10)
	}

	if user.StreakCount >= 3 {
		UnlockAchievement(userID, 6)
	}

	hoursSinceJoin := time.Since(user.CreatedAt).Hours()
	if hoursSinceJoin >= (24 * 30) {
		UnlockAchievement(userID, 14)
	}

	if score == 100 {
		var perfectQuizzes int64
		config.DB.Model(&models.History{}).
			Where("user_id = ? AND score = 100", userID).
			Distinct("quiz_id").
			Count(&perfectQuizzes)

		if perfectQuizzes >= 3 {
			UnlockAchievement(userID, 15)
		}
	}

	var distinctTopics int64
	config.DB.Table("histories").
		Joins("JOIN quizzes ON quizzes.id = histories.quiz_id").
		Where("histories.user_id = ?", userID).
		Distinct("quizzes.topic_id").
		Count(&distinctTopics)

	if distinctTopics >= 3 {
		UnlockAchievement(userID, 8)
	}

	var totalWins int64
	config.DB.Model(&models.Challenge{}).Where("winner_id = ?", userID).Count(&totalWins)

	if totalWins >= 1 {
		UnlockAchievement(userID, 4)
	}

	if totalWins >= 5 {
		UnlockAchievement(userID, 9)
	}
}

func GetLevelingFactor() float64 {
	var conf models.SystemConfig
	config.DB.Where("key = ?", "leveling_factor").Find(&conf)

	if conf.Value == "" {
		return 100.0
	}

	val, err := strconv.ParseFloat(conf.Value, 64)
	if err != nil {
		return 100.0
	}
	return val
}

func CalculateLevel(xp int64) int {
	factor := GetLevelingFactor()
	return int(math.Sqrt(float64(xp)/factor)) + 1
}

func CalculateMinXPForLevel(level int) int64 {
	if level <= 1 {
		return 0
	}
	factor := GetLevelingFactor()
	return int64(factor * math.Pow(float64(level-1), 2))
}

func DetermineWinner(challengeID uint) {
	var challenge models.Challenge
	// Load participants untuk hitung skor
	if err := config.DB.Preload("Participants").First(&challenge, challengeID).Error; err != nil {
		return
	}

	// Reset dulu status pemenang
	challenge.WinnerID = nil
	challenge.WinningTeam = ""

	if challenge.Mode == "2v2" {
		// --- LOGIC 2V2 (TEAM BASED) ---
		scoreA, timeA := 0, 0
		scoreB, timeB := 0, 0

		for _, p := range challenge.Participants {
			// Hanya hitung peserta yang main (Accepted & Score valid)
			if p.Status != "accepted" || p.Score == -1 {
				continue
			}

			if p.Team == "A" {
				scoreA += p.Score
				timeA += p.TimeTaken
			} else if p.Team == "B" {
				scoreB += p.Score
				timeB += p.TimeTaken
			}
		}

		// Bandingkan Skor Tim Total
		if scoreA > scoreB {
			challenge.WinningTeam = "A"
		} else if scoreB > scoreA {
			challenge.WinningTeam = "B"
		} else {
			// Jika Seri Skor, Cek Waktu Total (Lebih cepat/kecil menang)
			if timeA < timeB {
				challenge.WinningTeam = "A"
			} else if timeB < timeA {
				challenge.WinningTeam = "B"
			} else {
				challenge.WinningTeam = "DRAW"
			}
		}

	} else {
		// --- LOGIC BATTLE ROYALE & 1V1 (INDIVIDUAL) ---
		// Pemenang ditentukan berdasarkan:
		// 1. Poin Tertinggi (Score)
		// 2. Waktu Tercepat (TimeTaken) - Jika poin sama

		var winnerID uint = 0
		highestScore := -1
		lowestTime := 999999999 // Inisialisasi dengan angka besar

		for _, p := range challenge.Participants {
			// Hanya hitung peserta yang main
			if p.Status != "accepted" || p.Score == -1 {
				continue
			}

			// Cek Poin
			if p.Score > highestScore {
				// Found new highest score
				highestScore = p.Score
				lowestTime = p.TimeTaken
				winnerID = p.UserID
			} else if p.Score == highestScore {
				// Jika Poin sama, cek siapa lebih cepat (TimeTaken lebih kecil)
				if p.TimeTaken < lowestTime {
					lowestTime = p.TimeTaken
					winnerID = p.UserID
				}
			}
		}

		// Simpan ID Pemenang jika ada
		if winnerID != 0 {
			challenge.WinnerID = &winnerID
		}
	}

	// Simpan Hasil ke Database
	config.DB.Save(&challenge)
	
	// [UPDATED] Kirim Notifikasi ke user bahwa challenge selesai
	for _, p := range challenge.Participants {
		SendNotification(p.UserID, "info", "Challenge Selesai", "Kompetisi telah berakhir. Cek hasilnya sekarang!", "/challenges")
	}
}