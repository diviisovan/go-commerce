package models

import (
	"time"

	"gorm.io/gorm"
)

// Product represents a product in the eCommerce system
// @Description Product information
type Product struct {
	ID          uint           `gorm:"primaryKey" json:"id" example:"1"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name" example:"Laptop"`
	Description string         `gorm:"type:text" json:"description" example:"High-performance laptop"`
	Price       float64        `gorm:"not null;type:decimal(10,2)" json:"price" example:"999.99"`
	Stock       int            `gorm:"not null;default:0" json:"stock" example:"10"`
	Category    string         `gorm:"type:varchar(100);not null" json:"category" example:"Electronics"`
}

// TableName specifies the table name for Product
func (Product) TableName() string {
	return "products"
}
