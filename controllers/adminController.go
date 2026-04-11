package controllers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
)

type DashboardAnalytics struct {
	TotalUsers       int64          `json:"total_users"`
	TotalQuizzes     int64          `json:"total_quizzes"`
	TotalQuestions   int64          `json:"total_questions"`
	TotalAttempts    int64          `json:"total_attempts"`
	AverageScore     float64        `json:"average_score"`
	ActiveUsers      int64          `json:"active_users"`
	WeeklyStats      []WeeklyStat   `json:"weekly_stats"`
	TopicStats       []TopicStat    `json:"topic_stats"`
	HardestQuestions []QuestionStat `json:"hardest_questions"`
}

type WeeklyStat struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type TopicStat struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

type QuestionStat struct {
	QuestionText string `json:"question_text"`
	Correct      int    `json:"correct"`
	Incorrect    int    `json:"incorrect"`
}

type QuestionAnalysis struct {
	ID             uint   `json:"id"`
	QuestionText   string `json:"question_text"`
	CorrectCount   int    `json:"correct_count"`
	IncorrectCount int    `json:"incorrect_count"`
	TotalAttempts  int    `json:"total_attempts"`
	Difficulty     string `json:"difficulty"`
	AccuracyRate   string `json:"accuracy_rate"`
}

type ConfigInput struct {
	Value string `json:"value"`
}

func GetLevelingConfig(c *fiber.Ctx) error {
	var conf models.SystemConfig
	if err := config.DB.Where("key = ?", "leveling_factor").First(&conf).Error; err != nil {
		// Jika belum ada, return default
		return utils.SuccessResponse(c, fiber.StatusOK, "Config retrieved", fiber.Map{"value": "100"})
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Config retrieved", conf)
}

func UpdateLevelingConfig(c *fiber.Ctx) error {
	var input ConfigInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", nil)
	}

	// Validasi angka
	if _, err := strconv.ParseFloat(input.Value, 64); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Value must be a number", nil)
	}

	var conf models.SystemConfig
	// Upsert (Update jika ada, Create jika tidak)
	config.DB.Where("key = ?", "leveling_factor").Assign(models.SystemConfig{Value: input.Value}).FirstOrCreate(&conf)

	// Simpan nilai baru
	conf.Value = input.Value
	config.DB.Save(&conf)

	return utils.SuccessResponse(c, fiber.StatusOK, "Leveling difficulty updated", conf)
}

