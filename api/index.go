package handler

import (
	"net/http"

	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/middleware"
	"github.com/ROFL1ST/quizzes-backend/routes"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"log"
)

var app *fiber.App

func init() {
	if err := config.ValidateCriticalConfig(); err != nil {
		log.Fatal(err)
	}

	config.ConnectDB()
	app = fiber.New()
	app.Use(cors.New(config.BuildCORSConfig()))
	app.Use(middleware.EnforceOriginForStateChanging())

	routes.SetupRoutes(app)
}

func Handler(w http.ResponseWriter, r *http.Request) {

	adaptor.FiberApp(app).ServeHTTP(w, r)
}
