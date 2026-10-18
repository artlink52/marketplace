package catalog

import (
	"context"
	"fmt"
	"time"

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
		Description: input.Description,
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

type Money struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type Product struct {
	ID          string    `json:"id"`
	SellerID    string    `json:"seller_id"`
	CategoryID  string    `json:"category_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	Price       Money     `json:"price"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ListProductsInput struct {
	CategoryID string
	SellerID   string
	PageSize   int32
	PageToken  string
}

type ProductsPage struct {
	Products      []Product `json:"products"`
	NextPageToken string    `json:"next_page_token,omitempty"`
}

func (c *Client) GetProduct(ctx context.Context, id string) (*Product, error) {
	resp, err := c.catalog.GetProduct(ctx, &catalogv1.GetProductRequest{ProductId: id})
	if err != nil {
		return nil, fmt.Errorf("get product: %w", err)
	}

	p := fromProtoProduct(resp.GetProduct())
	return &p, nil
}

func (c *Client) ListProducts(ctx context.Context, input ListProductsInput) (*ProductsPage, error) {
	resp, err := c.catalog.ListProducts(ctx, &catalogv1.ListProductsRequest{
		CategoryId: input.CategoryID,
		SellerId:   input.SellerID,
		Page: &commonv1.PageRequest{
			PageSize:  input.PageSize,
			PageToken: input.PageToken,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}

	page := &ProductsPage{
		Products:      make([]Product, 0, len(resp.GetProducts())),
		NextPageToken: resp.GetPage().GetNextPageToken(),
	}
	for _, p := range resp.GetProducts() {
		page.Products = append(page.Products, fromProtoProduct(p))
	}
	return page, nil
}

func fromProtoProduct(p *catalogv1.Product) Product {
	return Product{
		ID:          p.GetId(),
		SellerID:    p.GetSellerId(),
		CategoryID:  p.GetCategoryId(),
		Title:       p.GetTitle(),
		Description: p.Description,
		Price: Money{
			Amount:   p.GetPrice().GetAmount(),
			Currency: p.GetPrice().GetCurrency(),
		},
		Status:    fromProtoProductStatus(p.GetStatus()),
		CreatedAt: p.GetCreatedAt().AsTime(),
		UpdatedAt: p.GetUpdatedAt().AsTime(),
	}
}

func fromProtoProductStatus(s catalogv1.ProductStatus) string {
	switch s {
	case catalogv1.ProductStatus_PRODUCT_STATUS_DRAFT:
		return "draft"
	case catalogv1.ProductStatus_PRODUCT_STATUS_ACTIVE:
		return "active"
	case catalogv1.ProductStatus_PRODUCT_STATUS_ARCHIVED:
		return "archived"
	default:
		return "unknown"
	}
}
