package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"bachelor_backend/database"
	"bachelor_backend/middleware"
	"bachelor_backend/models"
	"bachelor_backend/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ProductResponse represents a product with additional metadata
type ProductResponse struct {
	models.Product
	IsRecommended bool `json:"is_recommended,omitempty"`
}

// CreateProductRequest represents the request to create a new product
type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255" example:"iPhone 15 Pro"`
	Description string  `json:"description" validate:"required,min=1,max=1000" example:"Latest iPhone with A17 Pro chip"`
	Price       float64 `json:"price" validate:"required,min=0.01" example:"999.99"`
	Category    string  `json:"category" validate:"required,min=1,max=100" example:"Electronics"`
	Stock       int     `json:"stock" validate:"required,min=0" example:"50"`
	ImageURL    string  `json:"image_url" validate:"omitempty,url" example:"https://example.com/image.jpg"`
}

// UpdateProductRequest represents the request to update a product
type UpdateProductRequest struct {
	Name        string  `json:"name" validate:"omitempty,min=1,max=255" example:"iPhone 15 Pro"`
	Description string  `json:"description" validate:"omitempty,min=1,max=1000" example:"Latest iPhone with A17 Pro chip"`
	Price       float64 `json:"price" validate:"omitempty,min=0.01" example:"999.99"`
	Category    string  `json:"category" validate:"omitempty,min=1,max=100" example:"Electronics"`
	Stock       int     `json:"stock" validate:"omitempty,min=0" example:"50"`
	ImageURL    string  `json:"image_url" validate:"omitempty,url" example:"https://example.com/image.jpg"`
}

// GetProducts returns a paginated list of products
// @Summary Get products
// @Description Get a paginated list of products with optional filtering and sorting
// @Tags Products
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param category query string false "Filter by category"
// @Param search query string false "Search in name and description"
// @Param sort query string false "Sort field (price, name, created_at)" default("created_at")
// @Param order query string false "Sort order (asc, desc)" default("desc")
// @Success 200 {object} map[string]interface{} "Products retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /products [get]
func GetProducts(c *fiber.Ctx) error {
	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	category := c.Query("category")
	search := c.Query("search")
	sortBy := c.Query("sort", "created_at")
	sortOrder := c.Query("order", "desc")

	// Calculate offset
	offset := (page - 1) * limit

	// Build query
	query := database.DB.Model(&models.Product{})

	// Apply filters
	if category != "" {
		query = query.Where("category = ?", category)
	}

	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)

		// Track search query for analytics
		go trackSearchQuery(c, search)
	}

	// Apply sorting
	if sortBy == "price" || sortBy == "name" || sortBy == "created_at" {
		if sortOrder == "asc" || sortOrder == "desc" {
			query = query.Order(sortBy + " " + sortOrder)
		}
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count products",
		})
	}

	// Get products
	var products []models.Product
	if err := query.Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch products",
		})
	}

	// Track product views for each product (for analytics)
	go trackProductViews(c, products)

	return c.JSON(fiber.Map{
		"products": products,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetProduct returns a single product by ID
// @Summary Get product by ID
// @Description Get detailed information about a specific product
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Success 200 {object} models.Product "Product retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid product ID"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Router /products/{id} [get]
func GetProduct(c *fiber.Ctx) error {
	productID := c.Params("id")

	// Parse UUID
	id, err := uuid.Parse(productID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	var product models.Product
	if err := database.DB.First(&product, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Product not found",
		})
	}

	// Track product view
	go trackSingleProductView(c, product.ID)

	// Track user interaction if user is authenticated
	if userID, ok := middleware.GetUserID(c); ok {
		go trackUserInteraction(userID, product.ID, "view", c.Get("X-Session-ID"))
	}

	return c.JSON(product)
}

// GetProductsByCategory returns products filtered by category
func GetProductsByCategory(c *fiber.Ctx) error {
	category := c.Params("category")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	offset := (page - 1) * limit

	var products []models.Product
	var total int64

	query := database.DB.Where("category = ?", category)

	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count products",
		})
	}

	if err := query.Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch products",
		})
	}

	return c.JSON(fiber.Map{
		"products": products,
		"category": category,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetCategories returns all available product categories
func GetCategories(c *fiber.Ctx) error {
	var categories []string

	if err := database.DB.Model(&models.Product{}).
		Distinct("category").
		Pluck("category", &categories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch categories",
		})
	}

	return c.JSON(fiber.Map{
		"categories": categories,
	})
}

