package handler

import (
	"net/http"

	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/routes"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var app *fiber.App

func init() {
	config.ConnectDB()
	app = fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Atau sesuaikan dengan domain frontend kamu
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	routes.SetupRoutes(app)
}

func Handler(w http.ResponseWriter, r *http.Request) {

	adaptor.FiberApp(app).ServeHTTP(w, r)
}
