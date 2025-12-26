package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"encoding/json"
	"strconv"
	"github.com/gofiber/fiber/v2"
)

func CreateQuiz(c *fiber.Ctx) error {
	var quiz models.Quiz
	if err := c.BodyParser(&quiz); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	var count int64
	config.DB.Model(&models.Topic{}).Where("id = ?", quiz.TopicID).Count(&count)
	if count == 0 {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Topic not found", nil)
	}

	if err := config.DB.Create(&quiz).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed create quiz", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Quiz created", quiz)
}

func GetQuizzesByTopicSlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var topic models.Topic
	if err := config.DB.Where("slug = ?", slug).First(&topic).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Topic not found", nil)
	}

	var quizzes []models.Quiz
	config.DB.Where("(topic_id = ?) AND active=TRUE", topic.ID).Find(&quizzes)

	return utils.SuccessResponse(c, fiber.StatusOK, "Quizzes retrieved", quizzes)
}

func CreateCommunityQuiz(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	var quiz models.Quiz

	if err := c.BodyParser(&quiz); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	creID := uint(userID)
	quiz.CreatorID = &creID
	quiz.Active = true
	quiz.Status = "published"

	if err := config.DB.Create(&quiz).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat kuis", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Kuis berhasil dibuat!", quiz)
}

func GetCommunityQuizzes(c *fiber.Ctx) error {
	var quizzes []models.Quiz

	// Filter: Bukan buatan Admin (CreatorID != NULL), Active, dan Public
	err := config.DB.Preload("Topic").Preload("Creator").
		Where("creator_id IS NOT NULL AND active = ? AND is_public = ?", true, true).
		Find(&quizzes).Error

	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Daftar Kuis Komunitas", quizzes)
}

func GetMyCommunityQuizzes(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	var quizzes []models.Quiz
	err := config.DB.Preload("Topic").
		Where("creator_id = ?", uint(userID)).
		Find(&quizzes).Error
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Daftar Kuis Buatanmu", quizzes)
}


func GetRemedialQuestions(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(float64)

    // 1. Ambil 10 history terakhir user
    var histories []models.History
    config.DB.Where("user_id = ?", userID).Order("created_at desc").Limit(10).Find(&histories)

    wrongQIDs := []uint{}
    seenQIDs := make(map[uint]bool) // Untuk mencegah soal duplikat

    // 2. Loop history untuk cari jawaban salah
    for _, h := range histories {
        var userAnswers map[string]string
        if err := json.Unmarshal(h.Snapshot, &userAnswers); err != nil {
            continue
        }

        for qIDStr, userAns := range userAnswers {
            qIDInt, _ := strconv.Atoi(qIDStr)
            qID := uint(qIDInt)

            if seenQIDs[qID] { continue } // Skip jika sudah masuk list

            // Cek ke Database (Optimasi: Bisa pakai map caching jika data besar)
            var q models.Question
            if err := config.DB.Select("id, correct_answer").First(&q, qID).Error; err == nil {
                // Logic simpel: jika string jawaban beda, anggap salah
                // (Untuk production, gunakan logic grading yang lebih detail seperti di SaveHistory)
                if q.CorrectAnswer != userAns {
                    wrongQIDs = append(wrongQIDs, q.ID)
                    seenQIDs[qID] = true
                }
            }
        }
        if len(wrongQIDs) >= 10 { break } // Cukup 10 soal
    }

    if len(wrongQIDs) == 0 {
        return utils.ErrorResponse(c, fiber.StatusNotFound, "Tidak ada soal remedial. Kamu hebat!", nil)
    }

    // 3. Ambil data soal lengkap
    var questions []models.Question
    config.DB.Preload("Quiz").Where("id IN ?", wrongQIDs).Find(&questions)

    return utils.SuccessResponse(c, fiber.StatusOK, "Sesi Remedial Dimulai", questions)
}