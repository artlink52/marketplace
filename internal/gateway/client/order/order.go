package orderclient

import (
	"context"
	"fmt"

	orderv1 "github.com/artlink52/marketplace/gen/go/order/v1"
	"github.com/google/uuid"
)

type CreateOrderInput struct {
	UserID         uuid.UUID
	Items          []OrderItem
	IdempotencyKey string
}

type OrderItem struct {
	ProductID uuid.UUID
	Quantity  int32
}

type CreateOrderResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

func (c *Client) CreateOrder(ctx context.Context, input CreateOrderInput) (*CreateOrderResponse, error) {
	resp, err := c.client.CreateOrder(ctx, &orderv1.CreateOrderRequest{
		UserId:         input.UserID.String(),
		Items:          toProtoItems(input.Items),
		IdempotencyKey: input.IdempotencyKey,
	})
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	return &CreateOrderResponse{
		OrderID: resp.OrderId,
		Status:  fromProtoStatus(resp.Status),
	}, nil
}

func fromProtoStatus(s orderv1.OrderStatus) string {
	switch s {
	case orderv1.OrderStatus_ORDER_STATUS_PENDING:
		return "pending"
	case orderv1.OrderStatus_ORDER_STATUS_RESERVED:
		return "reserved"
	case orderv1.OrderStatus_ORDER_STATUS_CONFIRMED:
		return "confirmed"
	case orderv1.OrderStatus_ORDER_STATUS_CANCELLED:
		return "cancelled"
	default:
		return "unspecified"
	}
}

func toProtoItems(items []OrderItem) []*orderv1.RequestedItem {
	protoItems := make([]*orderv1.RequestedItem, len(items))
	for i, item := range items {
		protoItems[i] = &orderv1.RequestedItem{
			ProductId: item.ProductID.String(),
			Quantity:  item.Quantity,
		}
	}
	return protoItems
}
