package catalog

import (
	"context"

	"github.com/artlink52/marketplace/internal/gateway/client/catalog"
)

type CategoryClient interface {
	CreateCategory(ctx context.Context, input catalog.CreateCategoryInput) (*catalog.Category, error)
}

type Handler struct {
	categoryClient CategoryClient
}

func New(categoryClient CategoryClient) *Handler {
	return &Handler{
		categoryClient: categoryClient,
	}
}
