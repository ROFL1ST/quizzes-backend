package config

import (
	"fmt"
	"log"

	"github.com/ROFL1ST/quizzes-backend/models"
)

// Struct sementara untuk menampung data lama dari database
type OldChallengeData struct {
	ID              uint
	ChallengerID    uint
	OpponentID      uint
	ChallengerScore int
	OpponentScore   int
	Status          string // pending, active, finished
}

func MigrateOldChallenges() {
	// 1. Cek apakah tabel challenge_participants masih kosong?
	// Jika sudah ada isinya, kita asumsikan migrasi sudah pernah dijalankan agar tidak duplikat.
	var count int64
	DB.Model(&models.ChallengeParticipant{}).Count(&count)
	if count > 0 {
		fmt.Println("⏩ Challenge migration skipped (data already exists).")
		return
	}

	fmt.Println("🔄 Starting migration for old challenges...")

	// 2. Ambil data lama menggunakan Raw SQL (karena field challenger_id sdh tidak ada di struct Challenge baru)
	var oldChallenges []OldChallengeData
	// Pastikan nama kolom sesuai dengan yang ada di database PostgreSQL Anda saat ini
	err := DB.Raw("SELECT id, challenger_id, opponent_id, challenger_score, opponent_score, status FROM challenges").Scan(&oldChallenges).Error
	if err != nil {
		log.Println("⚠️ Failed to read old challenges (maybe table is empty or columns missing):", err)
		return
	}

	if len(oldChallenges) == 0 {
		fmt.Println("✅ No old challenges to migrate.")
		return
	}

	// 3. Proses setiap data lama
	tx := DB.Begin() // Gunakan transaction agar aman

	for _, old := range oldChallenges {
		// A. Update Header Challenge (Creator & Mode)
		// Kita set Creator = Challenger, Mode = 1v1
		if err := tx.Exec("UPDATE challenges SET creator_id = ?, mode = '1v1', is_realtime = false WHERE id = ?", old.ChallengerID, old.ID).Error; err != nil {
			tx.Rollback()
			log.Fatal("❌ Failed update challenge header:", err)
		}

		// B. Buat Peserta 1: Challenger (Pasti Accepted karena dia pembuat)
		p1 := models.ChallengeParticipant{
			ChallengeID: old.ID,
			UserID:      old.ChallengerID,
			Status:      "accepted",
			Score:       old.ChallengerScore,
		}
		if err := tx.Create(&p1).Error; err != nil {
			tx.Rollback()
			log.Fatal("❌ Failed insert challenger participant:", err)
		}

		// C. Buat Peserta 2: Opponent
		// Status Opponent tergantung status challenge lama
		oppStatus := "pending"
		if old.Status == "active" || old.Status == "finished" {
			oppStatus = "accepted"
		} else if old.Status == "rejected" {
			oppStatus = "rejected"
		}

		p2 := models.ChallengeParticipant{
			ChallengeID: old.ID,
			UserID:      old.OpponentID,
			Status:      oppStatus,
			Score:       old.OpponentScore,
		}
		if err := tx.Create(&p2).Error; err != nil {
			tx.Rollback()
			log.Fatal("❌ Failed insert opponent participant:", err)
		}
	}

	tx.Commit()
	fmt.Println("✅ Success! Migrated", len(oldChallenges), "old challenges to new structure.")
}
