package config

import (
	"fmt"
	"log"
	"os"

	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	// Load .env hanya saat lokal (di Vercel file .env tidak ada)
	if os.Getenv("VERCEL_ENV") == "" {
		_ = godotenv.Load()
	}

	var dsn string

	// === PRIORITAS 1 : DATABASE_URL ===
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		dsn = dbURL
		fmt.Println("🔍 Using DATABASE_URL")
	} else {

		// Pastikan minimal DB_HOST tersedia
		host := os.Getenv("DB_HOST")
		if host == "" {
			log.Fatal("❌ DATABASE_URL kosong, dan DB_HOST kosong! Environment variable tidak terbaca!")
		}

		sslMode := os.Getenv("DB_SSLMODE")
		if sslMode == "" {
			sslMode = "require"
		}

		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
			host,
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
			sslMode,
		)

		fmt.Println("🔍 Using separated DB_* variables")
	}

	// === CONNECT TO DB ===
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: false,
	})
	if err != nil {
		log.Fatal("🔥 Gagal konek DB:", err)
	}

	// === MIGRATIONS ===
	err = database.AutoMigrate(
		&models.Role{},
		&models.Admin{},
		&models.User{},
		&models.PasswordReset{},
		&models.Friendship{},
		&models.Topic{},
		&models.Quiz{},
		&models.Question{},
		&models.QuestionAnalysis{},
		&models.History{},
		&models.Achievement{},
		&models.Activity{},
		&models.Challenge{},
		&models.ChallengeParticipant{},
		&models.UserAchievement{},
		&models.SystemConfig{},
		&models.Notification{},
		&models.Item{},
		&models.UserItem{},
		&models.DailyClaim{},
		&models.DailyRewardConfig{},
		&models.Mission{},
		&models.UserMission{},
		&models.StreakLog{},
		&models.Report{},
		&models.QuizReview{},
		&models.Classroom{},
		&models.ClassroomMember{},
		&models.Assignment{},
		&models.Announcement{},
		&models.Translation{},
		&models.Translation{},
		&models.UserAdaptivity{},
		&models.EssaySubmission{},
	)

	if err != nil {
		log.Fatal("❌ Gagal migrate database:", err)
	}

	DB = database
	fmt.Println("🚀 Database Connected Successfully!")
}
