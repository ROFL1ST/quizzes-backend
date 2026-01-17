package controllers

import (
	"math"

	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
)

// GetTranslations returns all translations formatted as nested JSON
func GetTranslations(c *fiber.Ctx) error {
	var translations []models.Translation
	if err := config.DB.Find(&translations).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Failed to fetch translations"})
	}

	// Re-implementing correctly using a helper map
	finalStructure := make(map[string]map[string]map[string]interface{}) // lang -> section -> key -> value

	for _, t := range translations {
		// content := t.Translations // This is JSONB
		// We need to unmarshal t.Translations into a map[string]interface{}
		var langs map[string]interface{}
		if err := c.App().Config().JSONDecoder(t.Translations, &langs); err != nil {
			continue
		}

		for lang, value := range langs {
			if finalStructure[lang] == nil {
				finalStructure[lang] = make(map[string]map[string]interface{})
			}
			if finalStructure[lang][t.Section] == nil {
				finalStructure[lang][t.Section] = make(map[string]interface{})
			}
			finalStructure[lang][t.Section][t.Key] = value
		}
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   finalStructure,
	})
}

// GetAdminTranslations returns flat paginated list for Admin UI
func GetAdminTranslations(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search")
	section := c.Query("section")
	status := c.Query("status")
	offset := (page - 1) * limit

	db := config.DB.Model(&models.Translation{})

	if section != "" && section != "all" {
		db = db.Where("section = ?", section)
	}

	if search != "" {
		term := "%" + search + "%"
		// Cast to text for searching inside JSON
		db = db.Where("section ILIKE ? OR key ILIKE ? OR translations::text ILIKE ?", term, term, term)
	}

	if status != "" && status != "all" {
		lang := ""
		switch status {
		case "missing_id":
			lang = "id"
		case "missing_en":
			lang = "en"
		case "missing_jp":
			lang = "jp"
		}
		if lang != "" {
			// Check if key exists but is empty string OR doesn't exist
			// Postgres JSONB syntax:
			// 1. Key missing: NOT (translations ? 'id')
			// 2. Value empty: translations->>'id' = ''
			// Combined: logic depends on how we store. We usually store {"id": ""}.
			// Let's assume checking for empty string value is enough.
			db = db.Where("translations->>? = '' OR NOT (translations ? ?)", lang, lang)
		}
	}

	var total int64
	db.Count(&total)

	var translations []models.Translation
	db.Offset(offset).Limit(limit).Order("section ASC, key ASC").Find(&translations)

	// Flatten for frontend
	type FlatTranslation struct {
		ID      string                 `json:"id"`
		Section string                 `json:"section"`
		Key     string                 `json:"key"`
		Values  map[string]interface{} `json:"values"`
	}

	var flatList []FlatTranslation

	for _, t := range translations {
		var content map[string]interface{}
		_ = c.App().Config().JSONDecoder(t.Translations, &content)
		if content == nil {
			content = make(map[string]interface{})
		}

		// Fill defaults
		if _, ok := content["id"]; !ok {
			content["id"] = ""
		}
		if _, ok := content["en"]; !ok {
			content["en"] = ""
		}
		if _, ok := content["jp"]; !ok {
			content["jp"] = ""
		}

		// Add IsObj flags? Frontend can detect type.
		// But for simple "values" map, let's just pass it.

		flatList = append(flatList, FlatTranslation{
			ID:      t.Section + "." + t.Key,
			Section: t.Section,
			Key:     t.Key,
			Values:  content,
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   flatList,
		"meta": fiber.Map{
			"page":      page,
			"limit":     limit,
			"total":     total,
			"last_page": math.Ceil(float64(total) / float64(limit)),
		},
	})
}

// SyncTranslations accepts a JSON object matching 'translations.js' structure and populates the DB.
// Expected Input: { "id": { "navbar": { "title": "..." } }, "en": ... }
func SyncTranslations(c *fiber.Ctx) error {
	// Value can be string or nested map
	var input map[string]map[string]map[string]interface{}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Invalid JSON", "error": err.Error()})
	}

	count := 0
	for lang, sections := range input {
		for section, keys := range sections {
			for key, value := range keys {
				var trans models.Translation
				// Find existing record by Section + Key
				err := config.DB.Where("section = ? AND key = ?", section, key).First(&trans).Error

				var content map[string]interface{}
				if err == nil {
					// Found, update the specific lang
					_ = c.App().Config().JSONDecoder(trans.Translations, &content)
					if content == nil {
						content = make(map[string]interface{})
					}
					content[lang] = value
					// Save back
					jsonBytes, _ := c.App().Config().JSONEncoder(content)
					trans.Translations = datatypes.JSON(jsonBytes)
					config.DB.Save(&trans)
				} else {
					// Not found, create new
					content = make(map[string]interface{})
					content[lang] = value
					jsonBytes, _ := c.App().Config().JSONEncoder(content)

					newTrans := models.Translation{
						Section:      section,
						Key:          key,
						Translations: datatypes.JSON(jsonBytes),
					}
					config.DB.Create(&newTrans)
				}
				count++
			}
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Synced successfully",
		"count":   count,
	})
}
