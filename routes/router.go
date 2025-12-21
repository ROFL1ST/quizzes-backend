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

	// verify email
	api.Post("/verify-email", controllers.VerifyEmail)
	// forgot password
	api.Post("/forgot-password", controllers.ForgotPassword)
	api.Post("/reset-password", controllers.ResetPassword)

	api.Get("/topics", controllers.GetAllTopics)
	api.Get("/auth/me", middleware.Protected(), controllers.AuthMe)
	// Admin Routes
	adminGroup := api.Group("/admin", middleware.Protected())

	adminGroup.Get("/analytics", middleware.AllowRoles("supervisor", "admin"), controllers.GetDashboardAnalytics)

	// config
	configGroup := adminGroup.Group("/config", middleware.AllowRoles("supervisor"))
	configGroup.Get("/leveling", controllers.GetLevelingConfig)
	configGroup.Put("/leveling", controllers.UpdateLevelingConfig)
	// topic admin routes
	topicAdmin := adminGroup.Group("/topics", middleware.AllowRoles("supervisor", "admin"))
	topicAdmin.Get("/", controllers.GetAllTopicsAdmin)
	topicAdmin.Post("/", controllers.PostTopicAdmin)
	topicAdmin.Put("/:slug", controllers.UpdateTopicAdmin)
	topicAdmin.Delete("/:slug", controllers.DeleteTopicAdmin)
	topicAdmin.Get("/:slug", controllers.GetTopicBySlug)

	// quiz admin routes
	adminGroup.Get("/users", controllers.GetAllUsers)
	quizzesAdmin := adminGroup.Group("/quizzes", middleware.AllowRoles("supervisor", "admin", "pengajar"))
	quizzesAdmin.Get("/", controllers.GetAllQuizzesAdmin)
	quizzesAdmin.Post("/", controllers.CreateQuiz)
	quizzesAdmin.Put("/:id", controllers.UpdateQuizAdmin)
	quizzesAdmin.Delete("/:id", controllers.DeleteQuizAdmin)
	quizzesAdmin.Get("/analysis/:id", controllers.GetQuizAnalysisAdminById)

	// role management
	roleGroup := adminGroup.Group("/roles", middleware.AllowRoles("supervisor"))
	roleGroup.Post("/", controllers.CreateRole)
	roleGroup.Get("/", controllers.GetAllRoles)

	// notif
	notifGroup := api.Group("/notifications", middleware.Protected())
	notifGroup.Get("/", controllers.GetMyNotifications)
	notifGroup.Put("/:id/read", controllers.MarkNotificationRead)
	notifGroup.Put("/read-all", controllers.MarkAllNotificationsRead)
	notifGroup.Delete("/", controllers.ClearAllNotifications)
	// realtime notif stream
	notifGroup.Get("/stream", controllers.StreamNotifications)

	// question admin routes
	questionGroup := adminGroup.Group("/questions", middleware.AllowRoles("supervisor", "admin", "pengajar"))
	questionGroup.Get("/", controllers.GetAllQuestionsAdmin)
	questionGroup.Post("/", controllers.CreateQuestion)
	questionGroup.Post("/bulk", controllers.BulkUploadQuestions)
	questionGroup.Put("/:id", controllers.UpdateQuestionAdmin)
	questionGroup.Delete("/:id", controllers.DeleteQuestionAdmin)

	// shop routes admin
	shopAdmin := adminGroup.Group("/shop", middleware.AllowRoles("supervisor", "admin"))
	shopAdmin.Post("/items", controllers.CreateShopItem)
	shopAdmin.Put("/items/:id", controllers.UpdateShopItem)
	shopAdmin.Delete("/items/:id", controllers.DeleteShopItem)
	// =============================================================

	// User Routes
	api.Get("/topics/:slug/quizzes", middleware.Protected(), controllers.GetQuizzesByTopicSlug)
	api.Get("/quizzes/:id/questions", middleware.Protected(), controllers.GetQuestionsByQuizID)

	history := api.Group("/history", middleware.Protected())
	history.Post("/", controllers.SaveHistory)
	history.Get("/", controllers.GetMyHistory)
	history.Get("/:id", controllers.GetHistoryByID)

	friends := api.Group("/friends", middleware.Protected())

	friends.Get("/", controllers.GetMyFriends)              // Lihat daftar teman (accepted)
	friends.Get("/requests", controllers.GetFriendRequests) // Lihat request masuk
	friends.Get("/sent", controllers.GetSentRequests)

	friends.Post("/request", controllers.RequestFriend) // Minta berteman
	friends.Post("/confirm", controllers.ConfirmFriend) // Terima teman
	friends.Post("/refuse", controllers.RefuseFriend)   // Tolak teman

	friends.Delete("/:id", controllers.RemoveFriend) // Hapus teman
	friends.Delete("/cancel/:id", controllers.CancelRequest)

	api.Get("/leaderboard/:slug", middleware.Protected(), controllers.GetLeaderboardByTopic)

	// Challenge Routes
	challenges := api.Group("/challenges", middleware.Protected())
	challenges.Post("/", controllers.CreateChallenge)
	challenges.Get("/", controllers.GetMyChallenges)
	challenges.Post("/:id/accept", controllers.AcceptChallenge)
	challenges.Post("/:id/refuse", controllers.RejectChallenge)
	challenges.Get("/:id/lobby-stream", controllers.StreamChallengeLobby)
	challenges.Post("/:id/start", controllers.StartGameRealtime)
	challenges.Post("/:id/progress", controllers.UpdateChallengeProgress)

	// Activity Feed
	api.Get("/feed", middleware.Protected(), controllers.GetFriendActivity)

	// User Profile & Settings
	userGroup := api.Group("/users", middleware.Protected())
	userGroup.Get("/search", controllers.SearchUsers)
	userGroup.Get("/me", controllers.GetMyProfile) // Lihat profil & statistik sendiri
	userGroup.Get("/achievements", controllers.GetMyAchievements)
	userGroup.Put("/me", controllers.UpdateProfile) // Ganti nama/password
	userGroup.Get("/:username", controllers.GetUserProfile)
	userGroup.Post("/share", controllers.ShareProfileTrigger)
	userGroup.Get("/analytics/smart", controllers.GetUserSmartAnalytics)
    userGroup.Get("/activity/calendar", controllers.GetActivityCalendar)

	// Shop Routes
	shopGroup := api.Group("/shop", middleware.Protected())
	shopGroup.Get("/items", controllers.GetShopItems)
	shopGroup.Post("/buy", controllers.BuyItem)
	shopGroup.Get("/inventory", controllers.GetMyInventory)
	shopGroup.Post("/equip", controllers.EquipItem)

	// daily routes
	daily := api.Group("/daily", middleware.Protected())
	daily.Get("/info", controllers.GetDailyInfo)
	daily.Post("/claim-login", controllers.ClaimLoginReward)
	daily.Post("/claim-mission", controllers.ClaimMissionReward)

	// comunity quiz
	comunityGroup := api.Group("/community", middleware.Protected())
	comunityGroup.Post("/quizzes", controllers.CreateCommunityQuiz)
	comunityGroup.Get("/quizzes", controllers.GetCommunityQuizzes)
	comunityGroup.Get("/quizzes/me", controllers.GetMyCommunityQuizzes)

	api.Get("/quizzes/remedial/start", middleware.Protected(), controllers.GetRemedialQuestions)
}
