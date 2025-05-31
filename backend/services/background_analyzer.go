package services

import (
	"log"
	"time"

	"bachelor_backend/database"
	"bachelor_backend/models"
)

// BackgroundAnalyzer handles periodic anomaly analysis of request logs
type BackgroundAnalyzer struct {
	anomalyService *AnomalyService
	ticker         *time.Ticker
	stopChan       chan bool
	isRunning      bool
}

// NewBackgroundAnalyzer creates a new background analyzer
func NewBackgroundAnalyzer() *BackgroundAnalyzer {
	return &BackgroundAnalyzer{
		anomalyService: nil, // Will be set when needed
		stopChan:       make(chan bool),
		isRunning:      false,
	}
}

// getAnomalyService returns the anomaly service instance, initializing it if needed
func (ba *BackgroundAnalyzer) getAnomalyService() *AnomalyService {
	if ba.anomalyService == nil {
		ba.anomalyService = AnomalyServiceInstance
	}
	return ba.anomalyService
}

// Start begins the background analysis process
func (ba *BackgroundAnalyzer) Start(intervalMinutes int) {
	if ba.isRunning {
		log.Println("Background analyzer is already running")
		return
	}

	// Ensure we have a valid anomaly service
	if ba.getAnomalyService() == nil {
		log.Println("Warning: Anomaly service not available, background analyzer will not start")
		return
	}

	ba.ticker = time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	ba.isRunning = true

	log.Printf("Starting background anomaly analyzer with %d minute intervals", intervalMinutes)

	go func() {
		// Run initial analysis
		ba.analyzeRecentRequests()

		for {
			select {
			case <-ba.ticker.C:
				ba.analyzeRecentRequests()
			case <-ba.stopChan:
				ba.ticker.Stop()
				ba.isRunning = false
				log.Println("Background analyzer stopped")
				return
			}
		}
	}()
}

// Stop stops the background analysis process
func (ba *BackgroundAnalyzer) Stop() {
	if !ba.isRunning {
		return
	}

	ba.stopChan <- true
}

// analyzeRecentRequests analyzes recent request logs that haven't been analyzed yet
func (ba *BackgroundAnalyzer) analyzeRecentRequests() {
	log.Println("Starting background anomaly analysis...")

	// Ensure we have a valid anomaly service
	anomalyService := ba.getAnomalyService()
	if anomalyService == nil {
		log.Println("Warning: Anomaly service not available, skipping analysis")
		return
	}

	// Get recent request logs that haven't been analyzed
	// We'll look for requests from the last 10 minutes that don't have corresponding anomaly alerts
	cutoffTime := time.Now().Add(-10 * time.Minute)

	var requestLogs []models.RequestLog
	result := database.DB.
		Where("timestamp >= ?", cutoffTime).
		Where("id NOT IN (SELECT COALESCE(request_log_id, '00000000-0000-0000-0000-000000000000') FROM anomaly_alerts WHERE request_log_id IS NOT NULL)").
		Order("timestamp DESC").
		Limit(100). // Limit to avoid overwhelming the ML service
		Find(&requestLogs)

	if result.Error != nil {
		log.Printf("Failed to fetch recent request logs: %v", result.Error)
		return
	}

	if len(requestLogs) == 0 {
		log.Println("No new request logs to analyze")
		return
	}

	log.Printf("Analyzing %d recent request logs", len(requestLogs))

	// Analyze each request log
	analyzed := 0
	errors := 0

	for _, requestLog := range requestLogs {
		analysis, err := anomalyService.AnalyzeRequest(requestLog)
		if err != nil {
			log.Printf("Failed to analyze request %s: %v", requestLog.ID, err)
			errors++
			continue
		}

		// Process the analysis result
		if err := anomalyService.ProcessAnomalyAlert(requestLog, analysis); err != nil {
			log.Printf("Failed to process anomaly alert for request %s: %v", requestLog.ID, err)
			errors++
			continue
		}

		analyzed++

		// Add a small delay to avoid overwhelming the ML service
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Background analysis completed: %d analyzed, %d errors", analyzed, errors)
}

// GetStatus returns the current status of the background analyzer
func (ba *BackgroundAnalyzer) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"is_running":   ba.isRunning,
		"last_run":     time.Now().Format(time.RFC3339),
		"service_name": "background_anomaly_analyzer",
	}
}

// AnalyzeSpecificRequest analyzes a specific request log immediately
func (ba *BackgroundAnalyzer) AnalyzeSpecificRequest(requestLogID string) error {
	var requestLog models.RequestLog
	result := database.DB.Where("id = ?", requestLogID).First(&requestLog)
	if result.Error != nil {
		return result.Error
	}

	analysis, err := ba.getAnomalyService().AnalyzeRequest(requestLog)
	if err != nil {
		return err
	}

	return ba.getAnomalyService().ProcessAnomalyAlert(requestLog, analysis)
}

// Global background analyzer instance
var BackgroundAnalyzerInstance = NewBackgroundAnalyzer()
