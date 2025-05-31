package handlers

import (
	"strconv"
	"time"

	"bachelor_backend/database"
	"bachelor_backend/middleware"
	"bachelor_backend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Favorite-related request/response types
type AddFavoriteRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// Upvote-related request/response types
type AddUpvoteRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// Comment-related request/response types
type AddCommentRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	Content   string `json:"content" validate:"required,min=1,max=1000" example:"Great product!"`
	Rating    int    `json:"rating" validate:"required,min=1,max=5" example:"5"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" validate:"omitempty,min=1,max=1000" example:"Updated comment"`
	Rating  int    `json:"rating" validate:"omitempty,min=1,max=5" example:"4"`
}

// Tag-related request/response types
type CreateTagRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=50" example:"Electronics"`
	Description string `json:"description" validate:"omitempty,max=255" example:"Electronic devices and gadgets"`
	Color       string `json:"color" validate:"omitempty,len=7" example:"#FF5733"`
}

type AddProductTagRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	TagID     string `json:"tag_id" validate:"required,uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// Discount-related request/response types
type CreateDiscountRequest struct {
	ProductID         *string   `json:"product_id" validate:"omitempty,uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	Category          *string   `json:"category" validate:"omitempty,min=1,max=100" example:"Electronics"`
	DiscountType      string    `json:"discount_type" validate:"required,oneof=percentage fixed_amount" example:"percentage"`
	DiscountValue     float64   `json:"discount_value" validate:"required,min=0" example:"20.0"`
	MinOrderAmount    float64   `json:"min_order_amount" validate:"omitempty,min=0" example:"100.0"`
	MaxDiscountAmount float64   `json:"max_discount_amount" validate:"omitempty,min=0" example:"50.0"`
	StartDate         time.Time `json:"start_date" validate:"required" example:"2024-01-01T00:00:00Z"`
	EndDate           time.Time `json:"end_date" validate:"required" example:"2024-12-31T23:59:59Z"`
	UsageLimit        int       `json:"usage_limit" validate:"omitempty,min=0" example:"100"`
}

// FAVORITES HANDLERS

// AddFavorite adds a product to user's favorites
// @Summary Add product to favorites
// @Description Add a product to the user's favorite list
// @Tags Favorites
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AddFavoriteRequest true "Product to add to favorites"
// @Success 201 {object} map[string]interface{} "Product added to favorites successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request or product already in favorites"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Router /favorites [post]
func AddFavorite(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req AddFavoriteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := middleware.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	// Check if product exists
	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Product not found",
		})
	}

	// Check if already favorited
	var existingFavorite models.Favorite
	if err := database.DB.Where("user_id = ? AND product_id = ?", userID, productID).
		First(&existingFavorite).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Product already in favorites",
		})
	}

	// Create favorite
	favorite := models.Favorite{
		UserID:    userID,
		ProductID: productID,
	}

	if err := database.DB.Create(&favorite).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add favorite",
		})
	}

	// Track interaction
	go trackUserInteraction(userID, productID, "favorite", c.Get("X-Session-ID"))

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "Product added to favorites successfully",
		"favorite": favorite,
	})
}

// RemoveFavorite removes a product from user's favorites
// @Summary Remove product from favorites
// @Description Remove a product from the user's favorite list
// @Tags Favorites
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param product_id path string true "Product ID (UUID)"
// @Success 200 {object} map[string]interface{} "Product removed from favorites successfully"
// @Failure 400 {object} map[string]interface{} "Invalid product ID"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 404 {object} map[string]interface{} "Favorite not found"
// @Router /favorites/{product_id} [delete]
func RemoveFavorite(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	productID, err := uuid.Parse(c.Params("product_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	// Find and delete favorite
	var favorite models.Favorite
	if err := database.DB.Where("user_id = ? AND product_id = ?", userID, productID).
		First(&favorite).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Favorite not found",
		})
	}

	if err := database.DB.Delete(&favorite).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove favorite",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Product removed from favorites successfully",
	})
}

