package payment

import "fmt"

// Method represents different payment methods
type Method string

const (
	MethodCreditCard Method = "Credit Card"
	MethodDebitCard  Method = "Debit Card"
	MethodPayPal     Method = "PayPal"
	MethodCash       Method = "Cash on Delivery"
)

// Status represents the status of a payment
type Status string

const (
	StatusPending   Status = "Pending"
	StatusCompleted Status = "Completed"
	StatusFailed    Status = "Failed"
	StatusRefunded  Status = "Refunded"
)

// Payment represents a payment transaction
type Payment struct {
	ID            int
	OrderID       int
	Amount        float64
	Method        Method
	Status        Status
	TransactionID string
}

// Processor handles payment processing
type Processor struct {
	Payments map[int]*Payment
	NextID   int
}

// NewProcessor creates a new payment processor
func NewProcessor() *Processor {
	return &Processor{
		Payments: make(map[int]*Payment),
		NextID:   1,
	}
}

// ProcessPayment processes a payment for an order
func (pp *Processor) ProcessPayment(orderID int, amount float64, method Method) (*Payment, error) {
	// Simulate payment processing
	payment := &Payment{
		ID:            pp.NextID,
		OrderID:       orderID,
		Amount:        amount,
		Method:        method,
		Status:        StatusPending,
		TransactionID: fmt.Sprintf("TXN%06d", pp.NextID),
	}

	// Simulate payment validation (in real system, this would call payment gateway)
	if amount <= 0 {
		payment.Status = StatusFailed
		return payment, fmt.Errorf("invalid payment amount")
	}

	// Simulate successful payment (90% success rate simulation)
	// In real system, this would be handled by payment gateway
	payment.Status = StatusCompleted

	pp.Payments[pp.NextID] = payment
	pp.NextID++

	return payment, nil
}

// GetPayment retrieves a payment by ID
func (pp *Processor) GetPayment(id int) (*Payment, bool) {
	payment, exists := pp.Payments[id]
	return payment, exists
}

// RefundPayment processes a refund for a payment
func (pp *Processor) RefundPayment(paymentID int) error {
	payment, exists := pp.Payments[paymentID]
	if !exists {
		return fmt.Errorf("payment with ID %d not found", paymentID)
	}

	if payment.Status != StatusCompleted {
		return fmt.Errorf("can only refund completed payments")
	}

	payment.Status = StatusRefunded
	return nil
}

// DisplayPayment displays payment details
func (p *Payment) DisplayPayment() {
	fmt.Printf("\n=== Payment Details ===\n")
	fmt.Printf("Payment ID:     %d\n", p.ID)
	fmt.Printf("Order ID:       %d\n", p.OrderID)
	fmt.Printf("Amount:         $%.2f\n", p.Amount)
	fmt.Printf("Method:         %s\n", p.Method)
	fmt.Printf("Status:         %s\n", p.Status)
	fmt.Printf("Transaction ID: %s\n", p.TransactionID)
}
