package order

import (
	"fmt"
	"go-ecommerce/internal/cart"
	"go-ecommerce/internal/product"
	"time"
)

// Status represents the status of an order
type Status string

const (
	StatusPending   Status = "Pending"
	StatusConfirmed Status = "Confirmed"
	StatusShipped   Status = "Shipped"
	StatusDelivered Status = "Delivered"
	StatusCancelled Status = "Cancelled"
)

// Item represents an item in an order
type Item struct {
	Product  *product.Product
	Quantity int
}

// Order represents an order in the system
type Order struct {
	ID           int
	UserID       int
	Items        []*Item
	TotalAmount  float64
	Status       Status
	ShippingCost float64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Manager manages orders in the system
type Manager struct {
	Orders map[int]*Order
	NextID int
}

// NewManager creates a new order manager
func NewManager() *Manager {
	return &Manager{
		Orders: make(map[int]*Order),
		NextID: 1,
	}
}

// CreateOrder creates a new order from a shopping cart
func (om *Manager) CreateOrder(userID int, cart *cart.ShoppingCart, shippingCost float64) (*Order, error) {
	if len(cart.Items) == 0 {
		return nil, fmt.Errorf("cannot create order with empty cart")
	}

	now := time.Now()
	order := &Order{
		ID:           om.NextID,
		UserID:       userID,
		Items:        make([]*Item, len(cart.Items)),
		TotalAmount:  cart.GetTotal() + shippingCost,
		Status:       StatusPending,
		ShippingCost: shippingCost,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Copy cart items to order
	for i, item := range cart.Items {
		order.Items[i] = &Item{
			Product:  item.Product,
			Quantity: item.Quantity,
		}
	}

	om.Orders[om.NextID] = order
	om.NextID++

	return order, nil
}

// GetOrder retrieves an order by ID
func (om *Manager) GetOrder(id int) (*Order, bool) {
	order, exists := om.Orders[id]
	return order, exists
}

// UpdateOrderStatus updates the status of an order
func (om *Manager) UpdateOrderStatus(orderID int, status Status) error {
	order, exists := om.Orders[orderID]
	if !exists {
		return fmt.Errorf("order with ID %d not found", orderID)
	}

	order.Status = status
	order.UpdatedAt = time.Now()
	return nil
}

// CancelOrder cancels an order and restores inventory
func (om *Manager) CancelOrder(orderID int, catalog *product.Catalog) error {
	order, exists := om.Orders[orderID]
	if !exists {
		return fmt.Errorf("order with ID %d not found", orderID)
	}

	if order.Status == StatusDelivered {
		return fmt.Errorf("cannot cancel a delivered order")
	}

	// Restore inventory
	for _, item := range order.Items {
		catalog.UpdateStock(item.Product.ID, item.Quantity)
	}

	order.Status = StatusCancelled
	order.UpdatedAt = time.Now()
	return nil
}

// DisplayOrder displays order details
func (o *Order) DisplayOrder() {
	fmt.Printf("\n=== Order Details ===\n")
	fmt.Printf("Order ID:      %d\n", o.ID)
	fmt.Printf("User ID:       %d\n", o.UserID)
	fmt.Printf("Status:        %s\n", o.Status)
	fmt.Printf("Created At:    %s\n", o.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated At:    %s\n", o.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("\nItems:\n")

	subtotal := 0.0
	for _, item := range o.Items {
		itemTotal := item.Product.Price * float64(item.Quantity)
		subtotal += itemTotal
		fmt.Printf("  - %s (x%d) @ $%.2f = $%.2f\n",
			item.Product.Name, item.Quantity, item.Product.Price, itemTotal)
	}

	fmt.Printf("\nSubtotal:      $%.2f\n", subtotal)
	fmt.Printf("Shipping:      $%.2f\n", o.ShippingCost)
	fmt.Printf("Total Amount:  $%.2f\n", o.TotalAmount)
}

// ListOrdersByUser lists all orders for a specific user
func (om *Manager) ListOrdersByUser(userID int) {
	fmt.Printf("\n=== Orders for User ID %d ===\n", userID)
	found := false
	for _, order := range om.Orders {
		if order.UserID == userID {
			found = true
			fmt.Printf("Order ID: %d | Status: %s | Total: $%.2f | Date: %s\n",
				order.ID, order.Status, order.TotalAmount,
				order.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	}
	if !found {
		fmt.Println("No orders found for this user.")
	}
}
