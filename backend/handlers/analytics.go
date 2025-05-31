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

	// Enhanced dashboard data
	// Quick insights and trends
	var quickInsights struct {
		WeeklyGrowth        float64 `json:"weekly_growth_percentage"`
		TopCategory         string  `json:"top_category"`
		RecommendationScore float64 `json:"recommendation_score"`
		ActivityLevel       string  `json:"activity_level"`
		NextPredictedOrder  string  `json:"next_predicted_order"`
	}

	// Weekly growth calculation
	currentWeek := time.Now().AddDate(0, 0, -7)
	previousWeek := time.Now().AddDate(0, 0, -14)

	var currentWeekSpent, previousWeekSpent float64
	database.DB.Model(&models.Order{}).
		Where("user_id = ? AND created_at >= ? AND status IN ?", userID, currentWeek, []string{"delivered", "completed"}).
		Select("COALESCE(SUM(total), 0)").
		Scan(&currentWeekSpent)

	database.DB.Model(&models.Order{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ? AND status IN ?", userID, previousWeek, currentWeek, []string{"delivered", "completed"}).
		Select("COALESCE(SUM(total), 0)").
		Scan(&previousWeekSpent)

	if previousWeekSpent > 0 {
		quickInsights.WeeklyGrowth = ((currentWeekSpent - previousWeekSpent) / previousWeekSpent) * 100
	}

	// Top category
	database.DB.Table("user_interactions ui").
		Select("p.category").
		Joins("JOIN products p ON ui.product_id = p.id").
		Where("ui.user_id = ? AND ui.created_at >= ?", userID, thirtyDaysAgo).
		Group("p.category").
		Order("COUNT(*) DESC").
		Limit(1).
		Scan(&quickInsights.TopCategory)

	// Activity level based on recent interactions
	var recentActivityCount int64
	database.DB.Model(&models.UserInteraction{}).
		Where("user_id = ? AND created_at >= ?", userID, currentWeek).
		Count(&recentActivityCount)

	if recentActivityCount >= 20 {
		quickInsights.ActivityLevel = "High"
	} else if recentActivityCount >= 10 {
		quickInsights.ActivityLevel = "Medium"
	} else if recentActivityCount >= 1 {
		quickInsights.ActivityLevel = "Low"
	} else {
		quickInsights.ActivityLevel = "Inactive"
	}

	// Recommendation score (based on engagement with recommendations)
	var totalRecs, clickedRecs int64
	database.DB.Model(&models.Recommendation{}).Where("user_id = ?", userID).Count(&totalRecs)
	database.DB.Model(&models.RecommendationFeedback{}).Where("user_id = ? AND feedback_type = ?", userID, "clicked").Count(&clickedRecs)

	if totalRecs > 0 {
		quickInsights.RecommendationScore = float64(clickedRecs) / float64(totalRecs) * 100
	}

	// Alerts and notifications
	var alerts []DashboardAlert

	// Low stock alert for cart items
	var lowStockItems []struct {
		ProductName string `json:"product_name"`
		Stock       int    `json:"stock"`
	}

	database.DB.Table("cart_items ci").
		Select("p.name as product_name, p.stock").
		Joins("JOIN shopping_carts sc ON ci.cart_id = sc.id").
		Joins("JOIN products p ON ci.product_id = p.id").
		Where("sc.user_id = ? AND p.stock < 10", userID).
		Scan(&lowStockItems)

	for _, item := range lowStockItems {
		alerts = append(alerts, DashboardAlert{
			Type:    "warning",
			Title:   "Low Stock Alert",
			Message: fmt.Sprintf("Product '%s' in your cart has only %d items left", item.ProductName, item.Stock),
		})
	}

	// Price drop alerts (mock implementation)
	if len(recentInteractions) > 0 {
		alerts = append(alerts, DashboardAlert{
			Type:    "info",
			Title:   "Price Drop Alert",
			Message: "Some products you viewed recently have price drops!",
		})
	}

	// Personalized recommendations summary
	var recommendationSummary struct {
		TotalRecommendations int64                   `json:"total_recommendations"`
		TopRecommendations   []RecommendationSummary `json:"top_recommendations"`
		LastUpdated          time.Time               `json:"last_updated"`
	}

	database.DB.Model(&models.Recommendation{}).
		Where("user_id = ?", userID).
		Count(&recommendationSummary.TotalRecommendations)

	var topRecs []RecommendationSummary
	database.DB.Table("recommendations r").
		Select("p.id, p.name, p.category, p.price, r.score, r.algorithm_type").
		Joins("JOIN products p ON r.product_id = p.id").
		Where("r.user_id = ?", userID).
		Order("r.score DESC").
		Limit(3).
		Scan(&topRecs)

	recommendationSummary.TopRecommendations = topRecs
	recommendationSummary.LastUpdated = time.Now()

	return c.JSON(fiber.Map{
		"user_stats":             userStats,
		"recent_interactions":    recentInteractions,
		"quick_insights":         quickInsights,
		"alerts":                 alerts,
		"recommendation_summary": recommendationSummary,
		"timestamp":              time.Now(),
	})
}

// Supporting structs for enhanced dashboard
type DashboardAlert struct {
	Type    string `json:"type"` // "info", "warning", "error", "success"
	Title   string `json:"title"`
	Message string `json:"message"`
}

type RecommendationSummary struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Category      string    `json:"category"`
	Price         float64   `json:"price"`
	Score         float64   `json:"score"`
	AlgorithmType string    `json:"algorithm_type"`
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

	// Enhanced analytics - Behavioral patterns
	var behaviorAnalytics struct {
		MostActiveHour      int     `json:"most_active_hour"`
		MostActiveDayOfWeek string  `json:"most_active_day_of_week"`
		AvgSessionDuration  float64 `json:"avg_session_duration_minutes"`
		TotalSessions       int64   `json:"total_sessions"`
		ConversionRate      float64 `json:"conversion_rate"`
		CartAbandonmentRate float64 `json:"cart_abandonment_rate"`
	}

	// Most active hour
	database.DB.Model(&models.UserInteraction{}).
		Select("EXTRACT(HOUR FROM created_at) as hour").
		Where("user_id = ? AND created_at >= ?", userID, cutoffDate).
		Group("hour").
		Order("COUNT(*) DESC").
		Limit(1).
		Scan(&behaviorAnalytics.MostActiveHour)

	// Most active day of week
	var dayOfWeek int
	database.DB.Model(&models.UserInteraction{}).
		Select("EXTRACT(DOW FROM created_at) as dow").
		Where("user_id = ? AND created_at >= ?", userID, cutoffDate).
		Group("dow").
		Order("COUNT(*) DESC").
		Limit(1).
		Scan(&dayOfWeek)

	dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	if dayOfWeek >= 0 && dayOfWeek < len(dayNames) {
		behaviorAnalytics.MostActiveDayOfWeek = dayNames[dayOfWeek]
	}

	// Session analytics
	database.DB.Model(&models.UserSession{}).
		Where("user_id = ? AND started_at >= ?", userID, cutoffDate).
		Count(&behaviorAnalytics.TotalSessions)

	// Conversion rate (orders / sessions)
	var totalOrders int64
	database.DB.Model(&models.Order{}).
		Where("user_id = ? AND created_at >= ?", userID, cutoffDate).
		Count(&totalOrders)

	if behaviorAnalytics.TotalSessions > 0 {
		behaviorAnalytics.ConversionRate = float64(totalOrders) / float64(behaviorAnalytics.TotalSessions) * 100
	}

	// Cart abandonment rate
	var cartItems, completedOrders int64
	database.DB.Table("cart_items ci").
		Joins("JOIN shopping_carts sc ON ci.cart_id = sc.id").
		Where("sc.user_id = ? AND ci.created_at >= ?", userID, cutoffDate).
		Count(&cartItems)

	database.DB.Model(&models.Order{}).
		Where("user_id = ? AND created_at >= ? AND status IN ?", userID, cutoffDate, []string{"delivered", "completed"}).
		Count(&completedOrders)

	if cartItems > 0 {
		behaviorAnalytics.CartAbandonmentRate = float64(cartItems-completedOrders) / float64(cartItems) * 100
	}

	// Product preferences and recommendations insights
	var productInsights struct {
		FavoriteProducts    []ProductInsight      `json:"favorite_products"`
		RecommendationStats RecommendationInsight `json:"recommendation_stats"`
		SearchPatterns      []SearchPattern       `json:"search_patterns"`
	}

	// Favorite products (most interacted with)
	var favoriteProducts []ProductInsight
	database.DB.Table("user_interactions ui").
		Select("p.id, p.name, p.category, p.price, COUNT(*) as interaction_count").
		Joins("JOIN products p ON ui.product_id = p.id").
		Where("ui.user_id = ? AND ui.created_at >= ?", userID, cutoffDate).
		Group("p.id, p.name, p.category, p.price").
		Order("interaction_count DESC").
		Limit(5).
		Scan(&favoriteProducts)

	productInsights.FavoriteProducts = favoriteProducts

	// Recommendation stats
	var recStats RecommendationInsight
	database.DB.Model(&models.Recommendation{}).
		Where("user_id = ?", userID).
		Count(&recStats.TotalRecommendations)

	database.DB.Model(&models.RecommendationFeedback{}).
		Where("user_id = ? AND feedback_type = ?", userID, "clicked").
		Count(&recStats.ClickedRecommendations)

	if recStats.TotalRecommendations > 0 {
		recStats.ClickThroughRate = float64(recStats.ClickedRecommendations) / float64(recStats.TotalRecommendations) * 100
	}

	productInsights.RecommendationStats = recStats

	// Search patterns
	var searchPatterns []SearchPattern
	database.DB.Model(&models.SearchQuery{}).
		Select("query, COUNT(*) as frequency").
		Where("user_id = ? AND created_at >= ?", userID, cutoffDate).
		Group("query").
		Order("frequency DESC").
		Limit(5).
		Scan(&searchPatterns)

	productInsights.SearchPatterns = searchPatterns

	// Financial insights
	var financialInsights struct {
		TotalSpent            float64 `json:"total_spent"`
		AvgOrderValue         float64 `json:"avg_order_value"`
		LargestOrder          float64 `json:"largest_order"`
		MostExpensiveCategory string  `json:"most_expensive_category"`
		SavingsOpportunity    float64 `json:"potential_savings"`
	}

	database.DB.Model(&models.Order{}).
		Where("user_id = ? AND created_at >= ? AND status IN ?", userID, cutoffDate, []string{"delivered", "completed"}).
		Select("COALESCE(SUM(total), 0)").
		Scan(&financialInsights.TotalSpent)

	database.DB.Model(&models.Order{}).
		Where("user_id = ? AND created_at >= ? AND status IN ?", userID, cutoffDate, []string{"delivered", "completed"}).
		Select("COALESCE(AVG(total), 0)").
		Scan(&financialInsights.AvgOrderValue)

	database.DB.Model(&models.Order{}).
		Where("user_id = ? AND created_at >= ? AND status IN ?", userID, cutoffDate, []string{"delivered", "completed"}).
		Select("COALESCE(MAX(total), 0)").
		Scan(&financialInsights.LargestOrder)

	// Most expensive category
	database.DB.Table("order_items oi").
		Select("p.category").
		Joins("JOIN products p ON oi.product_id = p.id").
		Joins("JOIN orders o ON oi.order_id = o.id").
		Where("o.user_id = ? AND o.created_at >= ? AND o.status IN ?", userID, cutoffDate, []string{"delivered", "completed"}).
		Group("p.category").
		Order("SUM(oi.quantity * oi.price) DESC").
		Limit(1).
		Scan(&financialInsights.MostExpensiveCategory)

	return c.JSON(fiber.Map{
		"interaction_stats":  interactionStats,
		"category_stats":     categoryStats,
		"spending_over_time": spendingOverTime,
		"behavior_analytics": behaviorAnalytics,
		"product_insights":   productInsights,
		"financial_insights": financialInsights,
		"period_days":        days,
		"user_id":            userID,
		"generated_at":       time.Now(),
	})
}