// SearchProducts performs enhanced search with relevance scoring
// @Summary Search products
// @Description Search products using enhanced search with relevance scoring and intelligent filtering
// @Tags Products
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param category query string false "Filter by category"
// @Param min_price query number false "Minimum price filter"
// @Param max_price query number false "Maximum price filter"
// @Success 200 {object} map[string]interface{} "Search results retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Search query is required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /products/search [get]
func SearchProducts(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Search query is required",
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset := (page - 1) * limit

	// Additional filters
	category := c.Query("category")
	minPrice, _ := strconv.ParseFloat(c.Query("min_price", "0"), 64)
	maxPrice, _ := strconv.ParseFloat(c.Query("max_price", "999999"), 64)

	// Enhanced search with relevance scoring
	searchResults, total, err := performEnhancedSearch(query, category, minPrice, maxPrice, offset, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search products",
		})
	}

	// Track search query with results count
	go trackSearchQueryWithResults(c, query, int(total))

	return c.JSON(fiber.Map{
		"products": searchResults,
		"query":    query,
		"filters": fiber.Map{
			"category":  category,
			"min_price": minPrice,
			"max_price": maxPrice,
		},
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// Enhanced search function with relevance scoring
func performEnhancedSearch(query, category string, minPrice, maxPrice float64, offset, limit int) ([]ProductSearchResult, int64, error) {
	// Normalize and prepare search terms
	normalizedQuery := strings.ToLower(strings.TrimSpace(query))
	searchTerms := strings.Fields(normalizedQuery)

	if len(searchTerms) == 0 {
		return nil, 0, fmt.Errorf("invalid search query")
	}

	// Build the search query with relevance scoring
	var searchResults []ProductSearchResult
	var total int64

	// Create a complex query that scores relevance
	baseQuery := database.DB.Model(&models.Product{}).
		Select(`
			products.*,
			(
				CASE 
					WHEN LOWER(name) = ? THEN 100
					WHEN LOWER(name) LIKE ? THEN 80
					WHEN LOWER(name) LIKE ? THEN 60
					ELSE 0
				END +
				CASE 
					WHEN LOWER(description) LIKE ? THEN 40
					WHEN LOWER(description) LIKE ? THEN 20
					ELSE 0
				END +
				CASE 
					WHEN LOWER(category) = ? THEN 50
					WHEN LOWER(category) LIKE ? THEN 30
					ELSE 0
				END
			) as relevance_score
		`,
			normalizedQuery,                        // Exact name match
			normalizedQuery+"%",                    // Name starts with query
			"%"+normalizedQuery+"%",                // Name contains query
			"%"+normalizedQuery+"%",                // Description contains query
			"%"+strings.Join(searchTerms, "%")+"%", // Description contains all terms
			normalizedQuery,                        // Exact category match
			"%"+normalizedQuery+"%",                // Category contains query
		)

	// Apply filters
	whereConditions := []string{}
	whereArgs := []interface{}{}

	// Search condition - must match at least one field
	searchCondition := `(
		LOWER(name) LIKE ? OR 
		LOWER(description) LIKE ? OR 
		LOWER(category) LIKE ?
	)`
	whereConditions = append(whereConditions, searchCondition)
	whereArgs = append(whereArgs, "%"+normalizedQuery+"%", "%"+normalizedQuery+"%", "%"+normalizedQuery+"%")

	// Additional term matching for better relevance
	for _, term := range searchTerms {
		if len(term) > 2 { // Only consider terms longer than 2 characters
			termCondition := `(
				LOWER(name) LIKE ? OR 
				LOWER(description) LIKE ? OR 
				LOWER(category) LIKE ?
			)`
			whereConditions = append(whereConditions, termCondition)
			whereArgs = append(whereArgs, "%"+term+"%", "%"+term+"%", "%"+term+"%")
		}
	}

	// Price filters
	if minPrice > 0 {
		whereConditions = append(whereConditions, "price >= ?")
		whereArgs = append(whereArgs, minPrice)
	}
	if maxPrice < 999999 {
		whereConditions = append(whereConditions, "price <= ?")
		whereArgs = append(whereArgs, maxPrice)
	}

	// Category filter
	if category != "" {
		whereConditions = append(whereConditions, "LOWER(category) = ?")
		whereArgs = append(whereArgs, strings.ToLower(category))
	}

	// Stock filter - only show available products
	whereConditions = append(whereConditions, "stock > 0")

	// Combine all conditions
	finalQuery := baseQuery.Where(strings.Join(whereConditions, " AND "), whereArgs...)

	// Create a separate query for counting (without the SELECT with calculated fields)
	countQuery := database.DB.Model(&models.Product{}).Where(strings.Join(whereConditions, " AND "), whereArgs...)

	// Count total results
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get results with relevance scoring and ordering
	if err := finalQuery.
		Order("relevance_score DESC, name ASC").
		Offset(offset).
		Limit(limit).
		Scan(&searchResults).Error; err != nil {
		return nil, 0, err
	}

	// Filter out results with zero relevance score in application logic
	filteredResults := make([]ProductSearchResult, 0, len(searchResults))
	for _, result := range searchResults {
		if result.RelevanceScore > 0 {
			filteredResults = append(filteredResults, result)
		}
	}

	return filteredResults, total, nil
}

// ProductSearchResult represents a search result with relevance score
type ProductSearchResult struct {
	models.Product
	RelevanceScore int `json:"relevance_score"`
}

// Enhanced search query tracking
func trackSearchQueryWithResults(c *fiber.Ctx, query string, resultsCount int) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Error in trackSearchQueryWithResults: %v", r)
			}
		}()

		var userID *uuid.UUID
		if id, ok := middleware.GetUserID(c); ok {
			userID = &id
		}

		searchQuery := models.SearchQuery{
			UserID:       userID,
			Query:        query,
			ResultsCount: resultsCount,
		}

		if err := database.DB.Create(&searchQuery).Error; err != nil {
			log.Printf("Failed to track search query: %v", err)
		}
	}()
}

