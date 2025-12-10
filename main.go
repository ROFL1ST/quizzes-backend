package main

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	config.ConnectDB()
	config.SeedDatabase()
	config.SeedExamData()
	config.SeedAchievements()
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://kuis-imk.vercel.app, https://planetpulse-admin-gwcx.vercel.app/, http://localhost:5173/, http://localhost:3000",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	routes.SetupRoutes(app)

	app.Listen(":8000")
}