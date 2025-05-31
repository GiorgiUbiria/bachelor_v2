package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"bachelor_backend/database"
	"bachelor_backend/models"

	"github.com/google/uuid"
)

// AnomalyService handles communication with ML anomaly detection service
type AnomalyService struct {
	mlServiceURL string
	httpClient   *http.Client
}

// NewAnomalyService creates a new anomaly service instance
func NewAnomalyService() *AnomalyService {
	mlServiceURL := os.Getenv("ML_SERVICE_URL")
	if mlServiceURL == "" {
		mlServiceURL = "http://localhost:8000"
	}

	return &AnomalyService{
		mlServiceURL: mlServiceURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RequestAnalysisRequest represents the request structure for ML analysis
type RequestAnalysisRequest struct {
	IPAddress    string  `json:"ip_address"`
	UserAgent    string  `json:"user_agent"`
	Method       string  `json:"method"`
	Path         string  `json:"path"`
	QueryParams  string  `json:"query_params"`
	StatusCode   int     `json:"status_code"`
	ResponseTime float64 `json:"response_time"`
	RequestSize  int     `json:"request_size"`
	ResponseSize int     `json:"response_size"`
	UserID       string  `json:"user_id,omitempty"`
}

// AnomalyAnalysisResponse represents the response from ML analysis
type AnomalyAnalysisResponse struct {
	Success bool `json:"success"`
	Data    struct {
		RequestID         string                 `json:"request_id"`
		AnomalyScore      float64                `json:"anomaly_score"`
		RiskLevel         string                 `json:"risk_level"`
		IsAnomaly         bool                   `json:"is_anomaly"`
		AnomalyReasons    []string               `json:"anomaly_reasons"`
		Recommendations   []string               `json:"recommendations"`
		AnalysisTimestamp string                 `json:"analysis_timestamp"`
		AnalysisDetails   map[string]interface{} `json:"analysis_details"`
	} `json:"data"`
}

// MLServiceResponse represents a generic ML service response
type MLServiceResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
}

// AnalyzeRequest sends a request to the ML service for anomaly analysis
func (as *AnomalyService) AnalyzeRequest(requestLog models.RequestLog) (*AnomalyAnalysisResponse, error) {
	// Prepare request data
	analysisRequest := RequestAnalysisRequest{
		IPAddress:    requestLog.IPAddress,
		UserAgent:    requestLog.UserAgent,
		Method:       requestLog.Method,
		Path:         requestLog.Path,
		QueryParams:  requestLog.QueryParams,
		StatusCode:   requestLog.StatusCode,
		ResponseTime: requestLog.ResponseTime,
		RequestSize:  requestLog.RequestSize,
		ResponseSize: requestLog.ResponseSize,
	}

	// Add user ID if available
	if requestLog.UserID != nil {
		analysisRequest.UserID = requestLog.UserID.String()
	}

	// Convert to JSON
	jsonData, err := json.Marshal(analysisRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	// Make HTTP request to ML service
	url := fmt.Sprintf("%s/anomaly-detection/analyze", as.mlServiceURL)
	resp, err := as.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var analysisResponse AnomalyAnalysisResponse
	if err := json.Unmarshal(body, &analysisResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !analysisResponse.Success {
		return nil, fmt.Errorf("ML service analysis failed")
	}

	return &analysisResponse, nil
}

// ProcessAnomalyAlert processes an anomaly detection result and creates alerts if necessary
func (as *AnomalyService) ProcessAnomalyAlert(requestLog models.RequestLog, analysis *AnomalyAnalysisResponse) error {
	// Only create alerts for significant anomalies
	if !analysis.Data.IsAnomaly || analysis.Data.AnomalyScore < 0.5 {
		return nil
	}

	// Convert anomaly reasons to JSON
	reasonsJSON, err := json.Marshal(analysis.Data.AnomalyReasons)
	if err != nil {
		log.Printf("Failed to marshal anomaly reasons: %v", err)
		reasonsJSON = []byte("[]")
	}

	// Create anomaly alert
	alert := models.AnomalyAlert{
		ID:             uuid.New(),
		RequestLogID:   &requestLog.ID,
		UserID:         requestLog.UserID,
		IPAddress:      requestLog.IPAddress,
		AnomalyScore:   analysis.Data.AnomalyScore,
		RiskLevel:      analysis.Data.RiskLevel,
		AnomalyReasons: string(reasonsJSON),
		IsResolved:     false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Save alert to database
	result := database.DB.Create(&alert)
	if result.Error != nil {
		return fmt.Errorf("failed to create anomaly alert: %w", result.Error)
	}

	// Log the alert
	log.Printf("Anomaly alert created: ID=%s, IP=%s, Score=%.3f, Risk=%s",
		alert.ID, alert.IPAddress, alert.AnomalyScore, alert.RiskLevel)

	// Update security metrics
	as.updateAnomalyMetrics(analysis.Data.RiskLevel)

	return nil
}

// AnalyzeRequestAsync analyzes a request asynchronously
func (as *AnomalyService) AnalyzeRequestAsync(requestLog models.RequestLog) {
	go func() {
		analysis, err := as.AnalyzeRequest(requestLog)
		if err != nil {
			log.Printf("Failed to analyze request %s: %v", requestLog.ID, err)
			return
		}

		if err := as.ProcessAnomalyAlert(requestLog, analysis); err != nil {
			log.Printf("Failed to process anomaly alert for request %s: %v", requestLog.ID, err)
		}
	}()
}

// GetSecurityDashboard retrieves security dashboard data from ML service
func (as *AnomalyService) GetSecurityDashboard() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/anomaly-detection/dashboard", as.mlServiceURL)
	resp, err := as.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML service returned status %d: %s", resp.StatusCode, string(body))
	}

	var response MLServiceResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("ML service request failed: %s", response.Error)
	}

	dashboardData, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid dashboard data format")
	}

	return dashboardData, nil
}