// GetRecommendations returns ML-generated product recommendations
// @Summary Get product recommendations
// @Description Get personalized product recommendations using ML algorithms with reasoning
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of recommendations" default(10)
// @Success 200 {object} map[string]interface{} "Recommendations retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Authentication required for recommendations"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /products/recommendations [get]
func GetRecommendations(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required for recommendations",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	// Try to get fresh recommendations from ML service
	go generateMLRecommendations(userID, limit)

	// Get recommendations from database with reasoning
	var recommendations []models.Recommendation
	if err := database.DB.Where("user_id = ?", userID).
		Preload("Product").
		Order("score DESC").
		Limit(limit).
		Find(&recommendations).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch recommendations",
		})
	}

	// Extract products from recommendations with reasoning
	var products []ProductRecommendationResponse
	for _, rec := range recommendations {
		reasoning := generateRecommendationReasoning(userID, rec)

		products = append(products, ProductRecommendationResponse{
			Product:       rec.Product,
			Score:         rec.Score,
			Algorithm:     rec.AlgorithmType,
			Reasoning:     reasoning,
			IsRecommended: true,
		})
	}

	// If no recommendations found, return popular products with reasoning
	if len(products) == 0 {
		var popularProducts []models.Product
		if err := database.DB.Order("created_at DESC").Limit(limit).Find(&popularProducts).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch popular products",
			})
		}

		for _, product := range popularProducts {
			reasoning := RecommendationReasoning{
				Primary:   "Popular item",
				Secondary: "This product is trending among other users",
				Factors:   []string{"Recently added", "High general interest"},
			}

			products = append(products, ProductRecommendationResponse{
				Product:       product,
				Score:         0.5, // Default score for popular items
				Algorithm:     "popular",
				Reasoning:     reasoning,
				IsRecommended: false,
			})
		}
	}

	// Get user insights for context
	userInsights := getUserRecommendationInsights(userID)

	return c.JSON(fiber.Map{
		"recommendations": products,
		"user_id":         userID,
		"insights":        userInsights,
		"total_count":     len(products),
	})
}

// ProductRecommendationResponse represents a recommendation with reasoning
type ProductRecommendationResponse struct {
	models.Product
	Score         float64                 `json:"score"`
	Algorithm     string                  `json:"algorithm"`
	Reasoning     RecommendationReasoning `json:"reasoning"`
	IsRecommended bool                    `json:"is_recommended"`
}

// RecommendationReasoning explains why an item is recommended
type RecommendationReasoning struct {
	Primary   string   `json:"primary_reason"`
	Secondary string   `json:"secondary_reason"`
	Factors   []string `json:"contributing_factors"`
}

// UserRecommendationInsights provides context about user preferences
type UserRecommendationInsights struct {
	TopCategories       []string `json:"top_categories"`
	RecentActivity      string   `json:"recent_activity"`
	PreferenceStyle     string   `json:"preference_style"`
	RecommendationCount int64    `json:"total_recommendations"`
}