func GetDashboardAnalytics(c *fiber.Ctx) error {
	var totalUsers, totalQuizzes, totalAttempts, totalQuestions int64
	var avgScore float64
	var activeUsers int64
	var weeklyStats []WeeklyStat
	var topicStats []TopicStat
	var hardestQuestions []QuestionStat

	role := c.Locals("role").(string)

	if role == "pengajar" {
		userID := uint(c.Locals("user_id").(float64))

		// 1. Total Students (Users)
		config.DB.Raw("SELECT COUNT(DISTINCT student_id) FROM classroom_members cm JOIN classrooms c ON c.id = cm.classroom_id WHERE c.teacher_id = ?", userID).Scan(&totalUsers)

		// 2. Total Classrooms (Mapped to Quizzes)
		config.DB.Model(&models.Classroom{}).Where("teacher_id = ?", userID).Count(&totalQuizzes)

		// 3. Total Assignments (Mapped to Questions)
		config.DB.Raw("SELECT COUNT(*) FROM assignments a JOIN classrooms c ON c.id = a.classroom_id WHERE c.teacher_id = ?", userID).Scan(&totalQuestions)

		// 4. Total Attempts (Dummy for now or Submissions)
		totalAttempts = 0

	} else {
		// Admin / Supervisor Logic
		if err := config.DB.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Error counting users", nil)
		}

		if err := config.DB.Model(&models.Quiz{}).Count(&totalQuizzes).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Error counting quizzes", nil)
		}

		if err := config.DB.Model(&models.Question{}).Count(&totalQuestions).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Error counting questions", nil)
		}

		if err := config.DB.Model(&models.History{}).Count(&totalAttempts).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Error counting history", nil)
		}

		if err := config.DB.Model(&models.History{}).Select("COALESCE(AVG(score), 0)").Scan(&avgScore).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Error calculating average", nil)
		}

		sevenDaysAgo := time.Now().AddDate(0, 0, -7)
		if err := config.DB.Model(&models.History{}).
			Where("created_at >= ?", sevenDaysAgo).
			Distinct("user_id").
			Count(&activeUsers).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Error counting active users", nil)
		}

		// 1. Weekly Stats (Last 7 Days)
		config.DB.Raw(`
			SELECT TO_CHAR(date_trunc('day', created_at), 'YYYY-MM-DD') as date, COUNT(*) as count 
			FROM histories 
			WHERE created_at >= ? 
			GROUP BY date 
			ORDER BY date ASC
		`, sevenDaysAgo).Scan(&weeklyStats)

		// 2. Topic Distribution
		config.DB.Raw(`
			SELECT t.title as label, COUNT(q.id) as value
			FROM topics t
			JOIN quizzes q ON q.topic_id = t.id
			GROUP BY t.title
		`).Scan(&topicStats)
	}

	// 3. Hardest Questions (Most Incorrect)
	config.DB.Model(&models.Question{}).
		Select("question_text, correct_count as correct, incorrect_count as incorrect").
		Order("incorrect_count DESC").
		Limit(5).
		Scan(&hardestQuestions)

	return utils.SuccessResponse(c, fiber.StatusOK, "Dashboard analytics retrieved", DashboardAnalytics{
		TotalUsers:       totalUsers,
		TotalQuizzes:     totalQuizzes,
		TotalQuestions:   totalQuestions,
		TotalAttempts:    totalAttempts,
		AverageScore:     avgScore,
		ActiveUsers:      activeUsers,
		WeeklyStats:      weeklyStats,
		TopicStats:       topicStats,
		HardestQuestions: hardestQuestions,
	})
}

func GetAllUsers(c *fiber.Ctx) error {
	params := utils.GetPaginationParams(c)
	var users []models.User
	var total int64

	if err := config.DB.Model(&models.User{}).Count(&total).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to count users", nil)
	}

	if err := config.DB.Order("created_at desc").Limit(params.PageSize).Offset(params.Offset).Find(&users).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch users", nil)
	}

	return utils.PaginatedSuccessResponse(c, fiber.StatusOK, "Users retrieved", users, total, params)
}

func GetAllAdmins(c *fiber.Ctx) error {
	params := utils.GetPaginationParams(c)
	var admins []models.Admin
	var total int64

	if err := config.DB.Model(&models.Admin{}).Count(&total).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to count admins", nil)
	}

	if err := config.DB.Preload("Role").Order("created_at desc").Limit(params.PageSize).Offset(params.Offset).Find(&admins).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch admins", nil)
	}

	return utils.PaginatedSuccessResponse(c, fiber.StatusOK, "Admins retrieved", admins, total, params)
}

// ========== TOPICS WITH PAGINATION ==========
func GetAllTopicsAdmin(c *fiber.Ctx) error {
	params := utils.GetPaginationParams(c)

	var topics []models.Topic
	var total int64

	// Hitung total data
	if err := config.DB.Model(&models.Topic{}).Count(&total).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to count topics", nil)
	}

	// Ambil data dengan pagination
	if err := config.DB.
		Order("created_at desc").
		Limit(params.PageSize).
		Offset(params.Offset).
		Find(&topics).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch topics", nil)
	}

	return utils.PaginatedSuccessResponse(c, fiber.StatusOK, "Topics retrieved", topics, total, params)
}

