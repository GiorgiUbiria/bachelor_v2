package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"bachelor_backend/database"
	"bachelor_backend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestLoggingConfig defines the configuration for request logging middleware
type RequestLoggingConfig struct {
	// Skip defines a function to skip middleware
	Skip func(c *fiber.Ctx) bool
	// LogSensitiveData determines if sensitive data should be logged
	LogSensitiveData bool
	// MaxBodySize defines the maximum body size to log (in bytes)
	MaxBodySize int
	// SampleRate defines the sampling rate (0.0 to 1.0, 1.0 = log all requests)
	SampleRate float64
}

// DefaultRequestLoggingConfig is the default configuration
var DefaultRequestLoggingConfig = RequestLoggingConfig{
	Skip: func(c *fiber.Ctx) bool {
		// Skip health checks and static files
		path := c.Path()
		return strings.HasPrefix(path, "/health") ||
			strings.HasPrefix(path, "/swagger") ||
			strings.HasPrefix(path, "/static") ||
			strings.HasPrefix(path, "/favicon")
	},
	LogSensitiveData: false,
	MaxBodySize:      4096, // 4KB
	SampleRate:       1.0,  // Log all requests
}

// RequestLogging creates a new request logging middleware
func RequestLogging(config ...RequestLoggingConfig) fiber.Handler {
	cfg := DefaultRequestLoggingConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *fiber.Ctx) error {
		// Skip if configured to skip
		if cfg.Skip != nil && cfg.Skip(c) {
			return c.Next()
		}

		// Sample requests based on sample rate
		if cfg.SampleRate < 1.0 {
			// Simple sampling - in production, use better random sampling
			if time.Now().UnixNano()%100 >= int64(cfg.SampleRate*100) {
				return c.Next()
			}
		}

		// Capture request start time
		startTime := time.Now()

		// Get request details
		method := c.Method()
		path := c.Path()
		ipAddress := getRealIP(c)
		userAgent := c.Get("User-Agent")

		// Get user ID if available
		var userID *uuid.UUID
		if id, exists := GetUserID(c); exists && id != uuid.Nil {
			userID = &id
		}

		// Get session ID
		sessionID := getSessionID(c)

		// Capture query parameters
		queryParams := captureQueryParams(c, cfg.LogSensitiveData)

		// Get request size
		requestSize := len(c.Body())

		// Continue with request processing
		err := c.Next()

		// Capture response details
		statusCode := c.Response().StatusCode()
		responseSize := len(c.Response().Body())
		responseTime := float64(time.Since(startTime).Nanoseconds()) / 1e6 // Convert to milliseconds

		// Log request asynchronously to avoid blocking
		go func() {
			if err := logRequest(models.RequestLog{
				UserID:       userID,
				IPAddress:    ipAddress,
				UserAgent:    userAgent,
				Method:       method,
				Path:         path,
				QueryParams:  queryParams,
				StatusCode:   statusCode,
				ResponseTime: responseTime,
				RequestSize:  requestSize,
				ResponseSize: responseSize,
				SessionID:    sessionID,
				Timestamp:    startTime,
			}); err != nil {
				log.Printf("Failed to log request: %v", err)
			}
		}()

		return err
	}
}

// getRealIP extracts the real IP address from the request
func getRealIP(c *fiber.Ctx) string {
	// Check X-Forwarded-For header first (for load balancers/proxies)
	xForwardedFor := c.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	xRealIP := c.Get("X-Real-IP")
	if xRealIP != "" && net.ParseIP(xRealIP) != nil {
		return xRealIP
	}

	// Check X-Forwarded header
	xForwarded := c.Get("X-Forwarded")
	if xForwarded != "" {
		// Extract IP from X-Forwarded header
		parts := strings.Split(xForwarded, ";")
		for _, part := range parts {
			if strings.HasPrefix(strings.TrimSpace(part), "for=") {
				ip := strings.TrimSpace(strings.TrimPrefix(part, "for="))
				// Remove quotes if present
				ip = strings.Trim(ip, "\"")
				if net.ParseIP(ip) != nil {
					return ip
				}
			}
		}
	}

	// Fall back to remote IP
	return c.IP()
}

// getSessionID extracts or generates a session ID
func getSessionID(c *fiber.Ctx) string {
	// Try to get session ID from header
	sessionID := c.Get("X-Session-ID")
	if sessionID != "" {
		return sessionID
	}

	// Try to get from cookie
	sessionID = c.Cookies("session_id")
	if sessionID != "" {
		return sessionID
	}

	// Generate a simple session ID based on IP and User-Agent
	// In production, use proper session management
	ip := c.IP()
	userAgent := c.Get("User-Agent")
	return fmt.Sprintf("%x", []byte(ip+userAgent))[:16]
}

