package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/middleware"
	"github.com/ROFL1ST/quizzes-backend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	if err := config.ValidateCriticalConfig(); err != nil {
		log.Fatal(err)
	}

	config.ConnectDB()
	// config.SeedDatabase()
	// config.SeedExamData()
	// config.SeedAchievements()
	// config.SeedShopItems()
	// config.SeedDailyData()
	// config.MigrateOldChallenges()
	app := fiber.New()

	app.Use(cors.New(config.BuildCORSConfig()))
	app.Use(middleware.EnforceOriginForStateChanging())

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