// GetAnomalyInsights retrieves anomaly insights from ML service
func (as *AnomalyService) GetAnomalyInsights() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/anomaly-detection/insights", as.mlServiceURL)
	resp, err := as.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML service returned status %d: %s", resp.StatusCode, string(body))
	}

	var response MLServiceResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("ML service request failed: %s", response.Error)
	}

	insightsData, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid insights data format")
	}

	return insightsData, nil
}

// SimulateAttack simulates different types of attacks for demonstration
func (as *AnomalyService) SimulateAttack(attackType string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/anomaly-detection/simulate/attack?attack_type=%s", as.mlServiceURL, attackType)
	resp, err := as.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML service returned status %d: %s", resp.StatusCode, string(body))
	}

	var response MLServiceResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("ML service request failed: %s", response.Error)
	}

	simulationData, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid simulation data format")
	}

	return simulationData, nil
}

// GetAttackPatterns retrieves information about detectable attack patterns
func (as *AnomalyService) GetAttackPatterns() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/anomaly-detection/patterns", as.mlServiceURL)
	resp, err := as.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML service returned status %d: %s", resp.StatusCode, string(body))
	}

	var response MLServiceResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("ML service request failed: %s", response.Error)
	}

	patternsData, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid patterns data format")
	}

	return patternsData, nil
}

// updateAnomalyMetrics updates security metrics with anomaly information
func (as *AnomalyService) updateAnomalyMetrics(riskLevel string) {
	today := time.Now().Truncate(24 * time.Hour)

	// Get or create today's metrics
	var metrics models.SecurityMetrics
	result := database.DB.Where("date = ?", today).First(&metrics)

	if result.Error != nil {
		// Create new metrics record
		metrics = models.SecurityMetrics{
			ID:                uuid.New(),
			Date:              today,
			AnomalousRequests: 1,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		if riskLevel == "high" || riskLevel == "critical" {
			metrics.HighRiskRequests = 1
		}

		database.DB.Create(&metrics)
	} else {
		// Update existing metrics
		metrics.AnomalousRequests++

		if riskLevel == "high" || riskLevel == "critical" {
			metrics.HighRiskRequests++
		}

		metrics.UpdatedAt = time.Now()
		database.DB.Save(&metrics)
	}
}

// GetRecentAlerts retrieves recent anomaly alerts from the database
func (as *AnomalyService) GetRecentAlerts(limit int) ([]models.AnomalyAlert, error) {
	var alerts []models.AnomalyAlert

	result := database.DB.
		Preload("RequestLog").
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Find(&alerts)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve alerts: %w", result.Error)
	}

	return alerts, nil
}

// ResolveAlert marks an anomaly alert as resolved
func (as *AnomalyService) ResolveAlert(alertID uuid.UUID, resolvedBy uuid.UUID, notes string) error {
	now := time.Now()

	result := database.DB.Model(&models.AnomalyAlert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"is_resolved": true,
			"resolved_at": &now,
			"resolved_by": &resolvedBy,
			"notes":       notes,
			"updated_at":  now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to resolve alert: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("alert not found")
	}

	return nil
}

// GetSecurityMetrics retrieves security metrics for a date range
func (as *AnomalyService) GetSecurityMetrics(days int) ([]models.SecurityMetrics, error) {
	var metrics []models.SecurityMetrics

	startDate := time.Now().AddDate(0, 0, -days).Truncate(24 * time.Hour)

	result := database.DB.
		Where("date >= ?", startDate).
		Order("date DESC").
		Find(&metrics)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve security metrics: %w", result.Error)
	}

	return metrics, nil
}

// Global anomaly service instance
var AnomalyServiceInstance *AnomalyService

// InitializeAnomalyService initializes the global anomaly service instance
func InitializeAnomalyService() {
	AnomalyServiceInstance = NewAnomalyService()
	log.Println("Anomaly service initialized successfully")
}
