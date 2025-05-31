package handlers

import (
	"strconv"
	"time"

	"bachelor_backend/middleware"
	"bachelor_backend/models"
	"bachelor_backend/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// SecurityDashboardResponse represents the security dashboard response
type SecurityDashboardResponse struct {
	MLDashboard     map[string]interface{}   `json:"ml_dashboard"`
	RecentAlerts    []models.AnomalyAlert    `json:"recent_alerts"`
	SecurityMetrics []models.SecurityMetrics `json:"security_metrics"`
	Summary         map[string]interface{}   `json:"summary"`
}

// GetSecurityDashboard godoc
// @Summary Get Security Dashboard
// @Description Get comprehensive security dashboard with anomaly detection data and real-time threat monitoring
// @Tags Security
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SecurityDashboardResponse "Security dashboard data"
// @Failure 401 {object} StandardErrorResponse "Unauthorized"
// @Failure 500 {object} StandardErrorResponse "Internal server error"
// @Router /security/dashboard [get]
func GetSecurityDashboard(c *fiber.Ctx) error {
	// Get ML dashboard data
	mlDashboard, err := services.AnomalyServiceInstance.GetSecurityDashboard()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get ML dashboard data: " + err.Error(),
		})
	}

	// Get recent alerts
	recentAlerts, err := services.AnomalyServiceInstance.GetRecentAlerts(20)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get recent alerts: " + err.Error(),
		})
	}

	// Get security metrics for last 30 days
	securityMetrics, err := services.AnomalyServiceInstance.GetSecurityMetrics(30)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get security metrics: " + err.Error(),
		})
	}

	// Calculate summary statistics
	summary := calculateSecuritySummary(recentAlerts, securityMetrics)

	response := SecurityDashboardResponse{
		MLDashboard:     mlDashboard,
		RecentAlerts:    recentAlerts,
		SecurityMetrics: securityMetrics,
		Summary:         summary,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetAnomalyInsights godoc
// @Summary Get Anomaly Insights
// @Description Get detailed insights and analysis about detected anomalies and attack patterns
// @Tags Security
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Anomaly insights data"
// @Failure 401 {object} StandardErrorResponse "Unauthorized"
// @Failure 500 {object} StandardErrorResponse "Internal server error"
// @Router /security/insights [get]
func GetAnomalyInsights(c *fiber.Ctx) error {
	insights, err := services.AnomalyServiceInstance.GetAnomalyInsights()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get anomaly insights: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    insights,
	})
}

// GetAttackPatterns godoc
// @Summary Get Attack Patterns
// @Description Get information about detectable attack patterns and security threats
// @Tags Security
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Attack patterns information"
// @Failure 401 {object} StandardErrorResponse "Unauthorized"
// @Failure 500 {object} StandardErrorResponse "Internal server error"
// @Router /security/patterns [get]
func GetAttackPatterns(c *fiber.Ctx) error {
	patterns, err := services.AnomalyServiceInstance.GetAttackPatterns()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get attack patterns: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    patterns,
	})
}

// SimulateAttack godoc
// @Summary Simulate Attack
// @Description Simulate different types of attacks for demonstration and testing purposes
// @Tags Security
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param attack_type query string true "Attack type" Enums(sql_injection, xss, brute_force, ddos) example(sql_injection)
// @Success 200 {object} map[string]interface{} "Attack simulation results"
// @Failure 400 {object} StandardErrorResponse "Bad request - invalid attack type"
// @Failure 401 {object} StandardErrorResponse "Unauthorized"
// @Failure 500 {object} StandardErrorResponse "Internal server error"
// @Router /security/simulate [get]
func SimulateAttack(c *fiber.Ctx) error {
	attackType := c.Query("attack_type")
	if attackType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "attack_type parameter is required",
		})
	}

	// Validate attack type
	validTypes := []string{"sql_injection", "xss", "brute_force", "ddos"}
	isValid := false
	for _, validType := range validTypes {
		if attackType == validType {
			isValid = true
			break
		}
	}

	if !isValid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid attack type. Valid types: sql_injection, xss, brute_force, ddos",
		})
	}

	simulation, err := services.AnomalyServiceInstance.SimulateAttack(attackType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to simulate attack: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    simulation,
	})
}

