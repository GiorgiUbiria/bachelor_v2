package handlers

import (
	"time"

	"bachelor_backend/database"
	"bachelor_backend/middleware"
	"bachelor_backend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max=255" example:"user@example.com"`
	Name     string `json:"name" validate:"required,min=2,max=100" example:"John Doe"`
	Password string `json:"password" validate:"required,min=8,max=128" example:"password123"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=1,max=128" example:"password123"`
}

// UpdateProfileRequest represents the profile update request payload
type UpdateProfileRequest struct {
	Name  string `json:"name" validate:"omitempty,min=2,max=100" example:"John Doe"`
	Phone string `json:"phone" validate:"omitempty,max=20" example:"+1234567890"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token string      `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  models.User `json:"user"`
}

// StandardErrorResponse represents a standard error response
type StandardErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"Invalid email or password"`
}

// UserProfileResponse represents a comprehensive user profile response
type UserProfileResponse struct {
	models.User
	Statistics     UserProfileStatistics `json:"statistics"`
	RecentActivity UserRecentActivity    `json:"recent_activity"`
}

// UserProfileStatistics represents user statistics
type UserProfileStatistics struct {
	TotalOrders             int64   `json:"total_orders"`
	TotalSpent              float64 `json:"total_spent"`
	AverageOrderValue       float64 `json:"average_order_value"`
	CartItemsCount          int64   `json:"cart_items_count"`
	CartTotalValue          float64 `json:"cart_total_value"`
	TotalInteractions       int64   `json:"total_interactions"`
	FavoriteCategory        string  `json:"favorite_category"`
	RecommendationsReceived int64   `json:"recommendations_received"`
	RecommendationsClicked  int64   `json:"recommendations_clicked"`
}

// UserRecentActivity represents recent user activity
type UserRecentActivity struct {
	RecentOrders       []models.Order           `json:"recent_orders"`
	RecentInteractions []models.UserInteraction `json:"recent_interactions"`
	RecentSearches     []models.SearchQuery     `json:"recent_searches"`
	RecentViews        []models.ProductView     `json:"recent_views"`
}

// Register handles user registration
// @Summary Register a new user
// @Description Create a new user account with email, name, and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} AuthResponse "User registered successfully"
// @Failure 400 {object} StandardErrorResponse "Invalid request body or validation error"
// @Failure 409 {object} StandardErrorResponse "User with this email already exists"
// @Failure 500 {object} StandardErrorResponse "Internal server error"
// @Router /auth/register [post]
func Register(c *fiber.Ctx) error {
	var req RegisterRequest

	// Parse and validate request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(StandardErrorResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
	}

	// Validate request using middleware validation
	if err := middleware.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(StandardErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	// Check if user already exists
	var existingUser models.User
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(StandardErrorResponse{
			Success: false,
			Error:   "User with this email already exists",
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(StandardErrorResponse{
			Success: false,
			Error:   "Failed to hash password",
		})
	}

	// Create user
	user := models.User{
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: string(hashedPassword),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(StandardErrorResponse{
			Success: false,
			Error:   "Failed to create user",
		})
	}

	// Generate JWT token
	token, err := generateJWTToken(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(StandardErrorResponse{
			Success: false,
			Error:   "Failed to generate token",
		})
	}

	// Create shopping cart for the user
	cart := models.ShoppingCart{
		UserID: user.ID,
	}
	database.DB.Create(&cart)

	return c.Status(fiber.StatusCreated).JSON(AuthResponse{
		Token: token,
		User:  user,
	})
}

// Login handles user login
// @Summary User login
// @Description Authenticate user with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse "Login successful"
// @Failure 400 {object} StandardErrorResponse "Invalid request body or validation error"
// @Failure 401 {object} StandardErrorResponse "Invalid email or password"
// @Failure 500 {object} StandardErrorResponse "Internal server error"
// @Router /auth/login [post]
func Login(c *fiber.Ctx) error {
	var req LoginRequest

	// Parse and validate request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(StandardErrorResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
	}

	// Validate request using middleware validation
	if err := middleware.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(StandardErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	// Find user by email
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(StandardErrorResponse{
			Success: false,
			Error:   "Invalid email or password",
		})
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(StandardErrorResponse{
			Success: false,
			Error:   "Invalid email or password",
		})
	}

	// Generate JWT token
	token, err := generateJWTToken(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(StandardErrorResponse{
			Success: false,
			Error:   "Failed to generate token",
		})
	}

	return c.JSON(AuthResponse{
		Token: token,
		User:  user,
	})
}

// GetProfile returns the current user's profile
// @Summary Get user profile
// @Description Get the authenticated user's comprehensive profile information including statistics and recent activity
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserProfileResponse "User profile retrieved successfully"
// @Failure 401 {object} StandardErrorResponse "User not authenticated"
// @Failure 404 {object} StandardErrorResponse "User not found"
// @Router /auth/profile [get]
func GetProfile(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(StandardErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
	}

	var user models.User
	if err := database.DB.Preload("Orders.OrderItems.Product").
		Preload("ShoppingCart.CartItems.Product").
		Preload("UserInteractions.Product").
		Preload("Recommendations.Product").
		First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(StandardErrorResponse{
			Success: false,
			Error:   "User not found",
		})
	}

	// Calculate user statistics
	statistics, err := calculateUserStatistics(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(StandardErrorResponse{
			Success: false,
			Error:   "Failed to calculate user statistics",
		})
	}

	// Get recent activity
	recentActivity, err := getUserRecentActivity(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(StandardErrorResponse{
			Success: false,
			Error:   "Failed to get recent activity",
		})
	}

	response := UserProfileResponse{
		User:           user,
		Statistics:     statistics,
		RecentActivity: recentActivity,
	}

	return c.JSON(response)
}

// calculateUserStatistics calculates comprehensive user statistics
func calculateUserStatistics(userID uuid.UUID) (UserProfileStatistics, error) {
	var stats UserProfileStatistics

	// Calculate order statistics
	var orderStats struct {
		TotalOrders int64
		TotalSpent  float64
	}

	if err := database.DB.Model(&models.Order{}).
		Where("user_id = ?", userID).
		Select("COUNT(*) as total_orders, COALESCE(SUM(total), 0) as total_spent").
		Scan(&orderStats).Error; err != nil {
		return stats, err
	}

	stats.TotalOrders = orderStats.TotalOrders
	stats.TotalSpent = orderStats.TotalSpent

	if stats.TotalOrders > 0 {
		stats.AverageOrderValue = stats.TotalSpent / float64(stats.TotalOrders)
	}

	// Calculate cart statistics
	var cartStats struct {
		ItemCount  int64
		TotalValue float64
	}

	if err := database.DB.Table("cart_items").
		Select("COUNT(*) as item_count, COALESCE(SUM(cart_items.quantity * products.price), 0) as total_value").
		Joins("JOIN shopping_carts ON cart_items.cart_id = shopping_carts.id").
		Joins("JOIN products ON cart_items.product_id = products.id").
		Where("shopping_carts.user_id = ?", userID).
		Scan(&cartStats).Error; err != nil {
		return stats, err
	}

	stats.CartItemsCount = cartStats.ItemCount
	stats.CartTotalValue = cartStats.TotalValue

	// Calculate interaction statistics
	if err := database.DB.Model(&models.UserInteraction{}).
		Where("user_id = ?", userID).
		Count(&stats.TotalInteractions).Error; err != nil {
		return stats, err
	}

	// Find favorite category
	var favoriteCategory struct {
		Category string
		Count    int64
	}

	if err := database.DB.Table("user_interactions").
		Select("products.category, COUNT(*) as count").
		Joins("JOIN products ON user_interactions.product_id = products.id").
		Where("user_interactions.user_id = ?", userID).
		Group("products.category").
		Order("count DESC").
		Limit(1).
		Scan(&favoriteCategory).Error; err == nil && favoriteCategory.Category != "" {
		stats.FavoriteCategory = favoriteCategory.Category
	}

	// Calculate recommendation statistics
	if err := database.DB.Model(&models.Recommendation{}).
		Where("user_id = ?", userID).
		Count(&stats.RecommendationsReceived).Error; err != nil {
		return stats, err
	}

	if err := database.DB.Model(&models.RecommendationFeedback{}).
		Where("user_id = ? AND feedback_type = ?", userID, "clicked").
		Count(&stats.RecommendationsClicked).Error; err != nil {
		return stats, err
	}

	return stats, nil
}

// getUserRecentActivity gets recent user activity
func getUserRecentActivity(userID uuid.UUID) (UserRecentActivity, error) {
	var activity UserRecentActivity

	// Get recent orders (last 5)
	if err := database.DB.Where("user_id = ?", userID).
		Preload("OrderItems.Product").
		Order("created_at DESC").
		Limit(5).
		Find(&activity.RecentOrders).Error; err != nil {
		return activity, err
	}

	// Get recent interactions (last 10)
	if err := database.DB.Where("user_id = ?", userID).
		Preload("Product").
		Order("created_at DESC").
		Limit(10).
		Find(&activity.RecentInteractions).Error; err != nil {
		return activity, err
	}

	// Get recent searches (last 10)
	if err := database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(10).
		Find(&activity.RecentSearches).Error; err != nil {
		return activity, err
	}

	// Get recent product views (last 10)
	if err := database.DB.Where("user_id = ?", userID).
		Preload("Product").
		Order("created_at DESC").
		Limit(10).
		Find(&activity.RecentViews).Error; err != nil {
		return activity, err
	}

	return activity, nil
}

// UpdateProfile updates the current user's profile
// @Summary Update user profile
// @Description Update the authenticated user's profile information
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateProfileRequest true "Profile update data"
// @Success 200 {object} models.User "Profile updated successfully"
// @Failure 400 {object} StandardErrorResponse "Invalid request body or validation error"
// @Failure 401 {object} StandardErrorResponse "User not authenticated"
// @Failure 404 {object} StandardErrorResponse "User not found"
// @Failure 500 {object} StandardErrorResponse "Internal server error"
// @Router /auth/profile [put]
func UpdateProfile(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(StandardErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
	}

	var req UpdateProfileRequest

	// Parse and validate request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(StandardErrorResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
	}

	// Validate request using middleware validation
	if err := middleware.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(StandardErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	// Update user
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(StandardErrorResponse{
			Success: false,
			Error:   "User not found",
		})
	}

	// Update fields if provided
	if req.Name != "" {
		user.Name = req.Name
	}

	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(StandardErrorResponse{
			Success: false,
			Error:   "Failed to update user",
		})
	}

	return c.JSON(user)
}

// generateJWTToken generates a JWT token for the user
func generateJWTToken(user models.User) (string, error) {
	claims := middleware.JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := getJWTSecret()
	return token.SignedString([]byte(secret))
}

// getJWTSecret gets JWT secret from environment
func getJWTSecret() string {
	// This should match the one in middleware/auth.go
	return middleware.GetJWTSecret()
}

// Helper function to expose JWT secret (we'll add this to middleware)
func init() {
	// This will be handled by the middleware package
}
