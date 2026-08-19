package models

import (
	"time"

	"gorm.io/gorm"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "Pending"
	OrderStatusConfirmed OrderStatus = "Confirmed"
	OrderStatusShipped   OrderStatus = "Shipped"
	OrderStatusDelivered OrderStatus = "Delivered"
	OrderStatusCancelled OrderStatus = "Cancelled"
)

// Order represents an order in the system
type Order struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	UserID       uint           `gorm:"not null;index" json:"user_id"`
	User         User           `gorm:"foreignKey:UserID" json:"user"`
	Items        []OrderItem    `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"items"`
	TotalAmount  float64        `gorm:"not null;type:decimal(10,2)" json:"total_amount"`
	Status       OrderStatus    `gorm:"type:varchar(20);default:'Pending'" json:"status"`
	ShippingCost float64        `gorm:"type:decimal(10,2);default:0" json:"shipping_cost"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	OrderID   uint           `gorm:"not null;index" json:"order_id"`
	ProductID uint           `gorm:"not null;index" json:"product_id"`
	Product   Product        `gorm:"foreignKey:ProductID" json:"product"`
	Quantity  int            `gorm:"not null;default:1" json:"quantity"`
	Price     float64        `gorm:"not null;type:decimal(10,2)" json:"price"` // Price at time of order
}

// TableName specifies the table name for Order
func (Order) TableName() string {
	return "orders"
}

// TableName specifies the table name for OrderItem
func (OrderItem) TableName() string {
	return "order_items"
}
