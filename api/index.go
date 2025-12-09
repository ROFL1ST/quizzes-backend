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

// Fungsi init berjalan sekali saat "Cold Start" (Serverless dinyalakan)
func init() {
	// 1. Konek Database (Pastikan Env Var sudah diset di Vercel Dashboard)
	config.ConnectDB()
    
    // Opsional: Seed data bisa dimatikan di production agar tidak lambat/duplicate
    // config.SeedDatabase() 

	// 2. Init Fiber
	app = fiber.New()

	// 3. CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Atau sesuaikan dengan domain frontend kamu
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// 4. Setup Routes
	routes.SetupRoutes(app)
}

// Handler adalah fungsi utama yang dicari Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	// Adaptor mengubah request Vercel menjadi request Fiber
	adaptor.FiberApp(app).ServeHTTP(w, r)
}