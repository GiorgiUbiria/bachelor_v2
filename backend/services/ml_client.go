package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

type MLClient struct {
	baseURL string
	client  *http.Client
}

type RecommendationRequest struct {
	UserID    string `json:"user_id"`
	Algorithm string `json:"algorithm"`
	Limit     int    `json:"limit"`
}

type RecommendationResponse struct {
	ProductID string  `json:"product_id"`
	Score     float64 `json:"score"`
	Algorithm string  `json:"algorithm"`
}

type RecommendationsResponse struct {
	Recommendations []RecommendationResponse `json:"recommendations"`
	UserID          string                   `json:"user_id"`
	Algorithm       string                   `json:"algorithm"`
	Total           int                      `json:"total"`
}

// NewMLClient creates a new ML service client
func NewMLClient() *MLClient {
	baseURL := os.Getenv("ML_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}

	return &MLClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateRecommendations calls the ML service to generate recommendations
func (ml *MLClient) GenerateRecommendations(userID uuid.UUID, algorithm string, limit int) (*RecommendationsResponse, error) {
	reqBody := RecommendationRequest{
		UserID:    userID.String(),
		Algorithm: algorithm,
		Limit:     limit,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/generate", ml.baseURL)
	resp, err := ml.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ML service returned status %d: %s", resp.StatusCode, string(body))
	}

	var result RecommendationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// TrainModels calls the ML service to train recommendation models
func (ml *MLClient) TrainModels() error {
	url := fmt.Sprintf("%s/train", ml.baseURL)
	resp, err := ml.client.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ML service returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetMLStatus checks the status of ML models
func (ml *MLClient) GetMLStatus() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/status", ml.baseURL)
	resp, err := ml.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ML service returned status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// Global ML client instance
var MLService = NewMLClient()
