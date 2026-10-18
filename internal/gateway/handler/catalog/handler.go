package catalog

import (
	"context"

	"github.com/artlink52/marketplace/internal/gateway/client/catalog"
	"github.com/google/uuid"
)

type CategoryClient interface {
	CreateCategory(ctx context.Context, input catalog.CreateCategoryInput) (*catalog.Category, error)
	ListCategories(ctx context.Context) ([]catalog.Category, error)
	CreateProduct(ctx context.Context, input catalog.CreateProductInput, sellerID string) (uuid.UUID, error)
	GetProduct(ctx context.Context, id string) (*catalog.Product, error)
	ListProducts(ctx context.Context, input catalog.ListProductsInput) (*catalog.ProductsPage, error)
}

type Handler struct {
	categoryClient CategoryClient
}

func New(categoryClient CategoryClient) *Handler {
	return &Handler{
		categoryClient: categoryClient,
	}
}