// GetFavorites returns user's favorite products
// @Summary Get user favorites
// @Description Get paginated list of user's favorite products
// @Tags Favorites
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{} "Favorites retrieved successfully"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Router /favorites [get]
func GetFavorites(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset := (page - 1) * limit

	var favorites []models.Favorite
	var total int64

	database.DB.Model(&models.Favorite{}).Where("user_id = ?", userID).Count(&total)

	if err := database.DB.Where("user_id = ?", userID).
		Preload("Product").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&favorites).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch favorites",
		})
	}

	return c.JSON(fiber.Map{
		"favorites": favorites,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// UPVOTES HANDLERS

// AddUpvote adds an upvote to a product
// @Summary Upvote product
// @Description Add an upvote to a product
// @Tags Upvotes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AddUpvoteRequest true "Product to upvote"
// @Success 201 {object} map[string]interface{} "Product upvoted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request or already upvoted"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Router /upvotes [post]
func AddUpvote(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req AddUpvoteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := middleware.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	// Check if product exists
	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Product not found",
		})
	}

	// Check if already upvoted
	var existingUpvote models.Upvote
	if err := database.DB.Where("user_id = ? AND product_id = ?", userID, productID).
		First(&existingUpvote).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Product already upvoted",
		})
	}

	// Create upvote
	upvote := models.Upvote{
		UserID:    userID,
		ProductID: productID,
	}

	if err := database.DB.Create(&upvote).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add upvote",
		})
	}

	// Track interaction
	go trackUserInteraction(userID, productID, "upvote", c.Get("X-Session-ID"))

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Product upvoted successfully",
		"upvote":  upvote,
	})
}

// RemoveUpvote removes an upvote from a product
// @Summary Remove upvote
// @Description Remove an upvote from a product
// @Tags Upvotes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param product_id path string true "Product ID (UUID)"
// @Success 200 {object} map[string]interface{} "Upvote removed successfully"
// @Failure 400 {object} map[string]interface{} "Invalid product ID"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 404 {object} map[string]interface{} "Upvote not found"
// @Router /upvotes/{product_id} [delete]
func RemoveUpvote(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	productID, err := uuid.Parse(c.Params("product_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	// Find and delete upvote
	var upvote models.Upvote
	if err := database.DB.Where("user_id = ? AND product_id = ?", userID, productID).
		First(&upvote).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Upvote not found",
		})
	}

	if err := database.DB.Delete(&upvote).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove upvote",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Upvote removed successfully",
	})
}

// GetProductUpvotes returns upvote count for a product
// @Summary Get product upvotes
// @Description Get upvote count and user's upvote status for a product
// @Tags Upvotes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param product_id path string true "Product ID (UUID)"
// @Success 200 {object} map[string]interface{} "Upvotes retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid product ID"
// @Router /upvotes/{product_id} [get]
func GetProductUpvotes(c *fiber.Ctx) error {
	productID, err := uuid.Parse(c.Params("product_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	var upvoteCount int64
	database.DB.Model(&models.Upvote{}).Where("product_id = ?", productID).Count(&upvoteCount)

	// Check if current user has upvoted (if authenticated)
	var userUpvoted bool
	if userID, ok := middleware.GetUserID(c); ok {
		var upvote models.Upvote
		if err := database.DB.Where("user_id = ? AND product_id = ?", userID, productID).
			First(&upvote).Error; err == nil {
			userUpvoted = true
		}
	}

	return c.JSON(fiber.Map{
		"product_id":   productID,
		"upvote_count": upvoteCount,
		"user_upvoted": userUpvoted,
	})
}

// COMMENTS HANDLERS

// AddComment adds a comment to a product
// @Summary Add product comment
// @Description Add a comment and rating to a product
// @Tags Comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AddCommentRequest true "Comment to add"
// @Success 201 {object} map[string]interface{} "Comment added successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Router /comments [post]
func AddComment(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req AddCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := middleware.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	// Check if product exists
	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Product not found",
		})
	}

	// Create comment
	comment := models.Comment{
		UserID:    userID,
		ProductID: productID,
		Content:   req.Content,
		Rating:    req.Rating,
	}

	if err := database.DB.Create(&comment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add comment",
		})
	}

	// Load comment with user info for response
	database.DB.Preload("User").First(&comment, comment.ID)

	// Track interaction
	go trackUserInteraction(userID, productID, "comment", c.Get("X-Session-ID"))

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Comment added successfully",
		"comment": comment,
	})
}