// Supporting structs for enhanced analytics
type ProductInsight struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Category         string    `json:"category"`
	Price            float64   `json:"price"`
	InteractionCount int64     `json:"interaction_count"`
}

type RecommendationInsight struct {
	TotalRecommendations   int64   `json:"total_recommendations"`
	ClickedRecommendations int64   `json:"clicked_recommendations"`
	ClickThroughRate       float64 `json:"click_through_rate"`
}

type SearchPattern struct {
	Query     string `json:"query"`
	Frequency int64  `json:"frequency"`
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
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /analytics/products [get]
func GetProductAnalytics(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

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
		"user_analytics": fiber.Map{
			"user_id":                   userID,
			"personal_top_categories":   getUserTopCategories(userID, cutoffDate),
			"personal_purchase_history": getUserPurchaseStats(userID, cutoffDate),
		},
		"period_days":  days,
		"generated_at": time.Now(),
	})
}

// Helper function to get user's top categories
func getUserTopCategories(userID uuid.UUID, cutoffDate time.Time) []struct {
	Category string  `json:"category"`
	Count    int64   `json:"interaction_count"`
	Revenue  float64 `json:"total_spent"`
} {
	var userCategories []struct {
		Category string  `json:"category"`
		Count    int64   `json:"interaction_count"`
		Revenue  float64 `json:"total_spent"`
	}

	database.DB.Table("user_interactions ui").
		Select("p.category, COUNT(*) as count, COALESCE(SUM(p.price), 0) as revenue").
		Joins("JOIN products p ON ui.product_id = p.id").
		Where("ui.user_id = ? AND ui.created_at >= ?", userID, cutoffDate).
		Group("p.category").
		Order("count DESC").
		Limit(5).
		Scan(&userCategories)

	return userCategories
}

