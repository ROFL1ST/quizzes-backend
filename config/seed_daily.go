package config

import (
	"fmt"
	"github.com/ROFL1ST/quizzes-backend/models"
)

func SeedDailyData() {
	fmt.Println("Seeding Daily Data (100 Days & 20 Missions)...")

	// ==========================================
	// 1. SEED 100 HARI DAILY REWARD
	// ==========================================
	// Pola: Naik bertahap setiap hari, Bonus Besar tiap kelipatan 7
	for day := 1; day <= 100; day++ {
		reward := 0
		
		// Logika Hadiah
		if day % 30 == 0 {
			reward = 500 // SUPER JACKPOT (Hari 30, 60, 90)
		} else if day % 7 == 0 {
			reward = 100 // BIG REWARD (Hari 7, 14, 21...)
		} else {
			// Hari biasa: 10 + (bonus kecil progress mingguan)
			// Contoh Hari 1=10, Hari 2=15, Hari 3=20...
			cycleDay := day % 7
			reward = 10 + (cycleDay * 5) 
		}

		// Simpan ke Database
		// FirstOrCreate agar tidak duplikat kalau dijalankan ulang
		DB.FirstOrCreate(&models.DailyRewardConfig{Day: day}, models.DailyRewardConfig{
			Day:    day,
			Reward: reward,
		})
	}

	// ==========================================
	// 2. SEED 20 MISI HARIAN
	// ==========================================
	missions := []models.Mission{
		// --- Kategori: Main Kuis ---
		{Key: "play_quiz_1", Title: "Pemanasan", Description: "Mainkan 1 Kuis mode apa saja", Target: 1, Reward: 20},
		{Key: "play_quiz_3", Title: "Marathon Kuis", Description: "Mainkan 3 Kuis hari ini", Target: 3, Reward: 50},
		{Key: "play_quiz_5", Title: "Kecanduan", Description: "Mainkan 5 Kuis hari ini", Target: 5, Reward: 100},
		
		// --- Kategori: Skor ---
		{Key: "score_100", Title: "Sempurna", Description: "Dapatkan nilai 100 dalam kuis", Target: 1, Reward: 40},
		{Key: "score_100_3x", Title: "Jenius Sejati", Description: "Dapatkan 3x nilai 100", Target: 3, Reward: 150},
		{Key: "total_score_500", Title: "Pengumpul Poin", Description: "Kumpulkan total 500 skor dari semua kuis", Target: 500, Reward: 60},

		// --- Kategori: Challenge (Duel) ---
		{Key: "win_challenge_1", Title: "Petarung", Description: "Menangkan 1 Challenge lawan teman", Target: 1, Reward: 50},
		{Key: "play_challenge_2v2", Title: "Teamwork", Description: "Mainkan 1 kali mode 2v2", Target: 1, Reward: 30},
		
		// --- Kategori: Waktu Login ---
		{Key: "login_morning", Title: "Semangat Pagi", Description: "Login antara jam 05:00 - 10:00", Target: 1, Reward: 15},
		{Key: "login_night", Title: "Anak Malam", Description: "Login antara jam 20:00 - 24:00", Target: 1, Reward: 15},

		// --- Kategori: Shop & Item ---
		{Key: "buy_item", Title: "Belanja", Description: "Beli 1 item apa saja di Shop", Target: 1, Reward: 25},
		{Key: "equip_avatar", Title: "Gaya Baru", Description: "Ganti/Pasang Avatar Frame", Target: 1, Reward: 10},

		// --- Kategori: XP & Level ---
		{Key: "earn_xp_1000", Title: "Grinding XP", Description: "Dapatkan 1000 XP hari ini", Target: 1000, Reward: 40},
		{Key: "level_up_daily", Title: "Naik Kelas", Description: "Naik 1 Level hari ini", Target: 1, Reward: 200},

		// --- Kategori: Sosial ---
		{Key: "add_friend", Title: "Mencari Teman", Description: "Tambah 1 teman baru", Target: 1, Reward: 20},
		{Key: "check_leaderboard", Title: "Ambis", Description: "Cek halaman Leaderboard", Target: 1, Reward: 5},

		// --- Kategori: Spesial ---
		{Key: "perfect_streak", Title: "Tanpa Salah", Description: "Jawab 10 soal berturut-turut tanpa salah (Logika khusus)", Target: 10, Reward: 100},
		{Key: "quiz_math", Title: "Ahli Hitung", Description: "Mainkan kuis topik Matematika", Target: 1, Reward: 30},
		{Key: "quiz_history", Title: "Sejarawan", Description: "Mainkan kuis topik Sejarah", Target: 1, Reward: 30},
		{Key: "share_app", Title: "Influencer", Description: "Bagikan profilmu (Tombol Share)", Target: 1, Reward: 10},
	}

	for _, m := range missions {
		DB.FirstOrCreate(&models.Mission{Key: m.Key}, m)
	}

	fmt.Println("Seeding Daily Data Done!")
}