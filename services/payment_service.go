package services

import (
	"errors"
	"fmt"
	"go-ecommerce/database"
	"go-ecommerce/models"
)

// PaymentService handles payment-related business logic
type PaymentService struct{}

// NewPaymentService creates a new payment service
func NewPaymentService() *PaymentService {
	return &PaymentService{}
}

// ProcessPayment processes a payment for an order
func (s *PaymentService) ProcessPayment(orderID uint, amount float64, method models.PaymentMethod) (*models.Payment, error) {
	if amount <= 0 {
		return nil, errors.New("invalid payment amount")
	}

	// Generate transaction ID
	var count int64
	database.GetDB().Model(&models.Payment{}).Count(&count)
	transactionID := fmt.Sprintf("TXN%06d", count+1)

	payment := &models.Payment{
		OrderID:       orderID,
		Amount:        amount,
		Method:        method,
		Status:        models.PaymentStatusPending,
		TransactionID: transactionID,
	}

	// Simulate payment processing (in real system, this would call payment gateway)
	payment.Status = models.PaymentStatusCompleted

	if err := database.GetDB().Create(payment).Error; err != nil {
		return nil, err
	}

	// Update order status to confirmed
	if err := database.GetDB().Model(&models.Order{}).Where("id = ?", orderID).Update("status", models.OrderStatusConfirmed).Error; err != nil {
		return nil, err
	}

	return payment, nil
}

// GetPayment retrieves a payment by ID
func (s *PaymentService) GetPayment(id uint) (*models.Payment, error) {
	var payment models.Payment
	err := database.GetDB().Preload("Order").First(&payment, id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// RefundPayment processes a refund for a payment
func (s *PaymentService) RefundPayment(paymentID uint) error {
	var payment models.Payment
	if err := database.GetDB().First(&payment, paymentID).Error; err != nil {
		return err
	}

	if payment.Status != models.PaymentStatusCompleted {
		return errors.New("can only refund completed payments")
	}

	return database.GetDB().Model(&payment).Update("status", models.PaymentStatusRefunded).Error
}
