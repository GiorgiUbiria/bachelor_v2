package handlers

import (
	"strconv"

	"bachelor_backend/database"
	"bachelor_backend/middleware"
	"bachelor_backend/models"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	PaymentMethod   string `json:"payment_method" validate:"required"`
	ShippingAddress string `json:"shipping_address" validate:"required"`
}

// GetOrders returns the user's order history
func GetOrders(c fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	var orders []models.Order
	var total int64

	query := database.DB.Where("user_id = ?", userID).
		Preload("OrderItems.Product").
		Order("created_at DESC")

	query.Count(&total)

	if err := query.Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch orders",
		})
	}

	return c.JSON(fiber.Map{
		"orders": orders,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetOrder returns a specific order by ID
func GetOrder(c fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	orderID := c.Params("id")
	id, err := uuid.Parse(orderID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid order ID",
		})
	}

	var order models.Order
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).
		Preload("OrderItems.Product").
		First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Order not found",
		})
	}

	return c.JSON(order)
}

// CreateOrder creates a new order from the user's cart
func CreateOrder(c fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req CreateOrderRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get user's cart
	var cart models.ShoppingCart
	if err := database.DB.Where("user_id = ?", userID).
		Preload("CartItems.Product").
		First(&cart).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Cart not found",
		})
	}

	if len(cart.CartItems) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cart is empty",
		})
	}

	// Calculate total and validate stock
	var total float64
	for _, item := range cart.CartItems {
		if item.Product.Stock < item.Quantity {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Insufficient stock for product: " + item.Product.Name,
			})
		}
		total += item.Product.Price * float64(item.Quantity)
	}

	// Start transaction
	tx := database.DB.Begin()

	// Create order
	order := models.Order{
		UserID: userID,
		Total:  total,
		Status: "pending",
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create order",
		})
	}

	// Create order items and update stock
	for _, cartItem := range cart.CartItems {
		orderItem := models.OrderItem{
			OrderID:   order.ID,
			ProductID: cartItem.ProductID,
			Quantity:  cartItem.Quantity,
			Price:     cartItem.Product.Price,
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create order item",
			})
		}

		// Update product stock
		if err := tx.Model(&cartItem.Product).Update("stock", cartItem.Product.Stock-cartItem.Quantity).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update product stock",
			})
		}

		// Track purchase interaction
		go trackUserInteraction(userID, cartItem.ProductID, "purchase", c.Get("X-Session-ID"))
	}

	// Clear cart
	if err := tx.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to clear cart",
		})
	}

	// Commit transaction
	tx.Commit()

	// Load order with items for response
	database.DB.Where("id = ?", order.ID).
		Preload("OrderItems.Product").
		First(&order)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Order created successfully",
		"order":   order,
	})
}

// UpdateOrderStatus updates the status of an order (admin only)
func UpdateOrderStatus(c fiber.Ctx) error {
	orderID := c.Params("id")
	id, err := uuid.Parse(orderID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid order ID",
		})
	}

	var req struct {
		Status string `json:"status" validate:"required"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate status
	validStatuses := []string{"pending", "processing", "shipped", "delivered", "cancelled"}
	isValid := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValid = true
			break
		}
	}

	if !isValid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid order status",
		})
	}

	// Update order status
	var order models.Order
	if err := database.DB.First(&order, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Order not found",
		})
	}

	order.Status = req.Status
	if err := database.DB.Save(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update order status",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Order status updated successfully",
		"order":   order,
	})
}

// CancelOrder cancels an order (only if status is pending)
func CancelOrder(c fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	orderID := c.Params("id")
	id, err := uuid.Parse(orderID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid order ID",
		})
	}

	// Get order
	var order models.Order
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).
		Preload("OrderItems.Product").
		First(&order).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Order not found",
		})
	}

	if order.Status != "pending" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order cannot be cancelled",
		})
	}

	// Start transaction
	tx := database.DB.Begin()

	// Restore product stock
	for _, item := range order.OrderItems {
		if err := tx.Model(&item.Product).Update("stock", item.Product.Stock+item.Quantity).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to restore product stock",
			})
		}
	}

	// Update order status
	order.Status = "cancelled"
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to cancel order",
		})
	}

	tx.Commit()

	return c.JSON(fiber.Map{
		"message": "Order cancelled successfully",
		"order":   order,
	})
}

// GetOrderStats returns order statistics for analytics
func GetOrderStats(c fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var stats struct {
		TotalOrders     int64   `json:"total_orders"`
		TotalSpent      float64 `json:"total_spent"`
		PendingOrders   int64   `json:"pending_orders"`
		CompletedOrders int64   `json:"completed_orders"`
	}

	// Get total orders and spent
	database.DB.Model(&models.Order{}).
		Where("user_id = ?", userID).
		Count(&stats.TotalOrders)

	database.DB.Model(&models.Order{}).
		Where("user_id = ? AND status IN ?", userID, []string{"delivered", "completed"}).
		Select("COALESCE(SUM(total), 0)").
		Scan(&stats.TotalSpent)

	// Get pending orders
	database.DB.Model(&models.Order{}).
		Where("user_id = ? AND status = ?", userID, "pending").
		Count(&stats.PendingOrders)

	// Get completed orders
	database.DB.Model(&models.Order{}).
		Where("user_id = ? AND status IN ?", userID, []string{"delivered", "completed"}).
		Count(&stats.CompletedOrders)

	return c.JSON(stats)
}
