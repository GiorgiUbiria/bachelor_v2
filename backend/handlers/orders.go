package handlers

import (
	"fmt"
	"strconv"

	"bachelor_backend/database"
	"bachelor_backend/middleware"
	"bachelor_backend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	PaymentMethod   string   `json:"payment_method" validate:"required,oneof=credit_card debit_card paypal bank_transfer" example:"credit_card"`
	ShippingAddress string   `json:"shipping_address" validate:"required,min=10,max=500" example:"123 Main St, City, State 12345"`
	CartItemIDs     []string `json:"cart_item_ids,omitempty" example:"123e4567-e89b-12d3-a456-426614174000,456e7890-e89b-12d3-a456-426614174001"` // Optional: specific cart items to order (as string UUIDs)
}

// UpdateOrderStatusRequest represents the request to update order status
type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending processing shipped delivered cancelled" example:"processing"`
}

// GetOrders returns the user's order history
// @Summary Get user orders
// @Description Get paginated list of user's order history with order items
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "Orders retrieved successfully"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /orders [get]
func GetOrders(c *fiber.Ctx) error {
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

	// Count total orders first (without preload to avoid issues)
	if err := database.DB.Model(&models.Order{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count orders",
		})
	}

	// Get orders with preload
	if err := database.DB.Where("user_id = ?", userID).
		Preload("OrderItems.Product").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&orders).Error; err != nil {
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
// @Summary Get order by ID
// @Description Get detailed information about a specific order
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} models.Order "Order retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid order ID"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 404 {object} map[string]interface{} "Order not found"
// @Router /orders/{id} [get]
func GetOrder(c *fiber.Ctx) error {
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
// @Summary Create order from cart
// @Description Create a new order from the user's current cart items or specific cart items with atomic stock management
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateOrderRequest true "Order creation data. cart_item_ids is optional - if not provided, orders all cart items. Example: {\"payment_method\":\"credit_card\",\"shipping_address\":\"123 Main St, City, State 12345\",\"cart_item_ids\":[\"f29ab370-a0df-453a-a183-c444a60d1251\"]}"
// @Success 201 {object} map[string]interface{} "Order created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request, empty cart, or insufficient stock"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 404 {object} map[string]interface{} "Cart not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /orders [post]
func CreateOrder(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
	}

	// Validate request using middleware validation
	if err := middleware.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Parse and validate cart item IDs if provided
	var cartItemIDs []uuid.UUID
	if len(req.CartItemIDs) > 0 {
		for i, idStr := range req.CartItemIDs {
			id, err := uuid.Parse(idStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   fmt.Sprintf("Invalid cart item ID at index %d: %s", i, idStr),
				})
			}
			cartItemIDs = append(cartItemIDs, id)
		}
	}

	// Start transaction with proper isolation level for stock management
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// Get user's cart with row-level locking to prevent concurrent modifications
	var cart models.ShoppingCart
	if err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("user_id = ?", userID).
		Preload("CartItems.Product").
		First(&cart).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Cart not found",
		})
	}

	if len(cart.CartItems) == 0 {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cart is empty",
		})
	}

	// Filter cart items based on request
	var itemsToOrder []models.CartItem
	if len(cartItemIDs) > 0 {
		// Order specific cart items
		cartItemMap := make(map[uuid.UUID]models.CartItem)
		for _, item := range cart.CartItems {
			cartItemMap[item.ID] = item
		}

		for _, requestedID := range cartItemIDs {
			if item, exists := cartItemMap[requestedID]; exists {
				itemsToOrder = append(itemsToOrder, item)
			} else {
				tx.Rollback()
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Cart item not found: " + requestedID.String(),
				})
			}
		}
	} else {
		// Order all cart items (default behavior)
		itemsToOrder = cart.CartItems
	}

	if len(itemsToOrder) == 0 {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No valid cart items to order",
		})
	}

	// Validate stock availability with row-level locking and atomic updates
	var total float64
	stockUpdates := make(map[uuid.UUID]int) // Track stock updates for rollback if needed

	for _, item := range itemsToOrder {
		// Lock the product row to prevent concurrent stock modifications
		var product models.Product
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			First(&product, item.ProductID).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to verify product availability",
			})
		}

		// Check stock availability
		if product.Stock < item.Quantity {
			tx.Rollback()
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Insufficient stock for product: " + product.Name +
					" (Available: " + strconv.Itoa(product.Stock) +
					", Requested: " + strconv.Itoa(item.Quantity) + ")",
			})
		}

		// Calculate total using current product price (not cart price which might be outdated)
		total += product.Price * float64(item.Quantity)
		stockUpdates[product.ID] = product.Stock - item.Quantity
	}

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

	// Create order items and update stock atomically
	var orderedCartItemIDs []uuid.UUID
	for _, cartItem := range itemsToOrder {
		// Get current product price for order item
		var currentProduct models.Product
		if err := tx.First(&currentProduct, cartItem.ProductID).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get current product information",
			})
		}

		orderItem := models.OrderItem{
			OrderID:   order.ID,
			ProductID: cartItem.ProductID,
			Quantity:  cartItem.Quantity,
			Price:     currentProduct.Price, // Use current price, not cart price
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create order item",
			})
		}

		// Update product stock atomically
		newStock := stockUpdates[cartItem.ProductID]
		result := tx.Model(&models.Product{}).
			Where("id = ? AND stock >= ?", cartItem.ProductID, cartItem.Quantity).
			Update("stock", newStock)

		if result.Error != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update product stock",
			})
		}

		// Check if the update actually affected a row (stock was sufficient)
		if result.RowsAffected == 0 {
			tx.Rollback()
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Insufficient stock for product: " + currentProduct.Name + " (concurrent modification detected)",
			})
		}

		// Track purchase interaction safely
		go trackUserInteraction(userID, cartItem.ProductID, "purchase", c.Get("X-Session-ID"))

		// Track which cart items were ordered for removal
		orderedCartItemIDs = append(orderedCartItemIDs, cartItem.ID)
	}

	// Remove only the ordered cart items (not the entire cart)
	if err := tx.Where("id IN ?", orderedCartItemIDs).Delete(&models.CartItem{}).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove ordered items from cart",
		})
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to complete order transaction",
		})
	}

	// Load order with items for response (using fresh connection)
	if err := database.DB.Where("id = ?", order.ID).
		Preload("OrderItems.Product").
		First(&order).Error; err != nil {
		// Order was created successfully, but we can't load it for response
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message":             "Order created successfully",
			"order_id":            order.ID,
			"ordered_items_count": len(orderedCartItemIDs),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":             "Order created successfully",
		"order":               order,
		"ordered_items_count": len(orderedCartItemIDs),
	})
}

