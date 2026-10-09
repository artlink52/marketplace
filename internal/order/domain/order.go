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

type Order struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Status         OrderStatus
	TotalAmount    int64
	Currency       string
	CancelReason   CancelReason
	IdempotencyKey uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type OrderItem struct {
	ID         uuid.UUID
	OrderID    uuid.UUID
	OrderItems []OrderItem
	ProductID  uuid.UUID
	Quantity   int64
	UnitPrice  int64
	Currency   string
}

type Money struct {
	Currency string
	Amount   int64
}

var (
	ErrOrderNotFound     = errors.New("order not found")
	ErrDuplicateOrder    = errors.New("order already exists")
	ErrInvalidTransition = errors.New("invalid status transition")
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
