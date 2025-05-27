package handlers

import (
	"strconv"
	"strings"

	"bachelor_backend/database"
	"bachelor_backend/middleware"
	"bachelor_backend/models"
	"bachelor_backend/services"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// ProductResponse represents a product with additional metadata
type ProductResponse struct {
	models.Product
	IsRecommended bool `json:"is_recommended,omitempty"`
}

// GetProducts returns a paginated list of products
func GetProducts(c fiber.Ctx) error {
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
	query.Count(&total)

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
func GetProduct(c fiber.Ctx) error {
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
func GetProductsByCategory(c fiber.Ctx) error {
	category := c.Params("category")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	offset := (page - 1) * limit

	var products []models.Product
	var total int64

	query := database.DB.Where("category = ?", category)
	query.Count(&total)

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
func GetCategories(c fiber.Ctx) error {
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

// SearchProducts performs enhanced search with ML integration
func SearchProducts(c fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Search query is required",
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset := (page - 1) * limit

	// Basic search implementation
	searchTerm := "%" + strings.ToLower(query) + "%"

	var products []models.Product
	var total int64

	dbQuery := database.DB.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(category) LIKE ?",
		searchTerm, searchTerm, searchTerm)

	dbQuery.Count(&total)

	if err := dbQuery.Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search products",
		})
	}

	// Track search query
	go trackSearchQuery(c, query)

	// TODO: Integrate with ML service for enhanced search ranking
	// This will be implemented when we create the ML service

	return c.JSON(fiber.Map{
		"products": products,
		"query":    query,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetRecommendations returns ML-generated product recommendations
func GetRecommendations(c fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required for recommendations",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	// Try to get fresh recommendations from ML service
	go generateMLRecommendations(userID, limit)

	// Get recommendations from database
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

	// Extract products from recommendations
	var products []ProductResponse
	for _, rec := range recommendations {
		products = append(products, ProductResponse{
			Product:       rec.Product,
			IsRecommended: true,
		})
	}

	// If no recommendations found, return popular products
	if len(products) == 0 {
		var popularProducts []models.Product
		database.DB.Order("created_at DESC").Limit(limit).Find(&popularProducts)

		for _, product := range popularProducts {
			products = append(products, ProductResponse{
				Product:       product,
				IsRecommended: false,
			})
		}
	}

	return c.JSON(fiber.Map{
		"recommendations": products,
		"user_id":         userID,
	})
}

// Helper functions for tracking

func trackSearchQuery(c fiber.Ctx, query string) {
	var userID *uuid.UUID
	if id, ok := middleware.GetUserID(c); ok {
		userID = &id
	}

	searchQuery := models.SearchQuery{
		UserID: userID,
		Query:  query,
	}

	database.DB.Create(&searchQuery)
}

func trackProductViews(c fiber.Ctx, products []models.Product) {
	var userID *uuid.UUID
	if id, ok := middleware.GetUserID(c); ok {
		userID = &id
	}

	sessionID := c.Get("X-Session-ID")

	for _, product := range products {
		view := models.ProductView{
			UserID:    userID,
			ProductID: product.ID,
			SessionID: sessionID,
		}
		database.DB.Create(&view)
	}
}

// GetMLStatus returns the status of ML models
func GetMLStatus(c fiber.Ctx) error {
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
func TrainMLModels(c fiber.Ctx) error {
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

func trackSingleProductView(c fiber.Ctx, productID uuid.UUID) {
	var userID *uuid.UUID
	if id, ok := middleware.GetUserID(c); ok {
		userID = &id
	}

	view := models.ProductView{
		UserID:    userID,
		ProductID: productID,
		SessionID: c.Get("X-Session-ID"),
	}

	database.DB.Create(&view)
}

func trackUserInteraction(userID, productID uuid.UUID, interactionType, sessionID string) {
	interaction := models.UserInteraction{
		UserID:          userID,
		ProductID:       productID,
		InteractionType: interactionType,
		SessionID:       sessionID,
	}

	database.DB.Create(&interaction)
}

// generateMLRecommendations calls the ML service to generate recommendations
func generateMLRecommendations(userID uuid.UUID, limit int) {
	// Try to call ML service first
	mlRecommendations, err := services.MLService.GenerateRecommendations(userID, "hybrid", limit)
	if err == nil && mlRecommendations != nil {
		// Save ML recommendations to database
		for _, mlRec := range mlRecommendations.Recommendations {
			productUUID, err := uuid.Parse(mlRec.ProductID)
			if err != nil {
				continue
			}

			rec := models.Recommendation{
				UserID:        userID,
				ProductID:     productUUID,
				AlgorithmType: mlRec.Algorithm,
				Score:         mlRec.Score,
			}
			database.DB.Create(&rec)
		}
		return
	}

	// Fallback: Check if user already has recent recommendations
	var count int64
	database.DB.Model(&models.Recommendation{}).Where("user_id = ?", userID).Count(&count)

	if count == 0 {
		// Create some sample recommendations for demo purposes
		var products []models.Product
		database.DB.Limit(limit).Find(&products)

		for i, product := range products {
			rec := models.Recommendation{
				UserID:        userID,
				ProductID:     product.ID,
				AlgorithmType: "hybrid",
				Score:         float64(limit-i) / float64(limit), // Decreasing scores
			}
			database.DB.Create(&rec)
		}
	}
}