// Helper function to get user's purchase statistics
func getUserPurchaseStats(userID uuid.UUID, cutoffDate time.Time) struct {
	TotalOrders   int64   `json:"total_orders"`
	TotalSpent    float64 `json:"total_spent"`
	AvgOrderValue float64 `json:"avg_order_value"`
} {
	var stats struct {
		TotalOrders   int64   `json:"total_orders"`
		TotalSpent    float64 `json:"total_spent"`
		AvgOrderValue float64 `json:"avg_order_value"`
	}

	database.DB.Model(&models.Order{}).
		Where("user_id = ? AND created_at >= ? AND status IN ?", userID, cutoffDate, []string{"delivered", "completed"}).
		Count(&stats.TotalOrders)

	database.DB.Model(&models.Order{}).
		Where("user_id = ? AND created_at >= ? AND status IN ?", userID, cutoffDate, []string{"delivered", "completed"}).
		Select("COALESCE(SUM(total), 0)").
		Scan(&stats.TotalSpent)

	if stats.TotalOrders > 0 {
		stats.AvgOrderValue = stats.TotalSpent / float64(stats.TotalOrders)
	}

	return stats
}

// GetMLTrends returns ML-powered trend analysis
// @Summary Get ML trends
// @Description Get ML-powered trend analysis and predictions
// @Tags Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "Number of days to analyze" default(30)
// @Success 200 {object} map[string]interface{} "ML trends retrieved successfully"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /analytics/trends [get]
func GetMLTrends(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	days, _ := strconv.Atoi(c.Query("days", "30"))
	cutoffDate := time.Now().AddDate(0, 0, -days)

	// Trending products (based on recent interactions)
	var trendingProducts []struct {
		ProductID   uuid.UUID `json:"product_id"`
		ProductName string    `json:"product_name"`
		Category    string    `json:"category"`
		TrendScore  int64     `json:"trend_score"`
	}

	database.DB.Table("user_interactions ui").
		Select("p.id as product_id, p.name as product_name, p.category, COUNT(*) as trend_score").
		Joins("JOIN products p ON ui.product_id = p.id").
		Where("ui.created_at >= ?", cutoffDate).
		Group("p.id, p.name, p.category").
		Order("trend_score DESC").
		Limit(10).
		Scan(&trendingProducts)

	// Category trends
	var categoryTrends []struct {
		Category   string `json:"category"`
		TrendScore int64  `json:"trend_score"`
		Growth     string `json:"growth_trend"`
	}

	database.DB.Table("user_interactions ui").
		Select("p.category, COUNT(*) as trend_score").
		Joins("JOIN products p ON ui.product_id = p.id").
		Where("ui.created_at >= ?", cutoffDate).
		Group("p.category").
		Order("trend_score DESC").
		Scan(&categoryTrends)

	// Add growth trend analysis
	for i := range categoryTrends {
		// Simple growth calculation based on recent vs older data
		var recentCount, olderCount int64
		recentCutoff := time.Now().AddDate(0, 0, -7) // Last 7 days
		olderCutoff := time.Now().AddDate(0, 0, -14) // 7-14 days ago

		database.DB.Table("user_interactions ui").
			Joins("JOIN products p ON ui.product_id = p.id").
			Where("p.category = ? AND ui.created_at >= ?", categoryTrends[i].Category, recentCutoff).
			Count(&recentCount)

		database.DB.Table("user_interactions ui").
			Joins("JOIN products p ON ui.product_id = p.id").
			Where("p.category = ? AND ui.created_at >= ? AND ui.created_at < ?", categoryTrends[i].Category, olderCutoff, recentCutoff).
			Count(&olderCount)

		if olderCount == 0 {
			categoryTrends[i].Growth = "new"
		} else if recentCount > olderCount {
			categoryTrends[i].Growth = "growing"
		} else if recentCount < olderCount {
			categoryTrends[i].Growth = "declining"
		} else {
			categoryTrends[i].Growth = "stable"
		}
	}

	// User's personal trends
	var userTrends struct {
		MostActiveCategory string `json:"most_active_category"`
		RecentInteractions int64  `json:"recent_interactions"`
		TrendingInterest   string `json:"trending_interest"`
	}

	database.DB.Table("user_interactions ui").
		Select("p.category").
		Joins("JOIN products p ON ui.product_id = p.id").
		Where("ui.user_id = ? AND ui.created_at >= ?", userID, cutoffDate).
		Group("p.category").
		Order("COUNT(*) DESC").
		Limit(1).
		Scan(&userTrends.MostActiveCategory)

	database.DB.Model(&models.UserInteraction{}).
		Where("user_id = ? AND created_at >= ?", userID, cutoffDate).
		Count(&userTrends.RecentInteractions)

	// Determine trending interest based on recent activity
	var recentActivity, previousActivity int64
	recentWeek := time.Now().AddDate(0, 0, -7)
	previousWeek := time.Now().AddDate(0, 0, -14)

	database.DB.Model(&models.UserInteraction{}).
		Where("user_id = ? AND created_at >= ?", userID, recentWeek).
		Count(&recentActivity)

	database.DB.Model(&models.UserInteraction{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, previousWeek, recentWeek).
		Count(&previousActivity)

	if previousActivity == 0 {
		userTrends.TrendingInterest = "new_user"
	} else if recentActivity > previousActivity {
		userTrends.TrendingInterest = "increasing"
	} else if recentActivity < previousActivity {
		userTrends.TrendingInterest = "decreasing"
	} else {
		userTrends.TrendingInterest = "stable"
	}

	return c.JSON(fiber.Map{
		"trending_products": trendingProducts,
		"category_trends":   categoryTrends,
		"user_trends":       userTrends,
		"period_days":       days,
		"user_id":           userID,
		"generated_at":      time.Now(),
	})
}

// GetSearchAnalytics returns search analytics
// @Summary Get search analytics
// @Description Get search analytics and query performance metrics
// @Tags Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "Number of days to analyze" default(30)
// @Success 200 {object} map[string]interface{} "Search analytics retrieved successfully"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /analytics/search [get]
func GetSearchAnalytics(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	days, _ := strconv.Atoi(c.Query("days", "30"))
	cutoffDate := time.Now().AddDate(0, 0, -days)

	// Top search queries
	var topQueries []struct {
		Query       string `json:"query"`
		SearchCount int64  `json:"search_count"`
	}

	database.DB.Model(&models.SearchQuery{}).
		Select("query, COUNT(*) as search_count").
		Where("created_at >= ?", cutoffDate).
		Group("query").
		Order("search_count DESC").
		Limit(10).
		Scan(&topQueries)

	// Search volume over time
	var searchVolume []struct {
		Date        time.Time `json:"date"`
		SearchCount int64     `json:"search_count"`
	}

	database.DB.Model(&models.SearchQuery{}).
		Select("DATE(created_at) as date, COUNT(*) as search_count").
		Where("created_at >= ?", cutoffDate).
		Group("DATE(created_at)").
		Order("date").
		Scan(&searchVolume)

	// Zero result queries
	var zeroResultQueries []struct {
		Query       string `json:"query"`
		SearchCount int64  `json:"search_count"`
	}

	database.DB.Model(&models.SearchQuery{}).
		Select("query, COUNT(*) as search_count").
		Where("created_at >= ? AND results_count = 0", cutoffDate).
		Group("query").
		Order("search_count DESC").
		Limit(10).
		Scan(&zeroResultQueries)

	// User's personal search stats
	var userSearchStats struct {
		TotalSearches   int64 `json:"total_searches"`
		UniqueQueries   int64 `json:"unique_queries"`
		AvgResultsCount int64 `json:"avg_results_count"`
	}

	database.DB.Model(&models.SearchQuery{}).
		Where("user_id = ? AND created_at >= ?", userID, cutoffDate).
		Count(&userSearchStats.TotalSearches)

	database.DB.Model(&models.SearchQuery{}).
		Select("COUNT(DISTINCT query)").
		Where("user_id = ? AND created_at >= ?", userID, cutoffDate).
		Scan(&userSearchStats.UniqueQueries)

	database.DB.Model(&models.SearchQuery{}).
		Select("COALESCE(AVG(results_count), 0)").
		Where("user_id = ? AND created_at >= ?", userID, cutoffDate).
		Scan(&userSearchStats.AvgResultsCount)

	return c.JSON(fiber.Map{
		"top_queries":         topQueries,
		"search_volume":       searchVolume,
		"zero_result_queries": zeroResultQueries,
		"user_search_stats":   userSearchStats,
		"period_days":         days,
		"generated_at":        time.Now(),
	})
}

// GetRecommendationMetrics returns recommendation system metrics
// @Summary Get recommendation metrics
// @Description Get recommendation system performance metrics
// @Tags Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Recommendation metrics retrieved successfully"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /analytics/recommendations/metrics [get]
func GetRecommendationMetrics(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// User's recommendation stats
	var userRecommendationStats struct {
		TotalRecommendations     int64   `json:"total_recommendations"`
		ClickedRecommendations   int64   `json:"clicked_recommendations"`
		PurchasedRecommendations int64   `json:"purchased_recommendations"`
		ClickThroughRate         float64 `json:"click_through_rate"`
		ConversionRate           float64 `json:"conversion_rate"`
	}

	// Total recommendations for user
	database.DB.Model(&models.Recommendation{}).
		Where("user_id = ?", userID).
		Count(&userRecommendationStats.TotalRecommendations)

	// Clicked recommendations
	database.DB.Model(&models.RecommendationFeedback{}).
		Where("user_id = ? AND feedback_type = ?", userID, "clicked").
		Count(&userRecommendationStats.ClickedRecommendations)

	// Purchased recommendations
	database.DB.Model(&models.RecommendationFeedback{}).
		Where("user_id = ? AND feedback_type = ?", userID, "purchased").
		Count(&userRecommendationStats.PurchasedRecommendations)

	// Calculate rates
	if userRecommendationStats.TotalRecommendations > 0 {
		userRecommendationStats.ClickThroughRate = float64(userRecommendationStats.ClickedRecommendations) / float64(userRecommendationStats.TotalRecommendations) * 100
		userRecommendationStats.ConversionRate = float64(userRecommendationStats.PurchasedRecommendations) / float64(userRecommendationStats.TotalRecommendations) * 100
	}

	// Algorithm performance breakdown
	var algorithmStats []struct {
		Algorithm string  `json:"algorithm"`
		Count     int64   `json:"count"`
		AvgScore  float64 `json:"avg_score"`
	}

	database.DB.Model(&models.Recommendation{}).
		Select("algorithm_type as algorithm, COUNT(*) as count, AVG(score) as avg_score").
		Where("user_id = ?", userID).
		Group("algorithm_type").
		Scan(&algorithmStats)

	// Recent recommendations
	var recentRecommendations []models.Recommendation
	database.DB.Where("user_id = ?", userID).
		Preload("Product").
		Order("created_at DESC").
		Limit(5).
		Find(&recentRecommendations)

	// Global recommendation system stats (for context)
	var globalStats struct {
		TotalUsers                int64   `json:"total_users_with_recommendations"`
		TotalRecommendations      int64   `json:"total_recommendations"`
		AvgRecommendationsPerUser float64 `json:"avg_recommendations_per_user"`
	}

	database.DB.Model(&models.Recommendation{}).
		Select("COUNT(DISTINCT user_id)").
		Scan(&globalStats.TotalUsers)

	database.DB.Model(&models.Recommendation{}).
		Count(&globalStats.TotalRecommendations)

	if globalStats.TotalUsers > 0 {
		globalStats.AvgRecommendationsPerUser = float64(globalStats.TotalRecommendations) / float64(globalStats.TotalUsers)
	}

	return c.JSON(fiber.Map{
		"user_stats":             userRecommendationStats,
		"algorithm_breakdown":    algorithmStats,
		"recent_recommendations": recentRecommendations,
		"global_stats":           globalStats,
		"user_id":                userID,
		"generated_at":           time.Now(),
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
