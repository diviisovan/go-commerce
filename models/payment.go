package models

import (
	"time"

	"gorm.io/gorm"
)

// PaymentMethod represents different payment methods
type PaymentMethod string

const (
	PaymentMethodCreditCard PaymentMethod = "Credit Card"
	PaymentMethodDebitCard  PaymentMethod = "Debit Card"
	PaymentMethodPayPal     PaymentMethod = "PayPal"
	PaymentMethodCash       PaymentMethod = "Cash on Delivery"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "Pending"
	PaymentStatusCompleted PaymentStatus = "Completed"
	PaymentStatusFailed    PaymentStatus = "Failed"
	PaymentStatusRefunded  PaymentStatus = "Refunded"
)

// Payment represents a payment transaction
type Payment struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	OrderID       uint           `gorm:"not null;index" json:"order_id"`
	Order         Order          `gorm:"foreignKey:OrderID" json:"order"`
	Amount        float64        `gorm:"not null;type:decimal(10,2)" json:"amount"`
	Method        PaymentMethod  `gorm:"type:varchar(50);not null" json:"method"`
	Status        PaymentStatus  `gorm:"type:varchar(20);default:'Pending'" json:"status"`
	TransactionID string         `gorm:"type:varchar(50);uniqueIndex" json:"transaction_id"`
}

// TableName specifies the table name for Payment
func (Payment) TableName() string {
	return "payments"
}
