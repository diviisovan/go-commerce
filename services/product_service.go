package services

import (
	"errors"
	"go-ecommerce/database"
	"go-ecommerce/models"
)

// ProductService handles product-related business logic
type ProductService struct{}

// NewProductService creates a new product service
func NewProductService() *ProductService {
	return &ProductService{}
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(product *models.Product) error {
	return database.GetDB().Create(product).Error
}

// GetProduct retrieves a product by ID
func (s *ProductService) GetProduct(id uint) (*models.Product, error) {
	var product models.Product
	err := database.GetDB().First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetAllProducts retrieves all products
func (s *ProductService) GetAllProducts() ([]models.Product, error) {
	var products []models.Product
	err := database.GetDB().Find(&products).Error
	return products, err
}

// UpdateProduct updates a product
func (s *ProductService) UpdateProduct(id uint, product *models.Product) error {
	return database.GetDB().Model(&models.Product{}).Where("id = ?", id).Updates(product).Error
}

// DeleteProduct deletes a product
func (s *ProductService) DeleteProduct(id uint) error {
	return database.GetDB().Delete(&models.Product{}, id).Error
}

// UpdateStock updates the stock quantity of a product
func (s *ProductService) UpdateStock(productID uint, quantity int) error {
	var product models.Product
	if err := database.GetDB().First(&product, productID).Error; err != nil {
		return err
	}

	newStock := product.Stock + quantity
	if newStock < 0 {
		return errors.New("insufficient stock")
	}

	return database.GetDB().Model(&product).Update("stock", newStock).Error
}

// IsInStock checks if a product has sufficient stock
func (s *ProductService) IsInStock(productID uint, quantity int) bool {
	var product models.Product
	if err := database.GetDB().First(&product, productID).Error; err != nil {
		return false
	}
	return product.Stock >= quantity
}
