package routes

import (
	"github.com/ROFL1ST/quizzes-backend/controllers"
	"github.com/ROFL1ST/quizzes-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Post("/register", controllers.RegisterUser)
	api.Post("/login", controllers.LoginUser)
	api.Post("/admin/register", controllers.RegisterAdmin)
	api.Post("/admin/login", controllers.LoginAdmin)

	api.Get("/topics", controllers.GetAllTopics)

	// Admin Routes
	roleGroup := api.Group("/admin/roles", middleware.Protected(), middleware.AllowRoles("supervisor"))
	roleGroup.Post("/", controllers.CreateRole)
	roleGroup.Get("/", controllers.GetAllRoles)

	topicGroup := api.Group("/admin/topics", middleware.Protected(), middleware.AllowRoles("supervisor", "admin"))
	topicGroup.Post("/", controllers.CreateTopic)

	quizGroup := api.Group("/admin/quizzes", middleware.Protected(), middleware.AllowRoles("supervisor", "admin", "pengajar"))
	quizGroup.Post("/", controllers.CreateQuiz)

	questionGroup := api.Group("/admin/questions", middleware.Protected(), middleware.AllowRoles("supervisor", "admin", "pengajar"))
	questionGroup.Post("/", controllers.CreateQuestion)

	// User Routes
	api.Get("/topics/:slug/quizzes", middleware.Protected(), controllers.GetQuizzesByTopicSlug)
	api.Get("/quizzes/:id/questions", middleware.Protected(), controllers.GetQuestionsByQuizID)

	history := api.Group("/history", middleware.Protected())
	history.Post("/", controllers.SaveHistory)
	history.Get("/", controllers.GetMyHistory)

	friends := api.Group("/friends", middleware.Protected())
	friends.Post("/add", controllers.AddFriend)
	friends.Get("/", controllers.GetMyFriends)
	friends.Delete("/:id", controllers.RemoveFriend)

	api.Get("/leaderboard/:slug", middleware.Protected(), controllers.GetLeaderboardByTopic)
}