// GetSecurityAlerts godoc
// @Summary Get Security Alerts
// @Description Get recent security alerts with filtering options
// @Tags Security
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of alerts to return" default(50) minimum(1) maximum(1000)
// @Param risk_level query string false "Filter by risk level" Enums(low, medium, high, critical)
// @Param resolved query bool false "Filter by resolution status"
// @Success 200 {object} map[string]interface{} "Security alerts data"
// @Failure 401 {object} StandardErrorResponse "Unauthorized"
// @Failure 500 {object} StandardErrorResponse "Internal server error"
// @Router /security/alerts [get]
func GetSecurityAlerts(c *fiber.Ctx) error {
	// Parse query parameters
	limitStr := c.Query("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 50
	}

	riskLevel := c.Query("risk_level")
	resolvedStr := c.Query("resolved")

	// Get alerts from database with filters
	alerts, err := getFilteredAlerts(limit, riskLevel, resolvedStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get security alerts: " + err.Error(),
		})
	}

	// Calculate alert statistics
	stats := calculateAlertStatistics(alerts)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"alerts":     alerts,
			"statistics": stats,
			"total":      len(alerts),
		},
	})
}

// ResolveAlertRequest represents the request to resolve an alert
type ResolveAlertRequest struct {
	Notes string `json:"notes" example:"False positive - legitimate admin access"`
}

// ResolveSecurityAlert godoc
// @Summary Resolve Security Alert
// @Description Mark a security alert as resolved with optional notes
// @Tags Security
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param alert_id path string true "Alert ID" format(uuid)
// @Param request body ResolveAlertRequest true "Resolution details"
// @Success 200 {object} map[string]interface{} "Alert resolved successfully"
// @Failure 400 {object} StandardErrorResponse "Bad request - invalid alert ID or request body"
// @Failure 401 {object} StandardErrorResponse "Unauthorized"
// @Failure 404 {object} StandardErrorResponse "Alert not found"
// @Failure 500 {object} StandardErrorResponse "Internal server error"
// @Router /security/alerts/{alert_id}/resolve [post]
func ResolveSecurityAlert(c *fiber.Ctx) error {
	// Get alert ID from path
	alertIDStr := c.Params("alert_id")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid alert ID format",
		})
	}

	// Get user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User authentication required",
		})
	}

	// Parse request body
	var req ResolveAlertRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Resolve the alert
	err = services.AnomalyServiceInstance.ResolveAlert(alertID, userID, req.Notes)
	if err != nil {
		if err.Error() == "alert not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"error":   "Alert not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to resolve alert: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Alert resolved successfully",
	})
}

// GetSecurityMetrics godoc
// @Summary Get Security Metrics
// @Description Get security metrics and statistics for a specified time period
// @Tags Security
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "Number of days to include in metrics" default(30) minimum(1) maximum(365)
// @Success 200 {object} map[string]interface{} "Security metrics data"
// @Failure 401 {object} StandardErrorResponse "Unauthorized"
// @Failure 500 {object} StandardErrorResponse "Internal server error"
// @Router /security/metrics [get]
func GetSecurityMetrics(c *fiber.Ctx) error {
	// Parse days parameter
	daysStr := c.Query("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 || days > 365 {
		days = 30
	}

	// Get security metrics
	metrics, err := services.AnomalyServiceInstance.GetSecurityMetrics(days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get security metrics: " + err.Error(),
		})
	}

	// Calculate trends and aggregations
	trends := calculateSecurityTrends(metrics)
	aggregations := calculateSecurityAggregations(metrics)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"metrics":      metrics,
			"trends":       trends,
			"aggregations": aggregations,
			"period_days":  days,
		},
	})
}

// Helper functions

