package cart

import (
	"fmt"
	"go-ecommerce/internal/product"
)

// Item represents an item in the shopping cart
type Item struct {
	Product  *product.Product
	Quantity int
}

// ShoppingCart represents a user's shopping cart
type ShoppingCart struct {
	UserID int
	Items  []*Item
}

// NewShoppingCart creates a new shopping cart for a user
func NewShoppingCart(userID int) *ShoppingCart {
	return &ShoppingCart{
		UserID: userID,
		Items:  make([]*Item, 0),
	}
}

// AddItem adds a product to the cart
func (sc *ShoppingCart) AddItem(prod *product.Product, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}

	if prod.Stock < quantity {
		return fmt.Errorf("insufficient stock. Available: %d, Requested: %d", prod.Stock, quantity)
	}

	// Check if item already exists in cart
	for _, item := range sc.Items {
		if item.Product.ID == prod.ID {
			item.Quantity += quantity
			return nil
		}
	}

	// Add new item to cart
	sc.Items = append(sc.Items, &Item{
		Product:  prod,
		Quantity: quantity,
	})

	return nil
}

// RemoveItem removes a product from the cart
func (sc *ShoppingCart) RemoveItem(productID int) error {
	for i, item := range sc.Items {
		if item.Product.ID == productID {
			sc.Items = append(sc.Items[:i], sc.Items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("product with ID %d not found in cart", productID)
}

// UpdateQuantity updates the quantity of an item in the cart
func (sc *ShoppingCart) UpdateQuantity(productID int, quantity int) error {
	if quantity <= 0 {
		return sc.RemoveItem(productID)
	}

	for _, item := range sc.Items {
		if item.Product.ID == productID {
			if item.Product.Stock < quantity {
				return fmt.Errorf("insufficient stock. Available: %d, Requested: %d",
					item.Product.Stock, quantity)
			}
			item.Quantity = quantity
			return nil
		}
	}
	return fmt.Errorf("product with ID %d not found in cart", productID)
}

// GetTotal calculates the total price of all items in the cart
func (sc *ShoppingCart) GetTotal() float64 {
	total := 0.0
	for _, item := range sc.Items {
		total += item.Product.Price * float64(item.Quantity)
	}
	return total
}

// GetItemCount returns the total number of items in the cart
func (sc *ShoppingCart) GetItemCount() int {
	count := 0
	for _, item := range sc.Items {
		count += item.Quantity
	}
	return count
}

// Clear empties the shopping cart
func (sc *ShoppingCart) Clear() {
	sc.Items = make([]*Item, 0)
}

// DisplayCart displays all items in the cart
func (sc *ShoppingCart) DisplayCart() {
	fmt.Println("\n=== Shopping Cart ===")
	if len(sc.Items) == 0 {
		fmt.Println("Your cart is empty.")
		return
	}

	for _, item := range sc.Items {
		itemTotal := item.Product.Price * float64(item.Quantity)
		fmt.Printf("ID: %d | %s | $%.2f x %d = $%.2f\n",
			item.Product.ID, item.Product.Name, item.Product.Price, item.Quantity, itemTotal)
	}
	fmt.Printf("\nTotal Items: %d\n", sc.GetItemCount())
	fmt.Printf("Total Price: $%.2f\n", sc.GetTotal())
}
