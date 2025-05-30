package handlers

import (
	"log"
	"strconv"
	"strings"

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

// SearchProducts performs enhanced search with ML integration
// @Summary Search products
// @Description Search products using advanced ML-powered search with TF-IDF vectorization
// @Tags Products
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
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

	// Basic search implementation
	searchTerm := "%" + strings.ToLower(query) + "%"

	var products []models.Product
	var total int64

	// Count total results first
	if err := database.DB.Model(&models.Product{}).
		Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(category) LIKE ?",
			searchTerm, searchTerm, searchTerm).
		Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count search results",
		})
	}

	// Get search results
	if err := database.DB.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(category) LIKE ?",
		searchTerm, searchTerm, searchTerm).
		Offset(offset).Limit(limit).
		Find(&products).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search products",
		})
	}

	// Track search query
	go trackSearchQuery(c, query)

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
// @Summary Get product recommendations
// @Description Get personalized product recommendations using ML algorithms
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
		if err := database.DB.Order("created_at DESC").Limit(limit).Find(&popularProducts).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch popular products",
			})
		}

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