func PostTopicAdmin(c *fiber.Ctx) error {
	var topic models.Topic
	if err := c.BodyParser(&topic); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}
	topic.Slug = utils.GenerateSlug(topic.Title)
	if err := config.DB.Create(&topic).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed create topic", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, "Topic created", topic)
}

func DeleteTopicAdmin(c *fiber.Ctx) error {
	slug := c.Params("slug")

	result := config.DB.Where("slug = ?", slug).Delete(&models.Topic{})

	if result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed delete topic", result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Topic not found", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Topic deleted", nil)
}

func UpdateTopicAdmin(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var topic models.Topic
	if err := config.DB.Where("slug = ?", slug).First(&topic).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Topic not found", nil)
	}
	var updateData models.Topic
	if err := c.BodyParser(&updateData); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}
	topic.Title = updateData.Title
	topic.Description = updateData.Description
	if updateData.Slug != "" {
		topic.Slug = updateData.Slug
	}

	if err := config.DB.Save(&topic).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed update topic", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Topic updated", topic)
}

// ========== QUIZZES WITH PAGINATION ==========
func GetAllQuizzesAdmin(c *fiber.Ctx) error {
	params := utils.GetPaginationParams(c)

	var quizzes []models.Quiz
	var total int64

	// Hitung total data
	query := config.DB.Model(&models.Quiz{})

	// 1. Filter by Classroom (Strict)
	if classroomID := c.Query("classroom_id"); classroomID != "" {
		query = query.Where("classroom_id = ?", classroomID)
	} else {
		// 2. Global Context
		onlyGlobal := c.Query("only_global") == "true"
		role := c.Locals("role").(string)

		if onlyGlobal {
			// Explicitly requested Global Only
			query = query.Where("topic_id IS NOT NULL")
		} else {
			// Default Behavior depends on Role
			if role == "admin" || role == "supervisor" {
				// Admin sees EVERYTHING (Global + All Classroom Quizzes)
				// No filter needed
			} else {
				// Teachers/User see only Global Quizzes in the main list
				// To avoid cluttering with other people's class quizzes
				query = query.Where("topic_id IS NOT NULL")
			}
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to count quizzes", nil)
	}

	// Ambil data dengan pagination
	if err := query.
		Preload("Topic").
		Order("created_at desc").
		Limit(params.PageSize).
		Offset(params.Offset).
		Find(&quizzes).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch quizzes", nil)
	}

	return utils.PaginatedSuccessResponse(c, fiber.StatusOK, "Quizzes retrieved", quizzes, total, params)
}

func PostQuizAdmin(c *fiber.Ctx) error {
	var quiz models.Quiz
	if err := c.BodyParser(&quiz); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}
	if err := config.DB.Create(&quiz).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed create quiz", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, "Quiz created", quiz)
}

func UpdateQuizAdmin(c *fiber.Ctx) error {
	id := c.Params("id")
	var quiz models.Quiz
	if err := config.DB.First(&quiz, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Quiz not found", nil)
	}
	if err := c.BodyParser(&quiz); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}
	if err := config.DB.Save(&quiz).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed update quiz", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Quiz updated", quiz)
}

func DeleteQuizAdmin(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := config.DB.Delete(&models.Quiz{}, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed delete quiz", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Quiz deleted", nil)
}

func GetQuizAnalysisAdminById(c *fiber.Ctx) error {
	quizID := c.Params("id")
	var questions []models.Question
	if err := config.DB.Where("quiz_id = ?", quizID).Order("id asc").Find(&questions).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Questions not found", nil)
	}
	var analysis []QuestionAnalysis
	for _, q := range questions {
		total := q.CorrectCount + q.IncorrectCount
		accuracy := 0.0
		difficulty := "Belum ada data"
		if total > 0 {
			accuracy = (float64(q.CorrectCount) / float64(total)) * 100
			if accuracy > 80 {
				difficulty = "Mudah"
			} else if accuracy > 40 {
				difficulty = "Sedang"
			} else {
				difficulty = "Sulit"
			}
		}
		analysis = append(analysis, QuestionAnalysis{
			ID:             q.ID,
			QuestionText:   q.QuestionText,
			CorrectCount:   q.CorrectCount,
			IncorrectCount: q.IncorrectCount,
			TotalAttempts:  total,
			Difficulty:     difficulty,
			AccuracyRate:   fmt.Sprintf("%.1f%%", accuracy),
		})
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Question analysis retrieved", analysis)
}

