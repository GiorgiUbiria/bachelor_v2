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

// Sentiment Analysis Types
type SentimentAnalysisResponse struct {
	ProductID        string                   `json:"product_id"`
	ProductName      string                   `json:"product_name"`
	TotalComments    int                      `json:"total_comments"`
	SentimentSummary map[string]interface{}   `json:"sentiment_summary"`
	Insights         []map[string]interface{} `json:"insights"`
	GeneratedAt      string                   `json:"generated_at"`
}

// Auto-Tagging Types
type AutoTaggingResponse struct {
	ProductID        string    `json:"product_id"`
	ProductName      string    `json:"product_name"`
	Category         string    `json:"category"`
	ExistingTags     []string  `json:"existing_tags"`
	SuggestedTags    []string  `json:"suggested_tags"`
	ConfidenceScores []float64 `json:"confidence_scores"`
	Reasoning        []string  `json:"reasoning"`
	GeneratedAt      string    `json:"generated_at"`
}

// Smart Discounts Types
type SmartDiscountResponse struct {
	ProductID                   string                 `json:"product_id"`
	ProductName                 string                 `json:"product_name"`
	Category                    string                 `json:"category"`
	CurrentPerformanceScore     float64                `json:"current_performance_score"`
	Recommendation              string                 `json:"recommendation"`
	SuggestedDiscountPercentage float64                `json:"suggested_discount_percentage"`
	DiscountType                string                 `json:"discount_type"`
	ExpectedImpact              map[string]interface{} `json:"expected_impact"`
	Reasoning                   string                 `json:"reasoning"`
	SeasonalFactor              float64                `json:"seasonal_factor"`
	Urgency                     string                 `json:"urgency"`
	GeneratedAt                 string                 `json:"generated_at"`
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

// AnalyzeProductSentiment calls the ML service to analyze product sentiment
func (ml *MLClient) AnalyzeProductSentiment(productID uuid.UUID) (*SentimentAnalysisResponse, error) {
	url := fmt.Sprintf("%s/sentiment/product/%s", ml.baseURL, productID.String())
	resp, err := ml.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ML service returned status %d: %s", resp.StatusCode, string(body))
	}

	var result SentimentAnalysisResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// AnalyzeCategorySentiment calls the ML service to analyze category sentiment
func (ml *MLClient) AnalyzeCategorySentiment(category string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/sentiment/category/%s", ml.baseURL, category)
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

// GetSentimentInsights calls the ML service to get sentiment insights
func (ml *MLClient) GetSentimentInsights() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/sentiment/insights", ml.baseURL)
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

// SuggestProductTags calls the ML service to suggest tags for a product
func (ml *MLClient) SuggestProductTags(productID uuid.UUID) (*AutoTaggingResponse, error) {
	url := fmt.Sprintf("%s/auto-tagging/suggest/%s", ml.baseURL, productID.String())
	resp, err := ml.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ML service returned status %d: %s", resp.StatusCode, string(body))
	}

	var result AutoTaggingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// AutoTagProducts calls the ML service to auto-tag products
func (ml *MLClient) AutoTagProducts(limit int) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/auto-tagging/auto-tag?limit=%d", ml.baseURL, limit)
	resp, err := ml.client.Post(url, "application/json", nil)
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

// GetTaggingInsights calls the ML service to get tagging insights
func (ml *MLClient) GetTaggingInsights() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/auto-tagging/insights", ml.baseURL)
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

// SuggestProductDiscount calls the ML service to suggest discount for a product
func (ml *MLClient) SuggestProductDiscount(productID uuid.UUID) (*SmartDiscountResponse, error) {
	url := fmt.Sprintf("%s/smart-discounts/suggest/product/%s", ml.baseURL, productID.String())
	resp, err := ml.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ML service returned status %d: %s", resp.StatusCode, string(body))
	}

	var result SmartDiscountResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// SuggestCategoryDiscounts calls the ML service to suggest discounts for a category
func (ml *MLClient) SuggestCategoryDiscounts(category string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/smart-discounts/suggest/category/%s", ml.baseURL, category)
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

// GetDiscountInsights calls the ML service to get discount insights
func (ml *MLClient) GetDiscountInsights() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/smart-discounts/insights", ml.baseURL)
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

// InitializeMLServices calls the ML service to initialize all new ML models
func (ml *MLClient) InitializeMLServices() error {
	services := []string{"sentiment", "auto-tagging", "smart-discounts"}

	for _, service := range services {
		url := fmt.Sprintf("%s/%s/initialize", ml.baseURL, service)
		resp, err := ml.client.Post(url, "application/json", nil)
		if err != nil {
			return fmt.Errorf("failed to initialize %s service: %w", service, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("%s service returned status %d", service, resp.StatusCode)
		}
	}

	return nil
}

// Global ML client instance
var MLService = NewMLClient()
