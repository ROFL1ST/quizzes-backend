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
	// Load .env file (abaikan error jika file tidak ada, misal di production)
	godotenv.Load()

	var dsn string

	if os.Getenv("DATABASE_URL") != "" {
		dsn = os.Getenv("DATABASE_URL")
	} else {

		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
		)
	}

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

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
		log.Fatal("Database Migration Failed:", err)
	}

	DB = database
	fmt.Println("🚀 Connected to Neon/Vercel Postgres Successfully!")
}