// GetProductComments returns comments for a product
// @Summary Get product comments
// @Description Get paginated list of comments for a product
// @Tags Comments
// @Accept json
// @Produce json
// @Param product_id path string true "Product ID (UUID)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{} "Comments retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid product ID"
// @Router /comments/{product_id} [get]
func GetProductComments(c *fiber.Ctx) error {
	productID, err := uuid.Parse(c.Params("product_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset := (page - 1) * limit

	var comments []models.Comment
	var total int64

	database.DB.Model(&models.Comment{}).Where("product_id = ?", productID).Count(&total)

	if err := database.DB.Where("product_id = ?", productID).
		Preload("User").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&comments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch comments",
		})
	}

	// Calculate average rating
	var avgRating float64
	database.DB.Model(&models.Comment{}).
		Where("product_id = ?", productID).
		Select("AVG(rating)").
		Scan(&avgRating)

	return c.JSON(fiber.Map{
		"comments": comments,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
		"average_rating": avgRating,
	})
}

// UpdateComment updates a user's comment
// @Summary Update comment
// @Description Update a user's own comment
// @Tags Comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param comment_id path string true "Comment ID (UUID)"
// @Param request body UpdateCommentRequest true "Updated comment data"
// @Success 200 {object} map[string]interface{} "Comment updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 403 {object} map[string]interface{} "Not authorized to update this comment"
// @Failure 404 {object} map[string]interface{} "Comment not found"
// @Router /comments/{comment_id} [put]
func UpdateComment(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	commentID, err := uuid.Parse(c.Params("comment_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid comment ID",
		})
	}

	var req UpdateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := middleware.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Find comment and verify ownership
	var comment models.Comment
	if err := database.DB.Where("id = ? AND user_id = ?", commentID, userID).
		First(&comment).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Comment not found or not authorized",
		})
	}

	// Update fields if provided
	if req.Content != "" {
		comment.Content = req.Content
	}
	if req.Rating > 0 {
		comment.Rating = req.Rating
	}

	if err := database.DB.Save(&comment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update comment",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Comment updated successfully",
		"comment": comment,
	})
}

// DeleteComment deletes a user's comment
// @Summary Delete comment
// @Description Delete a user's own comment
// @Tags Comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param comment_id path string true "Comment ID (UUID)"
// @Success 200 {object} map[string]interface{} "Comment deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid comment ID"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 404 {object} map[string]interface{} "Comment not found"
// @Router /comments/{comment_id} [delete]
func DeleteComment(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	commentID, err := uuid.Parse(c.Params("comment_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid comment ID",
		})
	}

	// Find and delete comment (only if user owns it)
	var comment models.Comment
	if err := database.DB.Where("id = ? AND user_id = ?", commentID, userID).
		First(&comment).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Comment not found or not authorized",
		})
	}

	if err := database.DB.Delete(&comment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete comment",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Comment deleted successfully",
	})
}

// TAGS HANDLERS

// CreateTag creates a new product tag
// @Summary Create tag
// @Description Create a new product tag (admin only)
// @Tags Tags
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateTagRequest true "Tag data"
// @Success 201 {object} map[string]interface{} "Tag created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request or tag already exists"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Router /tags [post]
func CreateTag(c *fiber.Ctx) error {
	var req CreateTagRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := middleware.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Check if tag already exists
	var existingTag models.Tag
	if err := database.DB.Where("name = ?", req.Name).First(&existingTag).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Tag already exists",
		})
	}

	// Create tag
	tag := models.Tag{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
	}

	if err := database.DB.Create(&tag).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create tag",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Tag created successfully",
		"tag":     tag,
	})
}

// GetTags returns all available tags
// @Summary Get all tags
// @Description Get list of all available product tags
// @Tags Tags
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Tags retrieved successfully"
// @Router /tags [get]
func GetTags(c *fiber.Ctx) error {
	var tags []models.Tag
	if err := database.DB.Order("name").Find(&tags).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch tags",
		})
	}

	return c.JSON(fiber.Map{
		"tags": tags,
	})
}