// Generate reasoning for why an item is recommended
func generateRecommendationReasoning(userID uuid.UUID, rec models.Recommendation) RecommendationReasoning {
	var reasoning RecommendationReasoning
	var factors []string

	// Base reasoning on algorithm type
	switch rec.AlgorithmType {
	case "collaborative":
		reasoning.Primary = "Users like you also liked this"
		reasoning.Secondary = "Based on similar user preferences and purchase patterns"
		factors = append(factors, "Similar user behavior", "Collaborative filtering")

	case "content_based":
		reasoning.Primary = "Matches your interests"
		reasoning.Secondary = "Similar to products you've viewed or purchased"
		factors = append(factors, "Content similarity", "Your browsing history")

	case "hybrid":
		reasoning.Primary = "Personalized for you"
		reasoning.Secondary = "Combines your preferences with similar user patterns"
		factors = append(factors, "Hybrid algorithm", "Personal + collaborative data")

	default:
		reasoning.Primary = "Recommended for you"
		reasoning.Secondary = "Based on our recommendation system"
		factors = append(factors, "General recommendation")
	}

	// Add score-based factors
	if rec.Score > 0.8 {
		factors = append(factors, "High confidence match")
	} else if rec.Score > 0.6 {
		factors = append(factors, "Good match")
	} else {
		factors = append(factors, "Potential interest")
	}

	// Get user's interaction with this category
	var categoryInteractions int64
	database.DB.Table("user_interactions ui").
		Joins("JOIN products p ON ui.product_id = p.id").
		Where("ui.user_id = ? AND p.category = ?", userID, rec.Product.Category).
		Count(&categoryInteractions)

	if categoryInteractions > 0 {
		factors = append(factors, fmt.Sprintf("You've shown interest in %s", rec.Product.Category))
	}

	// Check if user has similar products
	var similarProducts int64
	database.DB.Table("user_interactions ui").
		Joins("JOIN products p ON ui.product_id = p.id").
		Where("ui.user_id = ? AND p.category = ? AND ui.interaction_type IN ?",
			userID, rec.Product.Category, []string{"view", "cart_add", "purchase"}).
		Count(&similarProducts)

	if similarProducts > 0 {
		factors = append(factors, "Similar to your previous choices")
	}

	// Check price range preference
	var avgSpent float64
	database.DB.Table("order_items oi").
		Joins("JOIN orders o ON oi.order_id = o.id").
		Joins("JOIN products p ON oi.product_id = p.id").
		Where("o.user_id = ? AND p.category = ?", userID, rec.Product.Category).
		Select("AVG(oi.price)").
		Scan(&avgSpent)

	if avgSpent > 0 {
		priceDiff := (rec.Product.Price - avgSpent) / avgSpent
		if priceDiff < 0.2 && priceDiff > -0.2 {
			factors = append(factors, "Within your usual price range")
		}
	}

	reasoning.Factors = factors
	return reasoning
}

// Get user insights for recommendation context
func getUserRecommendationInsights(userID uuid.UUID) UserRecommendationInsights {
	var insights UserRecommendationInsights

	// Get top categories
	var topCategories []struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}

	database.DB.Table("user_interactions ui").
		Select("p.category, COUNT(*) as count").
		Joins("JOIN products p ON ui.product_id = p.id").
		Where("ui.user_id = ?", userID).
		Group("p.category").
		Order("count DESC").
		Limit(3).
		Scan(&topCategories)

	for _, cat := range topCategories {
		insights.TopCategories = append(insights.TopCategories, cat.Category)
	}

	// Recent activity
	var recentInteractions int64
	weekAgo := time.Now().AddDate(0, 0, -7)
	database.DB.Model(&models.UserInteraction{}).
		Where("user_id = ? AND created_at >= ?", userID, weekAgo).
		Count(&recentInteractions)

	if recentInteractions > 20 {
		insights.RecentActivity = "Very active"
	} else if recentInteractions > 10 {
		insights.RecentActivity = "Active"
	} else if recentInteractions > 0 {
		insights.RecentActivity = "Moderate"
	} else {
		insights.RecentActivity = "Low activity"
	}

	// Preference style based on interaction patterns
	var viewToPurchaseRatio float64
	var totalViews, totalPurchases int64

	database.DB.Model(&models.UserInteraction{}).
		Where("user_id = ? AND interaction_type = ?", userID, "view").
		Count(&totalViews)

	database.DB.Model(&models.UserInteraction{}).
		Where("user_id = ? AND interaction_type = ?", userID, "purchase").
		Count(&totalPurchases)

	if totalViews > 0 {
		viewToPurchaseRatio = float64(totalPurchases) / float64(totalViews)
	}

	if viewToPurchaseRatio > 0.1 {
		insights.PreferenceStyle = "Decisive buyer"
	} else if viewToPurchaseRatio > 0.05 {
		insights.PreferenceStyle = "Careful shopper"
	} else {
		insights.PreferenceStyle = "Browser"
	}

	// Total recommendations count
	database.DB.Model(&models.Recommendation{}).
		Where("user_id = ?", userID).
		Count(&insights.RecommendationCount)

	return insights
}