// ========== QUESTIONS WITH PAGINATION ==========
func GetAllQuestionsAdmin(c *fiber.Ctx) error {
	params := utils.GetPaginationParams(c)

	var questions []models.Question
	var total int64

	// Query Dasar
	query := config.DB.Model(&models.Question{})

	// 1. Filter by Quiz ID (Jika ada)
	if quizID := c.Query("quiz_id"); quizID != "" && quizID != "0" {
		query = query.Where("quiz_id = ?", quizID)
	}

	// 2. Filter by Search Keyword (Pertanyaan)
	if key := c.Query("key"); key != "" {
		search := "%" + strings.ToLower(key) + "%"
		query = query.Where("LOWER(question_text) LIKE ?", search)
	}

	// Hitung Total Data (setelah filter)
	if err := query.Count(&total).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to count questions", nil)
	}

	// Ambil Data dengan Pagination
	if err := query.
		Preload("Quiz").
		Order("created_at desc").
		Limit(params.PageSize).
		Offset(params.Offset).
		Find(&questions).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch questions", nil)
	}

	return utils.PaginatedSuccessResponse(c, fiber.StatusOK, "Questions retrieved", questions, total, params)
}

func PostQuestionAdmin(c *fiber.Ctx) error {
	var question models.Question
	if err := c.BodyParser(&question); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}
	if err := config.DB.Create(&question).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed create question", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, "Question created", question)
}

func UpdateQuestionAdmin(c *fiber.Ctx) error {
	id := c.Params("id")
	var question models.Question
	if err := config.DB.First(&question, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Question not found", nil)
	}
	if err := c.BodyParser(&question); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}
	if err := config.DB.Save(&question).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed update question", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Question updated", question)
}

func DeleteQuestionAdmin(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := config.DB.Delete(&models.Question{}, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed delete question", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Question deleted", nil)
}

func BulkUploadQuestions(c *fiber.Ctx) error {
	quizID, _ := strconv.Atoi(c.FormValue("quiz_id"))

	file, err := c.FormFile("file")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "File required", nil)
	}

	f, _ := file.Open()
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Failed parse CSV", nil)
	}

	var questions []models.Question
	for i, row := range records {
		if i == 0 {
			continue
		}

		if len(row) < 5 {
			continue
		}

		// Kolom CSV:
		// 0: Question
		// 1: Type (mcq, short_answer, boolean, multi_select)
		// 2: Options (dipisah koma, misal: "A,B,C")
		// 3: CorrectAnswer
		// 4: Hint

		qType := strings.TrimSpace(strings.ToLower(row[1]))
		if qType == "" {
			qType = "mcq"
		}

		var options []string

		// Logic Opsi Berdasarkan Tipe
		if qType == "boolean" {
			options = []string{"Benar", "Salah"}
		} else if qType == "short_answer" {
			options = []string{} // Kosongkan opsi
		} else {
			// Untuk MCQ & Multi Select, split string opsi berdasarkan koma
			if row[2] != "" {
				rawOpts := strings.Split(row[2], ",")
				for _, o := range rawOpts {
					options = append(options, strings.TrimSpace(o))
				}
			}
		}

		// Logic Jawaban Benar
		correct := row[3]
		// Jika Multi Select, pastikan formatnya JSON String array jika belum
		// (User di CSV mungkin nulis "A, B". Kita ubah jadi '["A","B"]')
		if qType == "multi_select" && !strings.HasPrefix(correct, "[") {
			answers := strings.Split(correct, ",")
			for j := range answers {
				answers[j] = strings.TrimSpace(answers[j])
			}
			jsonBytes, _ := json.Marshal(answers)
			correct = string(jsonBytes)
		}

		// Difficulty (Col 5, optional)
		difficulty := 0.5
		if len(row) > 5 {
			if val, err := strconv.ParseFloat(strings.TrimSpace(row[5]), 64); err == nil {
				difficulty = val
			}
		}

		q := models.Question{
			QuizID:        uint(quizID),
			QuestionText:  row[0],
			Type:          qType,
			Options:       pq.StringArray(options),
			CorrectAnswer: correct,
			Hint:          row[4],
			Difficulty:    difficulty,
		}
		questions = append(questions, q)
	}

	if err := config.DB.Create(&questions).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed insert questions", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Bulk upload success", fiber.Map{
		"total_inserted": len(questions),
	})
}