// UpdateOrderStatus updates the status of an order (admin only)
// @Summary Update order status
// @Description Update the status of an order with validation for status transitions (admin access required)
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID (UUID)"
// @Param request body UpdateOrderStatusRequest true "New order status"
// @Success 200 {object} map[string]interface{} "Order status updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request or status transition"
// @Failure 404 {object} map[string]interface{} "Order not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /orders/{id}/status [put]
func UpdateOrderStatus(c *fiber.Ctx) error {
	orderID := c.Params("id")
	id, err := uuid.Parse(orderID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid order ID",
		})
	}

	var req UpdateOrderStatusRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request using middleware validation
	if err := middleware.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Get current order to validate status transition
	var order models.Order
	if err := database.DB.First(&order, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Order not found",
		})
	}

	// Validate status transition logic
	if !isValidStatusTransition(order.Status, req.Status) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid status transition from " + order.Status + " to " + req.Status,
		})
	}

	// Update order status
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

// isValidStatusTransition validates if a status transition is allowed
func isValidStatusTransition(currentStatus, newStatus string) bool {
	// Define valid status transitions
	validTransitions := map[string][]string{
		"pending":    {"processing", "cancelled"},
		"processing": {"shipped", "cancelled"},
		"shipped":    {"delivered"},
		"delivered":  {}, // Final state
		"cancelled":  {}, // Final state
	}

	allowedTransitions, exists := validTransitions[currentStatus]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == newStatus {
			return true
		}
	}

	return false
}

// CancelOrder cancels an order (only if status is pending)
// @Summary Cancel order
// @Description Cancel a pending order and restore product stock
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} map[string]interface{} "Order cancelled successfully"
// @Failure 400 {object} map[string]interface{} "Invalid order ID or order cannot be cancelled"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 404 {object} map[string]interface{} "Order not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /orders/{id}/cancel [put]
func CancelOrder(c *fiber.Ctx) error {
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
// @Summary Get order statistics
// @Description Get comprehensive order statistics for the authenticated user
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Order statistics retrieved successfully"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Router /orders/stats [get]
func GetOrderStats(c *fiber.Ctx) error {
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
