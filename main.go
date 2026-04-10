package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/middleware"
	"github.com/ROFL1ST/quizzes-backend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {

	if os.Getenv("VERCEL_ENV") == "" {
		_ = godotenv.Load()
	}

	if len(os.Getenv("JWT_SECRET")) < 32 {
		log.Println("⚠️ WARNING: JWT_SECRET is missing or too short!")
		if os.Getenv("GO_ENV") == "production" {
			log.Fatal("❌ Cannot start in production with weak JWT_SECRET")
		}
	}

	if os.Getenv("API_KEY") == "" {
		log.Fatal("❌ API_KEY must be set in environment variables")
	}

	config.ConnectDB()

	app := fiber.New(fiber.Config{
		ProxyHeader: "X-Forwarded-For",
	})

	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:5173,http://localhost:3000" // Default dev
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-CSRF-Token, X-API-KEY",
		AllowCredentials: true,
	}))

	// 2. Security Middleware
	app.Use(helmet.New())
	app.Use(middleware.StrictOriginMiddleware()) // CSRF Protection via Origin Check
	app.Use(middleware.XApiKeyMiddleware())

	// 3. Rate Limiting
	app.Use("/api/login", middleware.AuthLimiter)
	app.Use("/api/register", middleware.RegisterLimiter)

	// Optional: General limiter for all other routes
	// app.Use(middleware.GeneralLimiter)

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
