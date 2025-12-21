package utils

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"math"
	"math/rand"
	"strconv"
	"time"
	"fmt"
)

func GetJakartaTime() time.Time {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {

		loc = time.FixedZone("WIB", 7*60*60)
	}
	return time.Now().In(loc)
}

func StripTime(t time.Time) time.Time {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*60*60)
	}

	t = t.In(loc)
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}

func CheckDailyMissions(userID uint, trigger string, val int, meta string) {
	today := StripTime(GetJakartaTime()) // Waktu Jakarta

	// 1. Ambil Misi Harian User yang Aktif Hari Ini & Belum Diklaim
	var userMissions []models.UserMission
	config.DB.Preload("Mission").
		Where("user_id = ? AND reset_date = ? AND is_claimed = ?", userID, today, false).
		Find(&userMissions)

	for _, um := range userMissions {
		key := um.Mission.Key
		increment := 0

		// 2. Logika Pencocokan Misi
		switch trigger {
		case "quiz":
			// Misi: Mainkan Kuis
			if key == "play_quiz_1" || key == "play_quiz_3" || key == "play_quiz_5" {
				increment = 1
			}
			// Misi: Skor
			if (key == "score_100" || key == "score_100_3x") && val == 100 {
				increment = 1
			}
			if key == "total_score_500" {
				increment = val // Tambah skor
			}
			if key == "perfect_streak" && val >= 10 {
				increment = 10
			}
			// Misi Topik
			if key == "quiz_math" && meta == "Matematika" {
				increment = 1
			}
			if key == "quiz_history" && meta == "Sejarah" {
				increment = 1
			}

		case "login":
			// Cek Jam Login (Waktu Jakarta)
			hour := GetJakartaTime().Hour()
			if key == "login_morning" && hour >= 5 && hour < 10 {
				increment = 1
			}
			if key == "login_night" && hour >= 20 && hour <= 23 {
				increment = 1
			}

		case "shop":
			if key == "buy_item" && meta == "buy" {
				increment = 1
			}
			if key == "equip_avatar" && meta == "equip" {
				increment = 1
			}

		case "social":
			if key == "add_friend" && meta == "add" {
				increment = 1
			}
			if key == "check_leaderboard" && meta == "view" {
				increment = 1
			}
			if key == "share_app" && meta == "share" {
				increment = 1
			}

		case "challenge":
			if key == "win_challenge_1" && meta == "win" {
				increment = 1
			}
			if key == "play_challenge_2v2" && meta == "2v2" {
				increment = 1
			}

		case "level":
			if key == "earn_xp_1000" {
				increment = val
			}

			if key == "level_up_daily" && meta == "levelup" {
				increment = 1
			}
		}

		// 3. Update Progress
		if increment > 0 {
			if um.Progress < um.Mission.Target {
				um.Progress += increment
				if um.Progress > um.Mission.Target {
					um.Progress = um.Mission.Target
				}
				config.DB.Save(&um)

				if um.Progress == um.Mission.Target {
					SendNotification(userID, "success", "Misi Selesai!", "Kamu menyelesaikan misi: "+um.Mission.Title, "/")
				}
			}
		}
	}
}

