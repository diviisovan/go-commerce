package services

import (
	"errors"
	"fmt"
	"go-ecommerce/database"
	"go-ecommerce/models"
)

// OrderService handles order-related business logic
type OrderService struct {
	productService *ProductService
	paymentService *PaymentService
}

// NewOrderService creates a new order service
func NewOrderService() *OrderService {
	return &OrderService{
		productService: NewProductService(),
		paymentService: NewPaymentService(),
	}
}

// CreateOrder creates a new order from a cart
func (s *OrderService) CreateOrder(userID uint, cartID uint, shippingCost float64) (*models.Order, error) {
	// Get cart with items
	var cart models.Cart
	if err := database.GetDB().Preload("Items.Product").First(&cart, cartID).Error; err != nil {
		return nil, errors.New("cart not found")
	}

	if len(cart.Items) == 0 {
		return nil, errors.New("cannot create order with empty cart")
	}

	// Calculate total
	var subtotal float64
	for _, item := range cart.Items {
		subtotal += item.Product.Price * float64(item.Quantity)
	}
	totalAmount := subtotal + shippingCost

	// Create order
	order := &models.Order{
		UserID:       userID,
		TotalAmount:  totalAmount,
		Status:       models.OrderStatusPending,
		ShippingCost: shippingCost,
	}

	// Create order items
	for _, cartItem := range cart.Items {
		orderItem := models.OrderItem{
			ProductID: cartItem.ProductID,
			Quantity:  cartItem.Quantity,
			Price:     cartItem.Product.Price,
		}
		order.Items = append(order.Items, orderItem)
	}

	// Start transaction
	tx := database.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Verify stock and update inventory
	for _, item := range cart.Items {
		if !s.productService.IsInStock(item.ProductID, item.Quantity) {
			tx.Rollback()
			return nil, fmt.Errorf("insufficient stock for product: %s", item.Product.Name)
		}
		if err := s.productService.UpdateStock(item.ProductID, -item.Quantity); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Create order
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Clear cart
	if err := tx.Delete(&models.CartItem{}, "cart_id = ?", cartID).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return order, nil
}

// GetOrder retrieves an order by ID
func (s *OrderService) GetOrder(id uint) (*models.Order, error) {
	var order models.Order
	err := database.GetDB().Preload("Items.Product").Preload("User").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetOrdersByUser retrieves all orders for a user
func (s *OrderService) GetOrdersByUser(userID uint) ([]models.Order, error) {
	var orders []models.Order
	err := database.GetDB().Preload("Items.Product").Where("user_id = ?", userID).Find(&orders).Error
	return orders, err
}

// UpdateOrderStatus updates the status of an order
func (s *OrderService) UpdateOrderStatus(orderID uint, status models.OrderStatus) error {
	return database.GetDB().Model(&models.Order{}).Where("id = ?", orderID).Update("status", status).Error
}

// CancelOrder cancels an order and restores inventory
func (s *OrderService) CancelOrder(orderID uint) error {
	var order models.Order
	if err := database.GetDB().Preload("Items").First(&order, orderID).Error; err != nil {
		return err
	}

	if order.Status == models.OrderStatusDelivered {
		return errors.New("cannot cancel a delivered order")
	}

	// Restore inventory
	for _, item := range order.Items {
		if err := s.productService.UpdateStock(item.ProductID, item.Quantity); err != nil {
			return err
		}
	}

	// Update order status
	return database.GetDB().Model(&order).Update("status", models.OrderStatusCancelled).Error
}
