package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusReserved  OrderStatus = "reserved"
	StatusConfirmed OrderStatus = "confirmed"
	StatusCancelled OrderStatus = "cancelled"
)

type CancelReason string

const (
	ReasonOutOfStock    CancelReason = "out_of_stock"
	ReasonPaymentFailed CancelReason = "payment_failed"
	ReasonTimeout       CancelReason = "timeout"
)

type ProductInfo struct {
	ID     uuid.UUID
	Price  Money
	Status string
}

type Order struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Status         OrderStatus
	Total          Money
	CancelReason   *CancelReason
	IdempotencyKey uuid.UUID
	Items          []OrderItem
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type OrderItem struct {
	ID        uuid.UUID
	OrderID   uuid.UUID
	ProductID uuid.UUID
	Quantity  int32
	UnitPrice Money
}

type Money struct {
	Currency string
	Amount   int64
}

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrEmptyOrder         = errors.New("empty order")
	ErrInvalidQuantity    = errors.New("invalid quantity")
	ErrProductNotFound    = errors.New("product not found")
	ErrProductUnavailable = errors.New("product unavailable")
	ErrInvalidTransition  = errors.New("invalid status transition")

	ErrInvalidInput     = errors.New("invalid input")
	ErrCurrencyMismatch = errors.New("currency mismatch")
)

var allowedTransitions = map[OrderStatus][]OrderStatus{
	StatusPending:  {StatusReserved, StatusCancelled},
	StatusReserved: {StatusConfirmed, StatusCancelled},
}

func (o *Order) Transition(to OrderStatus) error {
	for _, v := range allowedTransitions[o.Status] {
		if v == to {
			o.Status = to
			return nil
		}
	}
	return ErrInvalidTransition
}