func AssignDailyMissions(userID uint) {
	today := StripTime(time.Now())

	// Cek apakah user sudah punya misi hari ini?
	var count int64
	config.DB.Model(&models.UserMission{}).
		Where("user_id = ? AND reset_date = ?", userID, today).
		Count(&count)

	if count > 0 {
		return
	} // Sudah ada, skip.

	// Ambil semua misi aktif
	var allMissions []models.Mission
	config.DB.Where("is_active = ?", true).Find(&allMissions)

	if len(allMissions) == 0 {
		return
	}

	// Acak (Shuffle)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(allMissions), func(i, j int) {
		allMissions[i], allMissions[j] = allMissions[j], allMissions[i]
	})

	// Ambil 5 Teratas
	limit := 5
	if len(allMissions) < 5 {
		limit = len(allMissions)
	}

	for i := 0; i < limit; i++ {
		um := models.UserMission{
			UserID:    userID,
			MissionID: allMissions[i].ID,
			ResetDate: today,
			Progress:  0,
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
	// Preload Participants untuk akses data user dan scoring
	if err := config.DB.Preload("Participants.User").First(&challenge, challengeID).Error; err != nil {
		return
	}

	challenge.WinnerID = nil
	challenge.WinningTeam = ""

	// =================================================================
	// 1. LOGIKA PENENTUAN PEMENANG (Sesuai kode Anda)
	// =================================================================
	if challenge.Mode == "2v2" {
		// --- LOGIC 2V2 ---
		scoreA, timeA := 0, 0
		scoreB, timeB := 0, 0
		
		for _, p := range challenge.Participants {
			if p.Status == "accepted" && p.IsFinished {
				if p.Team == "A" {
					scoreA += p.Score
					timeA += p.TimeTaken
				} else if p.Team == "B" {
					scoreB += p.Score
					timeB += p.TimeTaken
				}
			}
		}

		if scoreA > scoreB {
			challenge.WinningTeam = "A"
		} else if scoreB > scoreA {
			challenge.WinningTeam = "B"
		} else {
			if timeA < timeB {
				challenge.WinningTeam = "A"
			} else if timeB < timeA {
				challenge.WinningTeam = "B"
			} else {
				challenge.WinningTeam = "DRAW"
			}
		}

	} else {
		// --- LOGIC BATTLE ROYALE / 1V1 ---
		var winnerID uint = 0
		highestScore := -1
		lowestTime := 999999999

		for _, p := range challenge.Participants {
			// Hanya hitung yang Accepted dan Sudah Selesai
			if p.Status == "accepted" && p.IsFinished {
				if p.Score > highestScore {
					highestScore = p.Score
					lowestTime = p.TimeTaken
					winnerID = p.UserID
				} else if p.Score == highestScore {
					if p.TimeTaken < lowestTime {
						lowestTime = p.TimeTaken
						winnerID = p.UserID
					}
				}
			}
		}

		if winnerID != 0 {
			challenge.WinnerID = &winnerID
		}
	}


	if challenge.WagerAmount > 0 {
		// CASE A: Mode 1v1 (Ada WinnerID)
		if challenge.WinnerID != nil {
			totalPot := challenge.WagerAmount * 2 // Taruhan Saya + Taruhan Lawan
			
			var winner models.User
			if err := config.DB.First(&winner, *challenge.WinnerID).Error; err == nil {
				winner.Coins += totalPot
				config.DB.Save(&winner)

				// Notifikasi Khusus Menang Uang
				msg := fmt.Sprintf("Jackpot! Kamu menang %d koin dari taruhan!", totalPot)
				SendNotification(winner.ID, "success", "Menang Taruhan!", msg, "/shop")
			}
		} 
		
		
		if challenge.Mode == "2v2" && challenge.WinningTeam != "" && challenge.WinningTeam != "DRAW" {
		
			rewardPerPerson := challenge.WagerAmount * 2

			for _, p := range challenge.Participants {
				if p.Team == challenge.WinningTeam && p.Status == "accepted" {
					var member models.User
					if err := config.DB.First(&member, p.UserID).Error; err == nil {
						member.Coins += rewardPerPerson
						config.DB.Save(&member)

						msg := fmt.Sprintf("Tim Menang! Kamu dapat %d koin!", rewardPerPerson)
						SendNotification(member.ID, "success", "Menang Taruhan 2v2!", msg, "/shop")
					}
				}
			}
		}
		
		// CASE C: DRAW (Kembalikan Uang - Optional)
		// Jika DRAW, biasanya uang hangus atau dikembalikan. 
		// Kode di bawah ini opsional untuk refund jika Draw.
		/*
		if (challenge.Mode == "2v2" && challenge.WinningTeam == "DRAW") {
			for _, p := range challenge.Participants {
				if p.Status == "accepted" {
					// Refund logic here
				}
			}
		}
		*/
	}

	// Simpan Perubahan Challenge
	config.DB.Save(&challenge)

	// Broadcast Notif Umum ke Semua Peserta
	for _, p := range challenge.Participants {
		// Hindari spam notif jika pemenang sudah dapat notif khusus di atas
		isWinner := false
		if challenge.WinnerID != nil && *challenge.WinnerID == p.UserID { isWinner = true }
		if challenge.Mode == "2v2" && p.Team == challenge.WinningTeam { isWinner = true }

		if !isWinner {
			SendNotification(p.UserID, "info", "Challenge Selesai", "Lihat siapa pemenangnya!", "/challenges")
		}
	}
}