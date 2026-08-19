package product

import "fmt"

// Product represents a product in the eCommerce system
type Product struct {
	ID          int
	Name        string
	Description string
	Price       float64
	Stock       int
	Category    string
}

// Catalog manages the product inventory
type Catalog struct {
	Products map[int]*Product
}

// NewCatalog creates a new product catalog
func NewCatalog() *Catalog {
	return &Catalog{
		Products: make(map[int]*Product),
	}
}

// AddProduct adds a new product to the catalog
func (pc *Catalog) AddProduct(product *Product) {
	pc.Products[product.ID] = product
}

// GetProduct retrieves a product by ID
func (pc *Catalog) GetProduct(id int) (*Product, bool) {
	product, exists := pc.Products[id]
	return product, exists
}

// UpdateStock updates the stock quantity of a product
func (pc *Catalog) UpdateStock(productID int, quantity int) error {
	product, exists := pc.Products[productID]
	if !exists {
		return fmt.Errorf("product with ID %d not found", productID)
	}

	if product.Stock+quantity < 0 {
		return fmt.Errorf("insufficient stock for product %s. Available: %d, Requested: %d",
			product.Name, product.Stock, -quantity)
	}

	product.Stock += quantity
	return nil
}

// IsInStock checks if a product has sufficient stock
func (pc *Catalog) IsInStock(productID int, quantity int) bool {
	product, exists := pc.Products[productID]
	if !exists {
		return false
	}
	return product.Stock >= quantity
}

// ListProducts displays all products in the catalog
func (pc *Catalog) ListProducts() {
	fmt.Println("\n=== Product Catalog ===")
	for _, product := range pc.Products {
		fmt.Printf("ID: %d | %s | $%.2f | Stock: %d | Category: %s\n",
			product.ID, product.Name, product.Price, product.Stock, product.Category)
	}
}

// GetProductDetails displays detailed information about a product
func (p *Product) GetProductDetails() {
	fmt.Printf("\n=== Product Details ===\n")
	fmt.Printf("ID:          %d\n", p.ID)
	fmt.Printf("Name:        %s\n", p.Name)
	fmt.Printf("Description: %s\n", p.Description)
	fmt.Printf("Price:       $%.2f\n", p.Price)
	fmt.Printf("Stock:       %d\n", p.Stock)
	fmt.Printf("Category:    %s\n", p.Category)
}
