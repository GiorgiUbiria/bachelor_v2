package handlers

import (
	"bachelor_backend/database"
	"bachelor_backend/middleware"
	"bachelor_backend/models"

	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AddToCartRequest represents the request to add item to cart
type AddToCartRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	Quantity  int    `json:"quantity" validate:"required,min=1,max=100" example:"2"`
}

// UpdateCartItemRequest represents the request to update cart item
type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" validate:"required,min=0,max=100" example:"3"`
}

// GetCart returns the user's shopping cart
// @Summary Get shopping cart
// @Description Get the current user's shopping cart with all items and total
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Cart retrieved successfully"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /cart [get]
func GetCart(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get or create cart
	var cart models.ShoppingCart
	if err := database.DB.Where("user_id = ?", userID).
		Preload("CartItems.Product").
		First(&cart).Error; err != nil {
		// Create cart if it doesn't exist
		cart = models.ShoppingCart{UserID: userID}
		if err := database.DB.Create(&cart).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create cart",
			})
		}
	}

	// Calculate total
	var total float64
	for _, item := range cart.CartItems {
		total += item.Product.Price * float64(item.Quantity)
	}

	return c.JSON(fiber.Map{
		"cart":       cart,
		"total":      total,
		"item_count": len(cart.CartItems),
	})
}

// AddToCart adds an item to the user's cart
// @Summary Add item to cart
// @Description Add a product to the user's shopping cart with specified quantity
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AddToCartRequest true "Item to add to cart"
// @Success 200 {object} map[string]interface{} "Item added to cart successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request or insufficient stock"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /cart/add [post]
func AddToCart(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not authenticated",
		})
	}

	var req AddToCartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Validate request using middleware validation
	if err := middleware.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Validate product ID
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid product ID",
		})
	}

	// Check if product exists and get current stock
	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Product not found",
		})
	}

	// Check stock availability
	if product.Stock < req.Quantity {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Insufficient stock (Available: " + strconv.Itoa(product.Stock) + ", Requested: " + strconv.Itoa(req.Quantity) + ")",
		})
	}

	// Get or create cart using transaction
	var cart models.ShoppingCart
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).First(&cart).Error; err != nil {
			cart = models.ShoppingCart{UserID: userID}
			if err := tx.Create(&cart).Error; err != nil {
				return err
			}
		}

		// Check if item already exists in cart
		var existingItem models.CartItem
		if err := tx.Where("cart_id = ? AND product_id = ?", cart.ID, productID).
			First(&existingItem).Error; err == nil {
			// Update quantity with stock validation
			newQuantity := existingItem.Quantity + req.Quantity
			if newQuantity > product.Stock {
				return fmt.Errorf("total quantity would exceed available stock (Available: %d, Total requested: %d)", product.Stock, newQuantity)
			}
			if newQuantity > 100 { // Maximum quantity per item
				return fmt.Errorf("maximum quantity per item is 100")
			}

			existingItem.Quantity = newQuantity
			if err := tx.Save(&existingItem).Error; err != nil {
				return err
			}
		} else {
			// Create new cart item
			cartItem := models.CartItem{
				CartID:    cart.ID,
				ProductID: productID,
				Quantity:  req.Quantity,
			}
			if err := tx.Create(&cartItem).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Track user interaction
	go trackUserInteraction(userID, productID, "cart_add", c.Get("X-Session-ID"))

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Item added to cart successfully",
	})
}

// UpdateCartItem updates the quantity of an item in the cart
// @Summary Update cart item quantity
// @Description Update the quantity of a specific item in the user's cart
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Cart Item ID (UUID)"
// @Param request body UpdateCartItemRequest true "New quantity (0 to remove item)"
// @Success 200 {object} map[string]interface{} "Cart item updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request or insufficient stock"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 404 {object} map[string]interface{} "Cart item not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /cart/item/{id} [put]
func UpdateCartItem(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not authenticated",
		})
	}

	itemID := c.Params("id")
	id, err := uuid.Parse(itemID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid item ID",
		})
	}

	var req UpdateCartItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Validate request using middleware validation
	if err := middleware.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Get cart item with product information
	var cartItem models.CartItem
	if err := database.DB.Joins("JOIN shopping_carts ON cart_items.cart_id = shopping_carts.id").
		Where("cart_items.id = ? AND shopping_carts.user_id = ?", id, userID).
		Preload("Product").
		First(&cartItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Cart item not found",
		})
	}

	if req.Quantity == 0 {
		// Remove item from cart
		if err := database.DB.Delete(&cartItem).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to remove item from cart",
			})
		}
	} else {
		// Validate stock availability for new quantity
		if req.Quantity > cartItem.Product.Stock {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Insufficient stock (Available: " + strconv.Itoa(cartItem.Product.Stock) + ", Requested: " + strconv.Itoa(req.Quantity) + ")",
			})
		}

		// Update quantity
		cartItem.Quantity = req.Quantity
		if err := database.DB.Save(&cartItem).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to update cart item",
			})
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Cart item updated successfully",
	})
}

// RemoveFromCart removes an item from the cart
// @Summary Remove item from cart
// @Description Remove a specific item from the user's shopping cart
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Cart Item ID (UUID)"
// @Success 200 {object} map[string]interface{} "Item removed from cart successfully"
// @Failure 400 {object} map[string]interface{} "Invalid item ID"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 404 {object} map[string]interface{} "Cart item not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /cart/item/{id} [delete]
func RemoveFromCart(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	itemID := c.Params("id")
	id, err := uuid.Parse(itemID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid item ID",
		})
	}

	// Get and delete cart item
	var cartItem models.CartItem
	if err := database.DB.Joins("JOIN shopping_carts ON cart_items.cart_id = shopping_carts.id").
		Where("cart_items.id = ? AND shopping_carts.user_id = ?", id, userID).
		First(&cartItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Cart item not found",
		})
	}

	if err := database.DB.Delete(&cartItem).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove item from cart",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Item removed from cart successfully",
	})
}

// ClearCart removes all items from the user's cart
// @Summary Clear shopping cart
// @Description Remove all items from the user's shopping cart
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Cart cleared successfully"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 404 {object} map[string]interface{} "Cart not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /cart/clear [delete]
func ClearCart(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get cart
	var cart models.ShoppingCart
	if err := database.DB.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Cart not found",
		})
	}

	// Delete all cart items
	if err := database.DB.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to clear cart",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Cart cleared successfully",
	})
}
