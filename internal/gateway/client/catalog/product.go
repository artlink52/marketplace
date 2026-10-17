package catalog

import (
	"context"
	"fmt"

	catalogv1 "github.com/artlink52/marketplace/gen/go/catalog/v1"
	commonv1 "github.com/artlink52/marketplace/gen/go/common/v1"
	"github.com/google/uuid"
)

type CreateProductInput struct {
	CategoryID  string  `json:"category_id"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Amount      int64   `json:"amount"`
	Currency    string  `json:"currency"`
}

func (c *Client) CreateProduct(ctx context.Context, input CreateProductInput, sellerID string) (uuid.UUID, error) {
	resp, err := c.catalog.CreateProduct(ctx, &catalogv1.CreateProductRequest{
		SellerId:    sellerID,
		CategoryId:  input.CategoryID,
		Title:       input.Title,
		Description: *input.Description,
		Price: &commonv1.Money{
			Currency: input.Currency,
			Amount:   input.Amount,
		},
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("error creating product: %w", err)
	}

	id, err := uuid.Parse(resp.ProductId)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error parsing product id: %w", err)
	}

	return id, nil
}