// Helper functions for tracking

func trackSearchQuery(c *fiber.Ctx, query string) {
	// Use a separate goroutine with proper error handling
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Error in trackSearchQuery: %v", r)
			}
		}()

		var userID *uuid.UUID
		if id, ok := middleware.GetUserID(c); ok {
			userID = &id
		}

		searchQuery := models.SearchQuery{
			UserID: userID,
			Query:  query,
		}

		if err := database.DB.Create(&searchQuery).Error; err != nil {
			log.Printf("Failed to track search query: %v", err)
		}
	}()
}

func trackProductViews(c *fiber.Ctx, products []models.Product) {
	// Use a separate goroutine with proper error handling
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Error in trackProductViews: %v", r)
			}
		}()

		var userID *uuid.UUID
		if id, ok := middleware.GetUserID(c); ok {
			userID = &id
		}

		sessionID := c.Get("X-Session-ID")

		// Batch insert for better performance
		views := make([]models.ProductView, 0, len(products))
		for _, product := range products {
			views = append(views, models.ProductView{
				UserID:    userID,
				ProductID: product.ID,
				SessionID: sessionID,
			})
		}

		if len(views) > 0 {
			if err := database.DB.CreateInBatches(views, 100).Error; err != nil {
				log.Printf("Failed to track product views: %v", err)
			}
		}
	}()
}

// GetMLStatus returns the status of ML models
func GetMLStatus(c *fiber.Ctx) error {
	status, err := services.MLService.GetMLStatus()
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":   "ML service unavailable",
			"details": err.Error(),
		})
	}

	return c.JSON(status)
}

// TrainMLModels triggers training of ML models
func TrainMLModels(c *fiber.Ctx) error {
	// Check if user is authenticated (you might want to add admin check here)
	_, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	err := services.MLService.TrainModels()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to train ML models",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "ML model training initiated successfully",
		"status":  "success",
	})
}

func trackSingleProductView(c *fiber.Ctx, productID uuid.UUID) {
	// Use a separate goroutine with proper error handling
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Error in trackSingleProductView: %v", r)
			}
		}()

		var userID *uuid.UUID
		if id, ok := middleware.GetUserID(c); ok {
			userID = &id
		}

		view := models.ProductView{
			UserID:    userID,
			ProductID: productID,
			SessionID: c.Get("X-Session-ID"),
		}

		if err := database.DB.Create(&view).Error; err != nil {
			log.Printf("Failed to track single product view: %v", err)
		}
	}()
}

func trackUserInteraction(userID, productID uuid.UUID, interactionType, sessionID string) {
	// Use a separate goroutine with proper error handling
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Error in trackUserInteraction: %v", r)
			}
		}()

		interaction := models.UserInteraction{
			UserID:          userID,
			ProductID:       productID,
			InteractionType: interactionType,
			SessionID:       sessionID,
		}

		if err := database.DB.Create(&interaction).Error; err != nil {
			log.Printf("Failed to track user interaction: %v", err)
		}
	}()
}