// AddProductTag adds a tag to a product
// @Summary Add tag to product
// @Description Add a tag to a product (admin only)
// @Tags Tags
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AddProductTagRequest true "Product and tag IDs"
// @Success 201 {object} map[string]interface{} "Tag added to product successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request or tag already added"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 404 {object} map[string]interface{} "Product or tag not found"
// @Router /tags/products [post]
func AddProductTag(c *fiber.Ctx) error {
	var req AddProductTagRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := middleware.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	tagID, err := uuid.Parse(req.TagID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid tag ID",
		})
	}

	// Check if product and tag exist
	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Product not found",
		})
	}

	var tag models.Tag
	if err := database.DB.First(&tag, tagID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Tag not found",
		})
	}

	// Check if association already exists
	var existingProductTag models.ProductTag
	if err := database.DB.Where("product_id = ? AND tag_id = ?", productID, tagID).
		First(&existingProductTag).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Tag already added to product",
		})
	}

	// Create association
	productTag := models.ProductTag{
		ProductID: productID,
		TagID:     tagID,
	}

	if err := database.DB.Create(&productTag).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add tag to product",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Tag added to product successfully",
		"product_tag": productTag,
	})
}

// GetProductTags returns tags for a product
// @Summary Get product tags
// @Description Get all tags associated with a product
// @Tags Tags
// @Accept json
// @Produce json
// @Param product_id path string true "Product ID (UUID)"
// @Success 200 {object} map[string]interface{} "Product tags retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid product ID"
// @Router /tags/products/{product_id} [get]
func GetProductTags(c *fiber.Ctx) error {
	productID, err := uuid.Parse(c.Params("product_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	var product models.Product
	if err := database.DB.Preload("Tags").First(&product, productID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Product not found",
		})
	}

	return c.JSON(fiber.Map{
		"product_id": productID,
		"tags":       product.Tags,
	})
}

// DISCOUNTS HANDLERS

// CreateDiscount creates a new discount
// @Summary Create discount
// @Description Create a new product or category discount (admin only)
// @Tags Discounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateDiscountRequest true "Discount data"
// @Success 201 {object} map[string]interface{} "Discount created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Router /discounts [post]
func CreateDiscount(c *fiber.Ctx) error {
	var req CreateDiscountRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := middleware.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Validate that either ProductID or Category is provided, but not both
	if (req.ProductID == nil && req.Category == nil) || (req.ProductID != nil && req.Category != nil) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Either product_id or category must be provided, but not both",
		})
	}

	// Validate date range
	if req.EndDate.Before(req.StartDate) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "End date must be after start date",
		})
	}

	// Create discount
	discount := models.Discount{
		DiscountType:      req.DiscountType,
		DiscountValue:     req.DiscountValue,
		MinOrderAmount:    req.MinOrderAmount,
		MaxDiscountAmount: req.MaxDiscountAmount,
		StartDate:         req.StartDate,
		EndDate:           req.EndDate,
		UsageLimit:        req.UsageLimit,
	}

	if req.ProductID != nil {
		productID, err := uuid.Parse(*req.ProductID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid product ID",
			})
		}
		discount.ProductID = &productID
	}

	if req.Category != nil {
		discount.Category = req.Category
	}

	if err := database.DB.Create(&discount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create discount",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "Discount created successfully",
		"discount": discount,
	})
}

// GetActiveDiscounts returns currently active discounts
// @Summary Get active discounts
// @Description Get list of currently active discounts
// @Tags Discounts
// @Accept json
// @Produce json
// @Param product_id query string false "Filter by product ID"
// @Param category query string false "Filter by category"
// @Success 200 {object} map[string]interface{} "Active discounts retrieved successfully"
// @Router /discounts/active [get]
func GetActiveDiscounts(c *fiber.Ctx) error {
	now := time.Now()
	query := database.DB.Where("is_active = ? AND start_date <= ? AND end_date >= ?", true, now, now)

	// Apply filters
	if productID := c.Query("product_id"); productID != "" {
		if id, err := uuid.Parse(productID); err == nil {
			query = query.Where("product_id = ?", id)
		}
	}

	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	var discounts []models.Discount
	if err := query.Preload("Product").Find(&discounts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch discounts",
		})
	}

	return c.JSON(fiber.Map{
		"discounts": discounts,
	})
}
