package service

import (
	"context"
	"fmt"

	"github.com/artlink52/marketplace/internal/order/domain"
	"github.com/google/uuid"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order domain.Order) (domain.Order, bool, error)
	GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error)
}

type CatalogClient interface {
	BatchGetProducts(ctx context.Context, ids []uuid.UUID) ([]domain.ProductInfo, error)
}
type OrderService struct {
	repo          OrderRepository
	catalogClient CatalogClient
}

func New(repo OrderRepository, catalogClient CatalogClient) *OrderService {
	return &OrderService{
		repo:          repo,
		catalogClient: catalogClient,
	}
}
func (s *OrderService) CreateOrder(ctx context.Context, order domain.Order) (domain.Order, bool, error) {
	const op = "OrderService.CreateOrder"

	if len(order.Items) <= 0 {
		return domain.Order{}, false, domain.ErrEmptyOrder
	}

	for _, item := range order.Items {
		if item.Quantity <= 0 {
			return domain.Order{}, false, domain.ErrInvalidQuantity
		}
	}

	if order.UserID == uuid.Nil || order.IdempotencyKey == uuid.Nil {
		return domain.Order{}, false, domain.ErrInvalidInput
	}

	ids := make([]uuid.UUID, 0, len(order.Items))

	for _, item := range order.Items {
		ids = append(ids, item.ProductID)
	}

	productInfo, err := s.catalogClient.BatchGetProducts(ctx, ids)
	if err != nil {
		return domain.Order{}, false, fmt.Errorf("%s: batch get products: %w", op, err)
	}

	byID := make(map[uuid.UUID]domain.ProductInfo, len(productInfo))
	for _, p := range productInfo {
		byID[p.ID] = p
	}

	for i, item := range order.Items {
		p, ok := byID[item.ProductID]
		if !ok {
			return domain.Order{}, false, domain.ErrProductNotFound
		}
		if p.Status != "active" {
			return domain.Order{}, false, domain.ErrProductUnavailable
		}

		order.Items[i].UnitPrice = p.Price
	}

	var total int64
	for _, item := range order.Items {
		total += item.UnitPrice.Amount * int64(item.Quantity)
	}

	currency := productInfo[0].Price.Currency
	for _, p := range productInfo[1:] {
		if p.Price.Currency != currency {
			return domain.Order{}, false, domain.ErrCurrencyMismatch
		}
	}

	order.Total = domain.Money{
		Amount:   total,
		Currency: currency,
	}

	result, created, err := s.repo.CreateOrder(ctx, order)
	if err != nil {
		return domain.Order{}, false, fmt.Errorf("%s: %w", op, err)
	}

	return domain.Order{
		ID:     result.ID,
		Status: result.Status,
	}, created, nil
}
