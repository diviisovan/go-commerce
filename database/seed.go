package database

import (
	"fmt"
	"go-ecommerce/models"
	"log"
)

// SeedDatabase populates the database with sample data
func SeedDatabase() error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	log.Println("Seeding database with sample data...")

	// Create sample users
	users := []models.User{
		{Name: "John Doe", Email: "john.doe@example.com", Address: "123 Main St, City, State 12345", Phone: "+1-555-0101", IsActive: true},
		{Name: "Jane Smith", Email: "jane.smith@example.com", Address: "456 Oak Ave, City, State 67890", Phone: "+1-555-0102", IsActive: true},
		{Name: "Bob Johnson", Email: "bob.johnson@example.com", Address: "789 Pine Rd, City, State 54321", Phone: "+1-555-0103", IsActive: true},
		{Name: "Alice Williams", Email: "alice.williams@example.com", Address: "321 Elm St, City, State 98765", Phone: "+1-555-0104", IsActive: true},
	}

	for i := range users {
		if err := DB.Create(&users[i]).Error; err != nil {
			log.Printf("Warning: Failed to create user %s: %v", users[i].Email, err)
		} else {
			log.Printf("Created user: %s (ID: %d)", users[i].Name, users[i].ID)
		}
	}

	// Create sample products
	products := []models.Product{
		{Name: "Laptop", Description: "High-performance laptop with 16GB RAM and 512GB SSD", Price: 999.99, Stock: 10, Category: "Electronics"},
		{Name: "Wireless Mouse", Description: "Ergonomic wireless mouse with long battery life", Price: 29.99, Stock: 50, Category: "Electronics"},
		{Name: "Mechanical Keyboard", Description: "RGB mechanical keyboard with blue switches", Price: 89.99, Stock: 30, Category: "Electronics"},
		{Name: "USB-C Cable", Description: "Fast charging USB-C cable, 6ft length", Price: 15.99, Stock: 100, Category: "Accessories"},
		{Name: "Monitor", Description: "27-inch 4K UHD monitor with HDR", Price: 349.99, Stock: 15, Category: "Electronics"},
		{Name: "Webcam", Description: "1080p HD webcam with autofocus", Price: 79.99, Stock: 25, Category: "Electronics"},
		{Name: "Headphones", Description: "Wireless noise-cancelling headphones", Price: 199.99, Stock: 20, Category: "Audio"},
		{Name: "Desk Lamp", Description: "LED desk lamp with adjustable brightness", Price: 39.99, Stock: 40, Category: "Accessories"},
		{Name: "Laptop Stand", Description: "Aluminum laptop stand for better ergonomics", Price: 49.99, Stock: 35, Category: "Accessories"},
		{Name: "External SSD", Description: "1TB external SSD with USB-C connection", Price: 129.99, Stock: 18, Category: "Storage"},
	}

	for i := range products {
		if err := DB.Create(&products[i]).Error; err != nil {
			log.Printf("Warning: Failed to create product %s: %v", products[i].Name, err)
		} else {
			log.Printf("Created product: %s (ID: %d, Price: $%.2f)", products[i].Name, products[i].ID, products[i].Price)
		}
	}

	// Create sample carts with items
	if len(users) > 0 && len(products) > 0 {
		// Cart for user 1
		cart1 := models.Cart{UserID: users[0].ID}
		if err := DB.Create(&cart1).Error; err == nil {
			cartItems1 := []models.CartItem{
				{CartID: cart1.ID, ProductID: products[0].ID, Quantity: 1}, // Laptop
				{CartID: cart1.ID, ProductID: products[1].ID, Quantity: 2}, // Wireless Mouse x2
				{CartID: cart1.ID, ProductID: products[3].ID, Quantity: 3}, // USB-C Cable x3
			}
			for _, item := range cartItems1 {
				DB.Create(&item)
			}
			log.Printf("Created cart for user %s with %d items", users[0].Name, len(cartItems1))
		}

		// Cart for user 2
		cart2 := models.Cart{UserID: users[1].ID}
		if err := DB.Create(&cart2).Error; err == nil {
			cartItems2 := []models.CartItem{
				{CartID: cart2.ID, ProductID: products[2].ID, Quantity: 1}, // Mechanical Keyboard
				{CartID: cart2.ID, ProductID: products[3].ID, Quantity: 2}, // USB-C Cable x2
			}
			for _, item := range cartItems2 {
				DB.Create(&item)
			}
			log.Printf("Created cart for user %s with %d items", users[1].Name, len(cartItems2))
		}
	}

	// Create sample orders with order items
	if len(users) > 0 && len(products) > 0 {
		// Order 1 for user 1
		order1 := models.Order{
			UserID:       users[0].ID,
			TotalAmount:  1122.94,
			Status:       models.OrderStatusConfirmed,
			ShippingCost: 15.00,
		}
		if err := DB.Create(&order1).Error; err == nil {
			orderItems1 := []models.OrderItem{
				{OrderID: order1.ID, ProductID: products[0].ID, Quantity: 1, Price: products[0].Price}, // Laptop
				{OrderID: order1.ID, ProductID: products[1].ID, Quantity: 2, Price: products[1].Price}, // Wireless Mouse x2
				{OrderID: order1.ID, ProductID: products[3].ID, Quantity: 3, Price: products[3].Price}, // USB-C Cable x3
			}
			for _, item := range orderItems1 {
				DB.Create(&item)
			}
			log.Printf("Created order %d for user %s with %d items (Total: $%.2f)", order1.ID, users[0].Name, len(orderItems1), order1.TotalAmount)
		}

		// Order 2 for user 2
		order2 := models.Order{
			UserID:       users[1].ID,
			TotalAmount:  131.97,
			Status:       models.OrderStatusShipped,
			ShippingCost: 10.00,
		}
		if err := DB.Create(&order2).Error; err == nil {
			orderItems2 := []models.OrderItem{
				{OrderID: order2.ID, ProductID: products[2].ID, Quantity: 1, Price: products[2].Price}, // Mechanical Keyboard
				{OrderID: order2.ID, ProductID: products[3].ID, Quantity: 2, Price: products[3].Price}, // USB-C Cable x2
			}
			for _, item := range orderItems2 {
				DB.Create(&item)
			}
			log.Printf("Created order %d for user %s with %d items (Total: $%.2f)", order2.ID, users[1].Name, len(orderItems2), order2.TotalAmount)
		}

		// Order 3 for user 1 (Delivered)
		order3 := models.Order{
			UserID:       users[0].ID,
			TotalAmount:  199.99,
			Status:       models.OrderStatusDelivered,
			ShippingCost: 0.00,
		}
		if err := DB.Create(&order3).Error; err == nil {
			orderItems3 := []models.OrderItem{
				{OrderID: order3.ID, ProductID: products[6].ID, Quantity: 1, Price: products[6].Price}, // Headphones
			}
			for _, item := range orderItems3 {
				DB.Create(&item)
			}
			log.Printf("Created order %d for user %s with %d items (Total: $%.2f)", order3.ID, users[0].Name, len(orderItems3), order3.TotalAmount)
		}

		// Order 4 for user 3
		order4 := models.Order{
			UserID:       users[2].ID,
			TotalAmount:  429.98,
			Status:       models.OrderStatusPending,
			ShippingCost: 20.00,
		}
		if err := DB.Create(&order4).Error; err == nil {
			orderItems4 := []models.OrderItem{
				{OrderID: order4.ID, ProductID: products[4].ID, Quantity: 1, Price: products[4].Price}, // Monitor
				{OrderID: order4.ID, ProductID: products[5].ID, Quantity: 1, Price: products[5].Price}, // Webcam
			}
			for _, item := range orderItems4 {
				DB.Create(&item)
			}
			log.Printf("Created order %d for user %s with %d items (Total: $%.2f)", order4.ID, users[2].Name, len(orderItems4), order4.TotalAmount)
		}

		// Create sample payments for orders
		payments := []models.Payment{
			{
				OrderID:       order1.ID,
				Amount:        order1.TotalAmount,
				Method:        models.PaymentMethodCreditCard,
				Status:        models.PaymentStatusCompleted,
				TransactionID: "TXN000001",
			},
			{
				OrderID:       order2.ID,
				Amount:        order2.TotalAmount,
				Method:        models.PaymentMethodPayPal,
				Status:        models.PaymentStatusCompleted,
				TransactionID: "TXN000002",
			},
			{
				OrderID:       order3.ID,
				Amount:        order3.TotalAmount,
				Method:        models.PaymentMethodDebitCard,
				Status:        models.PaymentStatusCompleted,
				TransactionID: "TXN000003",
			},
			{
				OrderID:       order4.ID,
				Amount:        order4.TotalAmount,
				Method:        models.PaymentMethodCreditCard,
				Status:        models.PaymentStatusPending,
				TransactionID: "TXN000004",
			},
		}

		for _, payment := range payments {
			if err := DB.Create(&payment).Error; err != nil {
				log.Printf("Warning: Failed to create payment for order %d: %v", payment.OrderID, err)
			} else {
				log.Printf("Created payment for order %d (Transaction: %s, Amount: $%.2f)", payment.OrderID, payment.TransactionID, payment.Amount)
			}
		}
	}

	log.Println("Database seeding completed successfully!")
	return nil
}