// generateMLRecommendations calls the ML service to generate recommendations
func generateMLRecommendations(userID uuid.UUID, limit int) {
	// Use a separate goroutine with proper error handling
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Error in generateMLRecommendations: %v", r)
			}
		}()

		// Try to call ML service first
		mlRecommendations, err := services.MLService.GenerateRecommendations(userID, "hybrid", limit)
		if err == nil && mlRecommendations != nil {
			// Save ML recommendations to database with batch insert
			recommendations := make([]models.Recommendation, 0, len(mlRecommendations.Recommendations))

			for _, mlRec := range mlRecommendations.Recommendations {
				productUUID, err := uuid.Parse(mlRec.ProductID)
				if err != nil {
					log.Printf("Invalid product ID in ML recommendation: %s", mlRec.ProductID)
					continue
				}

				recommendations = append(recommendations, models.Recommendation{
					UserID:        userID,
					ProductID:     productUUID,
					AlgorithmType: mlRec.Algorithm,
					Score:         mlRec.Score,
				})
			}

			if len(recommendations) > 0 {
				if err := database.DB.CreateInBatches(recommendations, 100).Error; err != nil {
					log.Printf("Failed to save ML recommendations: %v", err)
				}
			}
			return
		}

		// Fallback: Check if user already has recent recommendations
		var count int64
		if err := database.DB.Model(&models.Recommendation{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
			log.Printf("Failed to count existing recommendations: %v", err)
			return
		}

		if count == 0 {
			// Create some sample recommendations for demo purposes
			var products []models.Product
			if err := database.DB.Limit(limit).Find(&products).Error; err != nil {
				log.Printf("Failed to fetch products for fallback recommendations: %v", err)
				return
			}

			recommendations := make([]models.Recommendation, 0, len(products))
			for i, product := range products {
				recommendations = append(recommendations, models.Recommendation{
					UserID:        userID,
					ProductID:     product.ID,
					AlgorithmType: "hybrid",
					Score:         float64(limit-i) / float64(limit), // Decreasing scores
				})
			}

			if len(recommendations) > 0 {
				if err := database.DB.CreateInBatches(recommendations, 100).Error; err != nil {
					log.Printf("Failed to create fallback recommendations: %v", err)
				}
			}
		}
	}()
}

// CreateProduct creates a new product (admin only)
// @Summary Create a new product
// @Description Create a new product in the catalog (admin access required)
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateProductRequest true "Product creation data"
// @Success 201 {object} map[string]interface{} "Product created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or validation error"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /products [post]
func CreateProduct(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Authentication required",
		})
	}

	var req CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Validate request
	if err := middleware.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Create product
	product := models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		Stock:       req.Stock,
		ImageURL:    req.ImageURL,
	}

	if err := database.DB.Create(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create product",
		})
	}

	// Track admin action
	go trackUserInteraction(userID, product.ID, "admin_create", c.Get("X-Session-ID"))

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Product created successfully",
		"product": product,
	})
}

// UpdateProduct updates an existing product (admin only)
// @Summary Update a product
// @Description Update an existing product in the catalog (admin access required)
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID (UUID)"
// @Param request body UpdateProductRequest true "Product update data"
// @Success 200 {object} map[string]interface{} "Product updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or product ID"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /products/{id} [put]
func UpdateProduct(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Authentication required",
		})
	}

	productID := c.Params("id")
	id, err := uuid.Parse(productID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid product ID",
		})
	}

	var req UpdateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Validate request
	if err := middleware.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Get existing product
	var product models.Product
	if err := database.DB.First(&product, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Product not found",
		})
	}

	// Update fields if provided
	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Price > 0 {
		product.Price = req.Price
	}
	if req.Category != "" {
		product.Category = req.Category
	}
	if req.Stock >= 0 {
		product.Stock = req.Stock
	}
	if req.ImageURL != "" {
		product.ImageURL = req.ImageURL
	}

	if err := database.DB.Save(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update product",
		})
	}

	// Track admin action
	go trackUserInteraction(userID, product.ID, "admin_update", c.Get("X-Session-ID"))

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Product updated successfully",
		"product": product,
	})
}

// DeleteProduct deletes a product (admin only)
// @Summary Delete a product
// @Description Delete a product from the catalog (admin access required)
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID (UUID)"
// @Success 200 {object} map[string]interface{} "Product deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid product ID or product has order history"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /products/{id} [delete]
func DeleteProduct(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Authentication required",
		})
	}

	productID := c.Params("id")
	id, err := uuid.Parse(productID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid product ID",
		})
	}

	// Get existing product
	var product models.Product
	if err := database.DB.First(&product, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Product not found",
		})
	}

	// Check if product has any orders (prevent deletion if it has order history)
	var orderCount int64
	if err := database.DB.Model(&models.OrderItem{}).Where("product_id = ?", id).Count(&orderCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to check product dependencies",
		})
	}

	if orderCount > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Cannot delete product with existing order history",
		})
	}

	// Delete the product
	if err := database.DB.Delete(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to delete product",
		})
	}

	// Track admin action
	go trackUserInteraction(userID, product.ID, "admin_delete", c.Get("X-Session-ID"))

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Product deleted successfully",
	})
}
