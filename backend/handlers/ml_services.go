package handlers

import (
	"bachelor_backend/middleware"
	"bachelor_backend/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GetProductSentiment analyzes sentiment for a specific product
// @Summary Analyze product sentiment
// @Description Get sentiment analysis for a specific product based on user comments
// @Tags ML
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID (UUID)"
// @Success 200 {object} map[string]interface{} "Product sentiment analysis retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid product ID"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /ml/sentiment/product/{id} [get]
func GetProductSentiment(c *fiber.Ctx) error {
	_, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not authenticated",
		})
	}

	productIDStr := c.Params("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid product ID",
		})
	}

	sentiment, err := services.MLService.AnalyzeProductSentiment(productID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to analyze product sentiment: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    sentiment,
	})
}

// GetCategorySentiment analyzes sentiment for a product category
// @Summary Analyze category sentiment
// @Description Get sentiment analysis for a product category based on user comments
// @Tags ML
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category path string true "Product Category"
// @Success 200 {object} map[string]interface{} "Category sentiment analysis retrieved successfully"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /ml/sentiment/category/{category} [get]
func GetCategorySentiment(c *fiber.Ctx) error {
	_, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not authenticated",
		})
	}

	category := c.Params("category")
	if category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Category is required",
		})
	}

	sentiment, err := services.MLService.AnalyzeCategorySentiment(category)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to analyze category sentiment: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    sentiment,
	})
}

// GetSentimentInsights gets overall sentiment insights
// @Summary Get sentiment insights
// @Description Get actionable sentiment insights across all products and categories
// @Tags ML
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Sentiment insights retrieved successfully"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /ml/sentiment/insights [get]
func GetSentimentInsights(c *fiber.Ctx) error {
	_, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not authenticated",
		})
	}

	insights, err := services.MLService.GetSentimentInsights()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get sentiment insights: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    insights,
	})
}

// SuggestProductTags suggests tags for a specific product
// @Summary Suggest product tags
// @Description Get AI-suggested tags for a specific product based on content and user behavior
// @Tags ML
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID (UUID)"
// @Success 200 {object} map[string]interface{} "Product tag suggestions retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid product ID"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /ml/auto-tagging/suggest/{id} [get]
func SuggestProductTags(c *fiber.Ctx) error {
	_, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not authenticated",
		})
	}

	productIDStr := c.Params("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid product ID",
		})
	}

	tags, err := services.MLService.SuggestProductTags(productID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to suggest product tags: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    tags,
	})
}

// AutoTagProducts automatically tags products that need tags
// @Summary Auto-tag products
// @Description Automatically assign tags to products that don't have sufficient tags
// @Tags ML
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Maximum number of products to tag" default(50)
// @Success 200 {object} map[string]interface{} "Products auto-tagged successfully"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /ml/auto-tagging/auto-tag [post]
func AutoTagProducts(c *fiber.Ctx) error {
	_, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not authenticated",
		})
	}

	limit := c.QueryInt("limit", 50)
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	result, err := services.MLService.AutoTagProducts(limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to auto-tag products: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// GetTaggingInsights gets auto-tagging insights
// @Summary Get tagging insights
// @Description Get insights about the auto-tagging system performance and coverage
// @Tags ML
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Tagging insights retrieved successfully"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /ml/auto-tagging/insights [get]
func GetTaggingInsights(c *fiber.Ctx) error {
	_, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not authenticated",
		})
	}

	insights, err := services.MLService.GetTaggingInsights()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get tagging insights: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    insights,
	})
}

// SuggestProductDiscount suggests optimal discount for a specific product
// @Summary Suggest product discount
// @Description Get AI-suggested optimal discount for a specific product based on performance metrics
// @Tags ML
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID (UUID)"
// @Success 200 {object} map[string]interface{} "Product discount suggestion retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid product ID"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /ml/smart-discounts/suggest/product/{id} [get]
func SuggestProductDiscount(c *fiber.Ctx) error {
	_, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not authenticated",
		})
	}

	productIDStr := c.Params("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid product ID",
		})
	}

	discount, err := services.MLService.SuggestProductDiscount(productID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to suggest product discount: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    discount,
	})
}

// SuggestCategoryDiscounts suggests discounts for a product category
// @Summary Suggest category discounts
// @Description Get AI-suggested discounts for all products in a specific category
// @Tags ML
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category path string true "Product Category"
// @Success 200 {object} map[string]interface{} "Category discount suggestions retrieved successfully"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /ml/smart-discounts/suggest/category/{category} [get]
func SuggestCategoryDiscounts(c *fiber.Ctx) error {
	_, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not authenticated",
		})
	}

	category := c.Params("category")
	if category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Category is required",
		})
	}

	discounts, err := services.MLService.SuggestCategoryDiscounts(category)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to suggest category discounts: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    discounts,
	})
}

// GetDiscountInsights gets smart discount insights
// @Summary Get discount insights
// @Description Get insights about discount opportunities and performance across the platform
// @Tags ML
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Discount insights retrieved successfully"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /ml/smart-discounts/insights [get]
func GetDiscountInsights(c *fiber.Ctx) error {
	_, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not authenticated",
		})
	}

	insights, err := services.MLService.GetDiscountInsights()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get discount insights: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    insights,
	})
}

// InitializeMLServices initializes all new ML services
// @Summary Initialize ML services
// @Description Initialize or retrain all new ML services (sentiment, auto-tagging, smart discounts)
// @Tags ML
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "ML services initialized successfully"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /ml/initialize-services [post]
func InitializeMLServices(c *fiber.Ctx) error {
	_, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not authenticated",
		})
	}

	err := services.MLService.InitializeMLServices()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to initialize ML services: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "All ML services initialized successfully",
	})
}
