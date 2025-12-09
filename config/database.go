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
	// 1. Coba load .env (Hanya berguna di Lokal, di Vercel ini akan di-skip/error harmless)
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  Note: File .env tidak ditemukan, menggunakan environment system.")
	}

	var dsn string

	// 2. Prioritas 1: Gunakan DATABASE_URL (Neon/Vercel/Cloud)
	if os.Getenv("DATABASE_URL") != "" {
		dsn = os.Getenv("DATABASE_URL")
		fmt.Println("🔍 Menggunakan koneksi via DATABASE_URL...")
	} else {
		// 3. Prioritas 2: Gunakan variabel terpisah (Lokal Docker/PGAdmin)
		// Kita cek dulu apakah DB_HOST ada, untuk menghindari error string kosong
		if os.Getenv("DB_HOST") == "" {
			log.Fatal("❌ Error: DATABASE_URL tidak ada, dan DB_HOST juga kosong! Pastikan konfigurasi database benar.")
		}

		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
		)
		fmt.Println("🔍 Menggunakan koneksi via Variable Host/Port...")
	}

	// 4. Buka Koneksi GORM
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("🔥 Gagal koneksi ke database:", err)
	}

	// 5. Auto Migrate
	err = database.AutoMigrate(
		&models.Role{},
		&models.Admin{},
		&models.User{},
		&models.Friendship{},
		&models.Topic{},
		&models.Quiz{},
		&models.Question{},
		&models.History{},
	)

	if err != nil {
		log.Fatal("❌ Database Migration Failed:", err)
	}

	DB = database
	fmt.Println("🚀 Sukses! Terhubung ke Database.")
}