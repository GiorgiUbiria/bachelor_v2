package handlers

import (
	"fmt"
	"strconv"
	"time"

	"bachelor_backend/database"
	"bachelor_backend/middleware"
	"bachelor_backend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GetDashboard returns comprehensive dashboard analytics
// @Summary Get dashboard analytics
// @Description Get comprehensive dashboard analytics including user statistics and recent interactions
// @Tags Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Dashboard analytics retrieved successfully"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Router /analytics/dashboard [get]
func GetDashboard(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get user statistics
	var userStats struct {
		TotalOrders   int64   `json:"total_orders"`
		TotalSpent    float64 `json:"total_spent"`
		TotalProducts int64   `json:"total_products"`
		RecentOrders  int64   `json:"recent_orders"`
	}

	// Total orders and spent
	database.DB.Model(&models.Order{}).
		Where("user_id = ?", userID).
		Count(&userStats.TotalOrders)

	database.DB.Model(&models.Order{}).
		Where("user_id = ? AND status IN ?", userID, []string{"delivered", "completed"}).
		Select("COALESCE(SUM(total), 0)").
		Scan(&userStats.TotalSpent)

	// Total products in cart
	database.DB.Table("cart_items").
		Joins("JOIN shopping_carts ON cart_items.cart_id = shopping_carts.id").
		Where("shopping_carts.user_id = ?", userID).
		Count(&userStats.TotalProducts)

	// Recent orders (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	database.DB.Model(&models.Order{}).
		Where("user_id = ? AND created_at >= ?", userID, thirtyDaysAgo).
		Count(&userStats.RecentOrders)

	// Get recent interactions
	var recentInteractions []models.UserInteraction
	database.DB.Where("user_id = ?", userID).
		Preload("Product").
		Order("created_at DESC").
		Limit(10).
		Find(&recentInteractions)

	return c.JSON(fiber.Map{
		"user_stats":          userStats,
		"recent_interactions": recentInteractions,
		"timestamp":           time.Now(),
	})
}

// GetUserAnalytics returns detailed user analytics
// @Summary Get user analytics
// @Description Get detailed user analytics including interaction stats, category preferences, and spending patterns
// @Tags Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "Number of days to analyze" default(30)
// @Success 200 {object} map[string]interface{} "User analytics retrieved successfully"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /analytics/user [get]
func GetUserAnalytics(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	days, _ := strconv.Atoi(c.Query("days", "30"))
	cutoffDate := time.Now().AddDate(0, 0, -days)

	// Interaction analytics
	var interactionStats []struct {
		InteractionType string `json:"interaction_type"`
		Count           int64  `json:"count"`
	}

	database.DB.Model(&models.UserInteraction{}).
		Select("interaction_type, COUNT(*) as count").
		Where("user_id = ? AND created_at >= ?", userID, cutoffDate).
		Group("interaction_type").
		Scan(&interactionStats)

	// Category preferences
	var categoryStats []struct {
		Category string  `json:"category"`
		Count    int64   `json:"count"`
		Revenue  float64 `json:"revenue"`
	}

	database.DB.Table("user_interactions ui").
		Select("p.category, COUNT(*) as count, COALESCE(SUM(p.price), 0) as revenue").
		Joins("JOIN products p ON ui.product_id = p.id").
		Where("ui.user_id = ? AND ui.created_at >= ?", userID, cutoffDate).
		Group("p.category").
		Scan(&categoryStats)

	// Spending over time
	var spendingOverTime []struct {
		Date   time.Time `json:"date"`
		Amount float64   `json:"amount"`
	}

	database.DB.Model(&models.Order{}).
		Select("DATE(created_at) as date, SUM(total) as amount").
		Where("user_id = ? AND created_at >= ? AND status IN ?",
			userID, cutoffDate, []string{"delivered", "completed"}).
		Group("DATE(created_at)").
		Order("date").
		Scan(&spendingOverTime)

	return c.JSON(fiber.Map{
		"interaction_stats":  interactionStats,
		"category_stats":     categoryStats,
		"spending_over_time": spendingOverTime,
		"period_days":        days,
		"generated_at":       time.Now(),
	})
}

// GetProductAnalytics returns product performance analytics
// @Summary Get product analytics
// @Description Get product performance analytics including top selling products, most viewed, and category performance
// @Tags Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "Number of days to analyze" default(30)
// @Param limit query int false "Number of items to return" default(20)
// @Success 200 {object} map[string]interface{} "Product analytics retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /analytics/products [get]
func GetProductAnalytics(c *fiber.Ctx) error {
	days, _ := strconv.Atoi(c.Query("days", "30"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	cutoffDate := time.Now().AddDate(0, 0, -days)

	// Top selling products
	var topProducts []struct {
		ProductID   uuid.UUID `json:"product_id"`
		ProductName string    `json:"product_name"`
		Category    string    `json:"category"`
		UnitsSold   int64     `json:"units_sold"`
		Revenue     float64   `json:"revenue"`
	}

	database.DB.Table("order_items oi").
		Select("p.id as product_id, p.name as product_name, p.category, SUM(oi.quantity) as units_sold, SUM(oi.quantity * oi.price) as revenue").
		Joins("JOIN products p ON oi.product_id = p.id").
		Joins("JOIN orders o ON oi.order_id = o.id").
		Where("o.created_at >= ? AND o.status IN ?", cutoffDate, []string{"delivered", "completed"}).
		Group("p.id, p.name, p.category").
		Order("revenue DESC").
		Limit(limit).
		Scan(&topProducts)

	// Most viewed products
	var mostViewed []struct {
		ProductID   uuid.UUID `json:"product_id"`
		ProductName string    `json:"product_name"`
		ViewCount   int64     `json:"view_count"`
	}

	database.DB.Table("user_interactions ui").
		Select("p.id as product_id, p.name as product_name, COUNT(*) as view_count").
		Joins("JOIN products p ON ui.product_id = p.id").
		Where("ui.interaction_type = ? AND ui.created_at >= ?", "view", cutoffDate).
		Group("p.id, p.name").
		Order("view_count DESC").
		Limit(limit).
		Scan(&mostViewed)

	// Category performance
	var categoryPerformance []struct {
		Category  string  `json:"category"`
		UnitsSold int64   `json:"units_sold"`
		Revenue   float64 `json:"revenue"`
		ViewCount int64   `json:"view_count"`
	}

	database.DB.Table("products p").
		Select(`p.category, 
			COALESCE(SUM(oi.quantity), 0) as units_sold,
			COALESCE(SUM(oi.quantity * oi.price), 0) as revenue,
			COALESCE(COUNT(ui.id), 0) as view_count`).
		Joins("LEFT JOIN order_items oi ON p.id = oi.product_id").
		Joins("LEFT JOIN orders o ON oi.order_id = o.id AND o.created_at >= ? AND o.status IN ?", cutoffDate, []string{"delivered", "completed"}).
		Joins("LEFT JOIN user_interactions ui ON p.id = ui.product_id AND ui.interaction_type = ? AND ui.created_at >= ?", "view", cutoffDate).
		Group("p.category").
		Order("revenue DESC").
		Scan(&categoryPerformance)

	return c.JSON(fiber.Map{
		"top_products":         topProducts,
		"most_viewed":          mostViewed,
		"category_performance": categoryPerformance,
		"period_days":          days,
		"generated_at":         time.Now(),
	})
}

// GetMLTrends returns ML-powered trend analysis (placeholder)
// @Summary Get ML trends
// @Description Get ML-powered trend analysis and predictions (placeholder endpoint)
// @Tags Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "ML trends retrieved successfully"
// @Router /analytics/trends [get]
func GetMLTrends(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "ML trends analysis coming soon",
		"status":  "placeholder",
	})
}

// GetSearchAnalytics returns search analytics (placeholder)
// @Summary Get search analytics
// @Description Get search analytics and query performance metrics (placeholder endpoint)
// @Tags Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "Number of days to analyze" default(30)
// @Success 200 {object} map[string]interface{} "Search analytics retrieved successfully"
// @Router /analytics/search [get]
func GetSearchAnalytics(c *fiber.Ctx) error {
	days, _ := strconv.Atoi(c.Query("days", "30"))

	return c.JSON(fiber.Map{
		"message":     "Search analytics coming soon",
		"period_days": days,
		"status":      "placeholder",
	})
}

// GetRecommendationMetrics returns recommendation system metrics (placeholder)
// @Summary Get recommendation metrics
// @Description Get recommendation system performance metrics (placeholder endpoint)
// @Tags Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Recommendation metrics retrieved successfully"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Router /analytics/recommendations/metrics [get]
func GetRecommendationMetrics(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Recommendation metrics coming soon",
		"user_id": userID,
		"status":  "placeholder",
	})
}

