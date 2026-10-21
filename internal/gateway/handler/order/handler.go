package order

import (
	"context"

	orderclient "github.com/artlink52/marketplace/internal/gateway/client/order"
)

type Client interface {
	CreateOrder(ctx context.Context, input orderclient.CreateOrderInput) (*orderclient.CreateOrderResponse, error)
}

type Handler struct {
	client Client
}

func New(client Client) *Handler {
	return &Handler{
		client: client,
	}
}
