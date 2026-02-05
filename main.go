package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

// validateConfig ensures required environment variables are set
func validateConfig() {
	// Load .env only in local development
	if os.Getenv("VERCEL_ENV") == "" {
		_ = godotenv.Load()
	}

	// Validate JWT_SECRET
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("❌ JWT_SECRET environment variable is required")
	}
	if len(jwtSecret) < 32 {
		log.Fatal("❌ JWT_SECRET must be at least 32 characters for security")
	}

	// Validate CORS_ORIGINS
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		log.Println("⚠️ CORS_ORIGINS not set, defaulting to localhost origins")
	}
}

func main() {
	validateConfig()
	config.ConnectDB()
	// config.SeedDatabase()
	// config.SeedExamData()
	// config.SeedAchievements()
	// config.SeedShopItems()
	// config.SeedDailyData()
	// config.MigrateOldChallenges()
	app := fiber.New()

	// Get CORS origins from environment, with a sensible default for local development
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "http://localhost:3000,http://localhost:5173"
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	app.Use(logger.New(logger.Config{
		// Format log custom (optional), defaultnya juga sudah bagus
		Format:     "[${time}] ${status} - ${latency} ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Asia/Jakarta",
	}))

	routes.SetupRoutes(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Fatal(app.Listen(":" + port))
}
