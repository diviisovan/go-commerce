package models

import (
	"time"

	"gorm.io/gorm"
)

// Cart represents a user's shopping cart
type Cart struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Items     []CartItem     `gorm:"foreignKey:CartID;constraint:OnDelete:CASCADE" json:"items"`
}

// CartItem represents an item in the shopping cart
type CartItem struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	CartID    uint           `gorm:"not null;index" json:"cart_id"`
	ProductID uint           `gorm:"not null;index" json:"product_id"`
	Product   Product        `gorm:"foreignKey:ProductID" json:"product"`
	Quantity  int            `gorm:"not null;default:1" json:"quantity"`
}

// TableName specifies the table name for Cart
func (Cart) TableName() string {
	return "carts"
}

// TableName specifies the table name for CartItem
func (CartItem) TableName() string {
	return "cart_items"
}