func CreateShopItem(c *fiber.Ctx) error {
	var item models.Item
	if err := c.BodyParser(&item); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}
	if err := config.DB.Create(&item).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed create item", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, "Item created", item)

}

func UpdateShopItem(c *fiber.Ctx) error {
	id := c.Params("id")

	// Ambil state lama sebelum diupdate
	var item models.Item
	if err := config.DB.First(&item, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Item not found", nil)
	}

	wasActive := item.IsActive

	if err := c.BodyParser(&item); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	tx := config.DB.Begin()

	if err := tx.Save(&item).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed update item", err.Error())
	}

	// Jika item baru saja di-deactivate, auto-unequip dari semua user
	if wasActive && !item.IsActive {
		if err := tx.Model(&models.UserItem{}).
			Where("item_id = ? AND is_equipped = ?", item.ID, true).
			Update("is_equipped", false).Error; err != nil {
			tx.Rollback()
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to unequip item from users", err.Error())
		}
	}

	tx.Commit()
	return utils.SuccessResponse(c, fiber.StatusOK, "Item updated", item)
}


func DeleteShopItem(c *fiber.Ctx) error {
	id := c.Params("id")

	tx := config.DB.Begin()

	// Unequip dari semua user terlebih dahulu
	if err := tx.Model(&models.UserItem{}).
		Where("item_id = ? AND is_equipped = ?", id, true).
		Update("is_equipped", false).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to unequip item", err.Error())
	}

	if err := tx.Delete(&models.Item{}, id).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed delete item", err.Error())
	}

	tx.Commit()
	return utils.SuccessResponse(c, fiber.StatusOK, "Item deleted", nil)
}

type CopyQuestionsInput struct {
	TargetQuizID uint   `json:"target_quiz_id"`
	QuestionIDs  []uint `json:"question_ids"`
}

func CopyQuestions(c *fiber.Ctx) error {
	var input CopyQuestionsInput
	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	if input.TargetQuizID == 0 || len(input.QuestionIDs) == 0 {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Target Quiz and Questions are required", nil)
	}

	var sourceQuestions []models.Question
	if err := config.DB.Where("id IN ?", input.QuestionIDs).Find(&sourceQuestions).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch source questions", err.Error())
	}

	var newQuestions []models.Question
	for _, q := range sourceQuestions {
		newQ := models.Question{
			QuizID:        input.TargetQuizID,
			QuestionText:  q.QuestionText,
			Type:          q.Type,
			Options:       q.Options,
			CorrectAnswer: q.CorrectAnswer,
			Hint:          q.Hint,
			Difficulty:    q.Difficulty,
			// ID will be auto-generated
			// CreatedAt/UpdatedAt auto-generated
		}
		newQuestions = append(newQuestions, newQ)
	}

	if len(newQuestions) > 0 {
		if err := config.DB.Create(&newQuestions).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to copy questions", err.Error())
		}
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, fmt.Sprintf("%d questions copied successfully", len(newQuestions)), nil)
}
