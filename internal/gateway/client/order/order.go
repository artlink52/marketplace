package order

import (
	"context"
	"fmt"

	orderv1 "github.com/artlink52/marketplace/gen/go/order/v1"
	"github.com/artlink52/marketplace/internal/order/domain"
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

func (c *Client) CreateOrder(ctx context.Context, input CreateOrderInput) (domain.Order, error) {
	resp, err := c.client.CreateOrder(ctx, &orderv1.CreateOrderRequest{
		UserId:         input.UserID.String(),
		Items:          toProtoItems(input.Items),
		IdempotencyKey: input.IdempotencyKey,
	})
	if err != nil {
		return domain.Order{}, fmt.Errorf("create order: %w", err)
	}

	id, err := uuid.Parse(resp.OrderId)
	if err != nil {
		return domain.Order{}, fmt.Errorf("parse order id: %w", err)
	}

	return domain.Order{
		ID:     id,
		Status: domain.OrderStatus(resp.Status),
	}, nil
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
