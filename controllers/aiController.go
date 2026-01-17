package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

type TranslateRequest struct {
	Text       string `json:"text"`
	TargetLang string `json:"target_lang"`
}

type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}

func TranslateText(c *fiber.Ctx) error {
	var req TranslateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "AI service not configured (API Key missing)",
		})
	}

	var prompt string
	// Check if input looks like JSON (starts with {)
	if len(req.Text) > 0 && req.Text[0] == '{' {
		prompt = fmt.Sprintf("Translate the values in the following JSON to %s. Keep the keys and structure exactly the same. Return ONLY the JSON string: %s", req.TargetLang, req.Text)
	} else {
		prompt = fmt.Sprintf("Translate the following text to %s (only return the translated text, no extra explanation): %s", req.TargetLang, req.Text)
	}

	requestBody := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to marshal request"})
	}

	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.0-flash-exp" // Default to latest experimental if not set
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to request AI service"})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read body to see error details
		var errorBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorBody)
		fmt.Printf("AI Service Error Body: %v\n", errorBody) // Log to console

		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("AI service returned status: %d. details: %v", resp.StatusCode, errorBody),
		})
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to parse AI response"})
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		translatedText := geminiResp.Candidates[0].Content.Parts[0].Text
		return c.JSON(fiber.Map{
			"status": "success",
			"data": fiber.Map{
				"translatedText": translatedText,
			},
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "No translation received"})
}

// Bulk Translation
type BulkItem struct {
	ID      int         `json:"id"`
	Text    string      `json:"text"`
	Hint    string      `json:"hint,omitempty"` // Added Hint
	Options interface{} `json:"options"`        // Can be []string or string representation
}

type BulkTranslateRequest struct {
	Items      []BulkItem `json:"items"`
	TargetLang string     `json:"target_lang"`
}

func TranslateBulk(c *fiber.Ctx) error {
	var req BulkTranslateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "AI service not configured (API Key missing)",
		})
	}

	// Construct Prompt
	// We send the array structure directly to Gemini and ask it to translate values
	jsonBytes, _ := json.Marshal(req.Items)
	jsonString := string(jsonBytes)

	prompt := fmt.Sprintf(`Translate the 'text', 'options', and 'hint' (if present) values in the following JSON array to %s.
	IMPORTANT RULES:
	1. Keep the 'id' fields EXACTLY the same.
	2. Keep the structure EXACTLY the same (Array of Objects).
	3. Only translate the content values.
	4. Return ONLY the valid JSON array string. No markdown formatting.
	
	Input JSON:
	%s`, req.TargetLang, jsonString)

	requestBody := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to marshal request"})
	}

	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.0-flash-exp"
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)

	client := &http.Client{Timeout: 60 * time.Second} // Longer timeout for bulk
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to request AI service"})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorBody)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("AI service returned status: %d", resp.StatusCode),
			"details": errorBody,
		})
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to parse AI response"})
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		translatedText := geminiResp.Candidates[0].Content.Parts[0].Text

		// Clean up potential markdown formatting from AI
		// Sometimes AI returns ```json [ ... ] ```
		if len(translatedText) > 3 && translatedText[:3] == "```" {
			// Find first [ and last ]
			start := 0
			end := len(translatedText)
			for i, char := range translatedText {
				if char == '[' {
					start = i
					break
				}
			}
			for i := len(translatedText) - 1; i >= 0; i-- {
				if translatedText[i] == ']' {
					end = i + 1
					break
				}
			}
			if start < end {
				translatedText = translatedText[start:end]
			}
		}

		return c.JSON(fiber.Map{
			"status": "success",
			"data": fiber.Map{
				"translatedData": translatedText, // Send as string, let frontend parse
			},
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "No translation received"})
}