// ExportAnalytics exports analytics data
// @Summary Export analytics data
// @Description Export user analytics data in JSON or CSV format
// @Tags Analytics
// @Accept json
// @Produce json,text/csv
// @Security BearerAuth
// @Param format query string false "Export format (json, csv)" default("json")
// @Param days query int false "Number of days to export" default(30)
// @Success 200 {object} map[string]interface{} "Analytics data exported successfully"
// @Failure 400 {object} map[string]interface{} "Unsupported format"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Router /analytics/export [get]
func ExportAnalytics(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	format := c.Query("format", "json")
	days, _ := strconv.Atoi(c.Query("days", "30"))
	cutoffDate := time.Now().AddDate(0, 0, -days)

	// Get user's data
	var orders []models.Order
	database.DB.Where("user_id = ? AND created_at >= ?", userID, cutoffDate).
		Preload("OrderItems.Product").
		Find(&orders)

	var interactions []models.UserInteraction
	database.DB.Where("user_id = ? AND created_at >= ?", userID, cutoffDate).
		Preload("Product").
		Find(&interactions)

	exportData := map[string]interface{}{
		"user_id":      userID,
		"period_days":  days,
		"export_date":  time.Now(),
		"orders":       orders,
		"interactions": interactions,
	}

	if format == "json" {
		return c.JSON(exportData)
	} else if format == "csv" {
		// For CSV, we'll return a simplified format
		c.Set("Content-Type", "text/csv")
		c.Set("Content-Disposition", "attachment; filename=analytics_export.csv")

		csvData := "Type,Date,Product,Category,Amount\n"

		for _, order := range orders {
			for _, item := range order.OrderItems {
				csvData += fmt.Sprintf("Order,%s,%s,%s,%.2f\n",
					order.CreatedAt.Format("2006-01-02"),
					item.Product.Name,
					item.Product.Category,
					item.Price*float64(item.Quantity))
			}
		}

		for _, interaction := range interactions {
			csvData += fmt.Sprintf("Interaction,%s,%s,%s,0\n",
				interaction.CreatedAt.Format("2006-01-02"),
				interaction.Product.Name,
				interaction.Product.Category)
		}

		return c.SendString(csvData)
	}

	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": "Unsupported format. Use 'json' or 'csv'",
	})
}