// captureQueryParams captures query parameters, optionally filtering sensitive data
func captureQueryParams(c *fiber.Ctx, logSensitive bool) string {
	if c.Request().URI().QueryString() == nil {
		return "{}"
	}

	queryString := string(c.Request().URI().QueryString())
	if queryString == "" {
		return "{}"
	}

	// Parse query parameters
	params := make(map[string]interface{})

	// Simple query parameter parsing
	pairs := strings.Split(queryString, "&")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)
		key := parts[0]
		value := ""
		if len(parts) > 1 {
			value = parts[1]
		}

		// Filter sensitive parameters if not logging sensitive data
		if !logSensitive && isSensitiveParam(key) {
			value = "[FILTERED]"
		}

		params[key] = value
	}

	// Convert to JSON
	jsonBytes, err := json.Marshal(params)
	if err != nil {
		log.Printf("Failed to marshal query params: %v", err)
		return "{}"
	}

	return string(jsonBytes)
}

// isSensitiveParam checks if a parameter is considered sensitive
func isSensitiveParam(param string) bool {
	sensitiveParams := []string{
		"password", "passwd", "pwd", "secret", "token", "key",
		"api_key", "apikey", "auth", "authorization", "session",
		"credit_card", "creditcard", "cc", "ssn", "social_security",
	}

	paramLower := strings.ToLower(param)
	for _, sensitive := range sensitiveParams {
		if strings.Contains(paramLower, sensitive) {
			return true
		}
	}
	return false
}

// logRequest saves the request log to the database
func logRequest(requestLog models.RequestLog) error {
	// Set ID if not set
	if requestLog.ID == uuid.Nil {
		requestLog.ID = uuid.New()
	}

	// Save to database
	result := database.DB.Create(&requestLog)
	if result.Error != nil {
		return fmt.Errorf("failed to save request log: %w", result.Error)
	}

	// Trigger anomaly analysis asynchronously
	go func() {
		// Import the services package to access AnomalyServiceInstance
		// Note: This creates a circular import, so we'll handle it differently
		// For now, we'll skip the automatic analysis and handle it separately

		// In a production system, you might use a message queue or event system
		// to decouple the request logging from anomaly analysis
	}()

	return nil
}

// RequestMetrics provides basic request metrics tracking
func RequestMetrics() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process the request
		err := c.Next()

		// Calculate response time
		responseTime := time.Since(start).Milliseconds()

		// Track metrics asynchronously to avoid blocking the request
		go func() {
			updateDailyMetrics(c, responseTime)
		}()

		return err
	}
}

// updateDailyMetrics updates daily security metrics
func updateDailyMetrics(c *fiber.Ctx, responseTime int64) {
	// Add safety checks to prevent nil pointer dereference
	if c == nil {
		log.Printf("Warning: Context is nil in updateDailyMetrics")
		return
	}

	// Check if response is available
	response := c.Response()
	if response == nil {
		log.Printf("Warning: Response is nil in updateDailyMetrics")
		return
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)

	// Get or create today's security metrics
	var metrics models.SecurityMetrics
	result := database.DB.Where("date = ?", today).First(&metrics)

	statusCode := response.StatusCode()

	if result.Error != nil {
		// Create new metrics record for today
		metrics = models.SecurityMetrics{
			Date:            today,
			TotalRequests:   1,
			UniqueIPs:       1,
			AvgResponseTime: float64(responseTime),
			ErrorRate:       0.0,
		}

		// Calculate error rate for new record
		if statusCode >= 400 {
			metrics.ErrorRate = 1.0
		}

		if err := database.DB.Create(&metrics).Error; err != nil {
			log.Printf("Failed to create security metrics: %v", err)
			return
		}
	} else {
		// Update existing metrics
		oldTotal := metrics.TotalRequests
		metrics.TotalRequests++

		// Update average response time
		metrics.AvgResponseTime = (metrics.AvgResponseTime*float64(oldTotal) + float64(responseTime)) / float64(metrics.TotalRequests)

		// Update error rate
		if statusCode >= 400 {
			// Recalculate error rate
			errorCount := int(metrics.ErrorRate * float64(oldTotal))
			errorCount++
			metrics.ErrorRate = float64(errorCount) / float64(metrics.TotalRequests)
		} else {
			// Recalculate error rate without adding an error
			errorCount := int(metrics.ErrorRate * float64(oldTotal))
			metrics.ErrorRate = float64(errorCount) / float64(metrics.TotalRequests)
		}

		if err := database.DB.Save(&metrics).Error; err != nil {
			log.Printf("Failed to update security metrics: %v", err)
			return
		}
	}
}
