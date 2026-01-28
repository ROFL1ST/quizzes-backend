package routes

import (
	"time"

	"github.com/ROFL1ST/quizzes-backend/controllers"
	"github.com/ROFL1ST/quizzes-backend/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Post("/register", controllers.RegisterUser)
	api.Post("/login", controllers.LoginUser)
	api.Post("/admin/register", controllers.RegisterAdmin)
	api.Post("/admin/login", controllers.LoginAdmin)

	// Public Announcements
	api.Get("/announcements", controllers.GetAnnouncements)

	// verify email
	api.Post("/verify-email", controllers.VerifyEmail)
	// forgot password
	api.Post("/forgot-password", controllers.ForgotPassword)
	api.Post("/reset-password", controllers.ResetPassword)

	// AI Service
	api.Post("/ai/translate", middleware.Protected(), controllers.TranslateText)
	api.Post("/ai/translate-bulk", middleware.Protected(), controllers.TranslateBulk)

	api.Get("/topics", controllers.GetAllTopics)
	api.Get("/auth/me", middleware.Protected(), controllers.AuthMe)
	api.Get("/admin/me", middleware.Protected(), controllers.AuthMe) // New: Dedicated Admin Profile
	// Admin Routes
	adminGroup := api.Group("/admin", middleware.Protected())

	// Superadmin only: Create Admin
	adminGroup.Post("/create-admin", middleware.AllowRoles("supervisor"), controllers.CreateAdmin)

	adminGroup.Get("/analytics", middleware.AllowRoles("supervisor", "admin"), controllers.GetDashboardAnalytics)
	adminGroup.Get("/health", middleware.AllowRoles("supervisor", "admin"), controllers.GetSystemHealth) // New Health Check

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
	adminGroup.Get("/admins", controllers.GetAllAdmins)
	quizzesAdmin := adminGroup.Group("/quizzes", middleware.AllowRoles("supervisor", "admin", "pengajar"))
	quizzesAdmin.Get("/", controllers.GetAllQuizzesAdmin)
	quizzesAdmin.Post("/", controllers.CreateQuiz)
	quizzesAdmin.Put("/:id", controllers.UpdateQuizAdmin)
	quizzesAdmin.Delete("/:id", middleware.AllowRoles("supervisor", "admin"), controllers.DeleteQuizAdmin)
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
	questionGroup.Delete("/:id", middleware.AllowRoles("supervisor", "admin"), controllers.DeleteQuestionAdmin)
	// Dev/Test Route
	questionGroup.Post("/randomize-difficulty", controllers.RandomizeDifficulty)
	questionGroup.Post("/recalculate-difficulty", controllers.RecalculateDifficulty)

	// shop routes admin
	shopAdmin := adminGroup.Group("/shop", middleware.AllowRoles("supervisor", "admin"))
	shopAdmin.Post("/items", controllers.CreateShopItem)
	shopAdmin.Put("/items/:id", controllers.UpdateShopItem)
	shopAdmin.Delete("/items/:id", controllers.DeleteShopItem)

	// Report Admin Routes
	reportAdmin := adminGroup.Group("/reports", middleware.AllowRoles("supervisor", "admin"))
	reportAdmin.Get("/", controllers.GetAllReports)
	reportAdmin.Put("/:id", controllers.ResolveReport)

	// Review Admin Routes
	reviewAdmin := adminGroup.Group("/reviews", middleware.AllowRoles("supervisor", "admin"))
	reviewAdmin.Get("/", controllers.GetAllReviews)
	reviewAdmin.Delete("/:id", controllers.DeleteReview) // Optional: If needed

	// Ban/Unban Routes
	adminGroup.Put("/users/:id/ban", controllers.BanUser)
	adminGroup.Put("/users/:id/unban", controllers.UnbanUser)

	// Rate Limiter for Public Translations
	translationLimiter := limiter.New(limiter.Config{
		Max:        20,              // 20 requests
		Expiration: 1 * time.Minute, // per minute
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // Limit by IP
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"status":  "error",
				"message": "Too many requests. Please try again later.",
			})
		},
	})

	// Public Translations with Rate Limiting
	api.Get("/public/translations", translationLimiter, controllers.GetTranslations)

	// Admin Translations
	adminGroup.Get("/translations", middleware.AllowRoles("supervisor", "admin"), controllers.GetAdminTranslations)
	adminGroup.Post("/translations/sync", middleware.AllowRoles("supervisor", "admin"), controllers.SyncTranslations)

	// Logout (User)
	api.Post("/logout", controllers.Logout)
	// Logout (Admin)
	api.Post("/admin/logout", controllers.LogoutAdmin)

	// Broadcast Route
	adminGroup.Post("/broadcast", middleware.AllowRoles("supervisor", "admin"), controllers.Broadcast)

	// Classroom Admin Routes
	classroomAdmin := adminGroup.Group("/classrooms", middleware.AllowRoles("supervisor", "admin", "pengajar"))
	classroomAdmin.Get("/", controllers.GetAllClassrooms)
	classroomAdmin.Get("/:id", controllers.GetClassroomDetails) // Reuse detail logic
	classroomAdmin.Post("/members", controllers.AddClassroomMember)
	classroomAdmin.Delete("/:id/members/:studentId", controllers.RemoveClassroomMember)
	classroomAdmin.Delete("/assignments/:id", controllers.DeleteAssignment)
	classroomAdmin.Get("/assignments/:id/submissions", controllers.GetAssignmentSubmissions)
	classroomAdmin.Put("/:id/teacher", controllers.AssignClassroomTeacher)

	// =============================================================

	// User Routes
	api.Get("/topics/:slug/quizzes", middleware.Protected(), controllers.GetQuizzesByTopicSlug)
	api.Get("/quizzes/:id/questions", middleware.Protected(), controllers.GetQuestionsByQuizID)
	api.Post("/quiz/adaptive/next", middleware.Protected(), controllers.GetNextAdaptiveQuestion)

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
	challenges.Post("/join", controllers.JoinChallengeByCode) // New
	challenges.Get("/", controllers.GetMyChallenges)
	challenges.Post("/:id/accept", controllers.AcceptChallenge)
	challenges.Post("/:id/refuse", controllers.RejectChallenge)
	challenges.Put("/:id/settings", controllers.UpdateLobbySettings) // New
	challenges.Get("/:id/lobby-stream", controllers.StreamChallengeLobby)
	challenges.Post("/:id/start", controllers.StartGameRealtime)
	challenges.Post("/:id/progress", controllers.UpdateChallengeProgress)
	challenges.Post("/:id/invite", controllers.InviteToLobby)  // NEW
	challenges.Post("/:id/code", controllers.GenerateRoomCode) // NEW
	challenges.Post("/:id/leave", controllers.LeaveLobby)

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
	userGroup.Get("/adaptivity", controllers.GetAdaptivity)

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

	// Global Leaderboard
	api.Get("/global/leaderboard", middleware.Protected(), controllers.GetGlobalLeaderboard)

	// Report Routes (User)
	api.Post("/reports", middleware.Protected(), controllers.CreateReport)

	// Review Routes
	api.Post("/quizzes/:id/reviews", middleware.Protected(), controllers.AddReview)
	api.Get("/quizzes/:id/reviews", middleware.Protected(), controllers.GetReviews)

	// Classroom Routes
	classroomGroup := api.Group("/classrooms", middleware.Protected())
	classroomGroup.Post("/", controllers.CreateClassroom)                 // Create Class (Teacher)
	classroomGroup.Get("/", controllers.GetMyClassrooms)                  // List my classes
	classroomGroup.Post("/join", controllers.JoinClassroom)               // Join Class (Student)
	classroomGroup.Get("/:id", controllers.GetClassroomDetails)           // Class Details
	classroomGroup.Post("/:id/assignments", controllers.CreateAssignment) // Create Assignment (Teacher)

	// Survival Mode
	api.Post("/survival/start", middleware.Protected(), controllers.StartSurvival)
	api.Post("/survival/answer", middleware.Protected(), controllers.AnswerSurvival)

}
