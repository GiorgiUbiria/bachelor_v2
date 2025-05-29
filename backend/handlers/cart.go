package handlers

import (
	"bachelor_backend/database"
	"bachelor_backend/middleware"
	"bachelor_backend/models"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// AddToCartRequest represents the request to add item to cart
type AddToCartRequest struct {
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required,min=1"`
}

// UpdateCartItemRequest represents the request to update cart item
type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" validate:"required,min=0"`
}

// GetCart returns the user's shopping cart
func GetCart(c fiber.Ctx) error {
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
func AddToCart(c fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req AddToCartRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate product ID
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

	// Check stock
	if product.Stock < req.Quantity {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Insufficient stock",
		})
	}

	// Get or create cart
	var cart models.ShoppingCart
	if err := database.DB.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		cart = models.ShoppingCart{UserID: userID}
		if err := database.DB.Create(&cart).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create cart",
			})
		}
	}

	// Check if item already exists in cart
	var existingItem models.CartItem
	if err := database.DB.Where("cart_id = ? AND product_id = ?", cart.ID, productID).
		First(&existingItem).Error; err == nil {
		// Update quantity
		existingItem.Quantity += req.Quantity
		if err := database.DB.Save(&existingItem).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update cart item",
			})
		}
	} else {
		// Create new cart item
		cartItem := models.CartItem{
			CartID:    cart.ID,
			ProductID: productID,
			Quantity:  req.Quantity,
		}
		if err := database.DB.Create(&cartItem).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to add item to cart",
			})
		}
	}

	// Track user interaction
	go trackUserInteraction(userID, productID, "cart_add", c.Get("X-Session-ID"))

	return c.JSON(fiber.Map{
		"message": "Item added to cart successfully",
	})
}

// UpdateCartItem updates the quantity of an item in the cart
func UpdateCartItem(c fiber.Ctx) error {
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

	var req UpdateCartItemRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get cart item
	var cartItem models.CartItem
	if err := database.DB.Joins("JOIN shopping_carts ON cart_items.cart_id = shopping_carts.id").
		Where("cart_items.id = ? AND shopping_carts.user_id = ?", id, userID).
		First(&cartItem).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Cart item not found",
		})
	}

	if req.Quantity == 0 {
		// Remove item from cart
		if err := database.DB.Delete(&cartItem).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to remove item from cart",
			})
		}
	} else {
		// Update quantity
		cartItem.Quantity = req.Quantity
		if err := database.DB.Save(&cartItem).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update cart item",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Cart item updated successfully",
	})
}

// RemoveFromCart removes an item from the cart
func RemoveFromCart(c fiber.Ctx) error {
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
func ClearCart(c fiber.Ctx) error {
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
