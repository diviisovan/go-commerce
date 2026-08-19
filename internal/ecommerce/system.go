package ecommerce

import (
	"fmt"
	"go-ecommerce/internal/cart"
	"go-ecommerce/internal/order"
	"go-ecommerce/internal/payment"
	"go-ecommerce/internal/product"
	"go-ecommerce/internal/user"
)

// System represents the main eCommerce system
type System struct {
	Catalog          *product.Catalog
	UserManager      *user.Manager
	OrderManager     *order.Manager
	PaymentProcessor *payment.Processor
	Carts            map[int]*cart.ShoppingCart // Maps userID to their cart
}

// NewSystem creates a new eCommerce system instance
func NewSystem() *System {
	return &System{
		Catalog:          product.NewCatalog(),
		UserManager:      user.NewManager(),
		OrderManager:     order.NewManager(),
		PaymentProcessor: payment.NewProcessor(),
		Carts:            make(map[int]*cart.ShoppingCart),
	}
}

// GetOrCreateCart gets an existing cart or creates a new one for a user
func (ecs *System) GetOrCreateCart(userID int) *cart.ShoppingCart {
	shoppingCart, exists := ecs.Carts[userID]
	if !exists {
		shoppingCart = cart.NewShoppingCart(userID)
		ecs.Carts[userID] = shoppingCart
	}
	return shoppingCart
}

// Checkout processes the checkout for a user's cart
func (ecs *System) Checkout(userID int, shippingCost float64, paymentMethod payment.Method) (*order.Order, *payment.Payment, error) {
	shoppingCart, exists := ecs.Carts[userID]
	if !exists || len(shoppingCart.Items) == 0 {
		return nil, nil, fmt.Errorf("cart is empty")
	}

	// Verify stock availability
	for _, item := range shoppingCart.Items {
		if !ecs.Catalog.IsInStock(item.Product.ID, item.Quantity) {
			return nil, nil, fmt.Errorf("insufficient stock for product: %s", item.Product.Name)
		}
	}

	// Create order
	ord, err := ecs.OrderManager.CreateOrder(userID, shoppingCart, shippingCost)
	if err != nil {
		return nil, nil, err
	}

	// Update inventory
	for _, item := range shoppingCart.Items {
		ecs.Catalog.UpdateStock(item.Product.ID, -item.Quantity)
	}

	// Process payment
	pay, err := ecs.PaymentProcessor.ProcessPayment(ord.ID, ord.TotalAmount, paymentMethod)
	if err != nil {
		// If payment fails, restore inventory and cancel order
		ecs.OrderManager.CancelOrder(ord.ID, ecs.Catalog)
		return nil, nil, fmt.Errorf("payment failed: %v", err)
	}

	// Update order status to confirmed
	if pay.Status == payment.StatusCompleted {
		ecs.OrderManager.UpdateOrderStatus(ord.ID, order.StatusConfirmed)
	}

	// Clear cart after successful checkout
	shoppingCart.Clear()

	return ord, pay, nil
}
