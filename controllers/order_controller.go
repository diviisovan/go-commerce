package controllers

import (
	"net/http"
	"strconv"

	"go-ecommerce/models"
	"go-ecommerce/services"

	"github.com/gin-gonic/gin"
)

// OrderController handles order-related HTTP requests
type OrderController struct {
	orderService   *services.OrderService
	paymentService *services.PaymentService
}

// NewOrderController creates a new order controller
func NewOrderController() *OrderController {
	return &OrderController{
		orderService:   services.NewOrderService(),
		paymentService: services.NewPaymentService(),
	}
}

// CreateOrder handles POST /api/orders
// @Summary      Create a new order
// @Description  Create an order from a shopping cart and process payment
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        order  body      object  true  "Order information"  example({"user_id": 1, "cart_id": 1, "shipping_cost": 15.00, "payment_method": "Credit Card"})
// @Success      201    {object}  map[string]interface{}
// @Failure      400    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /orders [post]
func (c *OrderController) CreateOrder(ctx *gin.Context) {
	var req struct {
		UserID        uint                 `json:"user_id" binding:"required"`
		CartID        uint                 `json:"cart_id" binding:"required"`
		ShippingCost  float64              `json:"shipping_cost"`
		PaymentMethod models.PaymentMethod `json:"payment_method" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := c.orderService.CreateOrder(req.UserID, req.CartID, req.ShippingCost)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process payment
	payment, err := c.paymentService.ProcessPayment(order.ID, order.TotalAmount, req.PaymentMethod)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"order":   order,
		"payment": payment,
	})
}

// GetOrder handles GET /api/orders/:id
// @Summary      Get order by ID
// @Description  Retrieve detailed information about a specific order
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Order ID"
// @Success      200  {object}  models.Order
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /orders/{id} [get]
func (c *OrderController) GetOrder(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	order, err := c.orderService.GetOrder(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	ctx.JSON(http.StatusOK, order)
}

// GetOrdersByUser handles GET /api/orders/user/:user_id
// @Summary      Get orders by user
// @Description  Retrieve all orders for a specific user
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        user_id  path      int  true  "User ID"
// @Success      200      {array}   models.Order
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /orders/user/{user_id} [get]
func (c *OrderController) GetOrdersByUser(ctx *gin.Context) {
	userID, err := strconv.ParseUint(ctx.Param("user_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	orders, err := c.orderService.GetOrdersByUser(uint(userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, orders)
}

// UpdateOrderStatus handles PUT /api/orders/:id/status
// @Summary      Update order status
// @Description  Update the status of an order (Pending, Confirmed, Shipped, Delivered, Cancelled)
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id      path      int                 true  "Order ID"
// @Param        status  body      object  true  "Status update"  example({"status": "Shipped"})
// @Success      200     {object}  map[string]string
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /orders/{id}/status [put]
func (c *OrderController) UpdateOrderStatus(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	var req struct {
		Status models.OrderStatus `json:"status" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.orderService.UpdateOrderStatus(uint(id), req.Status); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "order status updated successfully"})
}

// CancelOrder handles POST /api/orders/:id/cancel
// @Summary      Cancel an order
// @Description  Cancel an order and restore inventory
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Order ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Router       /orders/{id}/cancel [post]
func (c *OrderController) CancelOrder(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	if err := c.orderService.CancelOrder(uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "order cancelled successfully"})
}
