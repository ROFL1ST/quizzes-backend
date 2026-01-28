package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type UserHistoryItem struct {
	QuestionID int     `json:"question_id"`
	IsCorrect  bool    `json:"is_correct"`
	Difficulty float64 `json:"difficulty"`
}

type RecommendationRequest struct {
	UserID       uint              `json:"user_id"`
	QuizID       uint              `json:"quiz_id"`
	History      []UserHistoryItem `json:"history"`
	PriorAbility *float64          `json:"prior_ability,omitempty"`
}

type RecommendationResponse struct {
	TargetDifficulty float64 `json:"target_difficulty"`
	PredictedAbility float64 `json:"predicted_ability"`
	Message          string  `json:"message"`
}

type EssayGradingRequest struct {
	QuestionText  string `json:"question_text"`
	TeacherKey    string `json:"teacher_key"`
	StudentAnswer string `json:"student_answer"`
}

type EssayGradingResponse struct {
	ScoreFinal float64                `json:"score_final"`
	Feedback   string                 `json:"feedback"`
	Debug      map[string]interface{} `json:"debug"`
}

// MLClient handles communication with the Python ML Service
type MLClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewMLClient(baseURL string) *MLClient {
	return &MLClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *MLClient) GetRecommendation(userID, quizID uint, history []UserHistoryItem, priorAbility *float64) (*RecommendationResponse, error) {
	reqBody := RecommendationRequest{
		UserID:       userID,
		QuizID:       quizID,
		History:      history,
		PriorAbility: priorAbility,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/recommend", c.BaseURL)
	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorBody)
		return nil, fmt.Errorf("ml-service returned status: %d, body: %v", resp.StatusCode, errorBody)
	}

	var result RecommendationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *MLClient) GradeEssay(questionText, teacherKey, studentAnswer string) (*EssayGradingResponse, error) {
	reqBody := EssayGradingRequest{
		QuestionText:  questionText,
		TeacherKey:    teacherKey,
		StudentAnswer: studentAnswer,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/grade-essay", c.BaseURL)
	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorBody)
		return nil, fmt.Errorf("ml-service grade-essay returned status: %d, body: %v", resp.StatusCode, errorBody)
	}

	var result EssayGradingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