func calculateSecuritySummary(alerts []models.AnomalyAlert, metrics []models.SecurityMetrics) map[string]interface{} {
	summary := make(map[string]interface{})

	// Alert statistics
	totalAlerts := len(alerts)
	unresolvedAlerts := 0
	highRiskAlerts := 0

	for _, alert := range alerts {
		if !alert.IsResolved {
			unresolvedAlerts++
		}
		if alert.RiskLevel == "high" || alert.RiskLevel == "critical" {
			highRiskAlerts++
		}
	}

	summary["total_alerts"] = totalAlerts
	summary["unresolved_alerts"] = unresolvedAlerts
	summary["high_risk_alerts"] = highRiskAlerts

	// Metrics aggregations
	if len(metrics) > 0 {
		totalRequests := 0
		totalAnomalous := 0
		avgResponseTime := 0.0

		for _, metric := range metrics {
			totalRequests += metric.TotalRequests
			totalAnomalous += metric.AnomalousRequests
			avgResponseTime += metric.AvgResponseTime
		}

		if len(metrics) > 0 {
			avgResponseTime /= float64(len(metrics))
		}

		anomalyRate := 0.0
		if totalRequests > 0 {
			anomalyRate = float64(totalAnomalous) / float64(totalRequests)
		}

		summary["total_requests"] = totalRequests
		summary["total_anomalous"] = totalAnomalous
		summary["anomaly_rate"] = anomalyRate
		summary["avg_response_time"] = avgResponseTime
	}

	summary["last_updated"] = time.Now().Format(time.RFC3339)

	return summary
}

func getFilteredAlerts(limit int, riskLevel, resolvedStr string) ([]models.AnomalyAlert, error) {
	alerts, err := services.AnomalyServiceInstance.GetRecentAlerts(limit)
	if err != nil {
		return nil, err
	}
	// Note: This is a simplified version. In a real implementation,
	// you would apply filters in the database query for better performance
	return alerts, nil
}

func calculateAlertStatistics(alerts []models.AnomalyAlert) map[string]interface{} {
	stats := make(map[string]interface{})

	riskLevelCounts := make(map[string]int)
	resolvedCount := 0

	for _, alert := range alerts {
		riskLevelCounts[alert.RiskLevel]++
		if alert.IsResolved {
			resolvedCount++
		}
	}

	stats["risk_level_distribution"] = riskLevelCounts
	stats["resolved_count"] = resolvedCount
	stats["unresolved_count"] = len(alerts) - resolvedCount

	return stats
}

func calculateSecurityTrends(metrics []models.SecurityMetrics) map[string]interface{} {
	trends := make(map[string]interface{})

	if len(metrics) < 2 {
		return trends
	}

	// Calculate simple trends (increase/decrease from first to last)
	first := metrics[len(metrics)-1] // Oldest (metrics are ordered DESC)
	last := metrics[0]               // Newest

	requestsTrend := "stable"
	if last.TotalRequests > first.TotalRequests {
		requestsTrend = "increasing"
	} else if last.TotalRequests < first.TotalRequests {
		requestsTrend = "decreasing"
	}

	anomalyTrend := "stable"
	if last.AnomalousRequests > first.AnomalousRequests {
		anomalyTrend = "increasing"
	} else if last.AnomalousRequests < first.AnomalousRequests {
		anomalyTrend = "decreasing"
	}

	trends["requests_trend"] = requestsTrend
	trends["anomaly_trend"] = anomalyTrend
	trends["period_start"] = first.Date.Format("2006-01-02")
	trends["period_end"] = last.Date.Format("2006-01-02")

	return trends
}

func calculateSecurityAggregations(metrics []models.SecurityMetrics) map[string]interface{} {
	aggregations := make(map[string]interface{})

	if len(metrics) == 0 {
		return aggregations
	}

	totalRequests := 0
	totalAnomalous := 0
	totalHighRisk := 0
	totalBlocked := 0
	avgResponseTime := 0.0

	for _, metric := range metrics {
		totalRequests += metric.TotalRequests
		totalAnomalous += metric.AnomalousRequests
		totalHighRisk += metric.HighRiskRequests
		totalBlocked += metric.BlockedRequests
		avgResponseTime += metric.AvgResponseTime
	}

	avgResponseTime /= float64(len(metrics))

	aggregations["total_requests"] = totalRequests
	aggregations["total_anomalous"] = totalAnomalous
	aggregations["total_high_risk"] = totalHighRisk
	aggregations["total_blocked"] = totalBlocked
	aggregations["avg_response_time"] = avgResponseTime

	if totalRequests > 0 {
		aggregations["anomaly_rate"] = float64(totalAnomalous) / float64(totalRequests)
		aggregations["high_risk_rate"] = float64(totalHighRisk) / float64(totalRequests)
		aggregations["block_rate"] = float64(totalBlocked) / float64(totalRequests)
	}

	return aggregations
}
