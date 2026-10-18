package repository

import (
	"context"
	"errors"
	"fmt"

	commonv1 "github.com/artlink52/marketplace/gen/go/common/v1"
	orderv1 "github.com/artlink52/marketplace/gen/go/order/v1"
	"github.com/artlink52/marketplace/internal/order/domain"
	dborder "github.com/artlink52/marketplace/internal/order/repository/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/proto"
)

type OrderRepository struct {
	q    *dborder.Queries
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{
		q:    dborder.New(pool),
		pool: pool,
	}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order domain.Order) (domain.Order, bool, error) {
	const op = "OrderRepository.CreateOrder"

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Order{}, false, fmt.Errorf("%s: begin tx: %w", op, err)
	}
	defer tx.Rollback(ctx)

	qtx := r.q.WithTx(tx)

	row, err := qtx.CreateOrder(ctx, dborder.CreateOrderParams{
		UserID:         order.UserID,
		TotalAmount:    order.Total.Amount,
		IdempotencyKey: order.IdempotencyKey,
	})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, false, fmt.Errorf("%s: insert order: %w", op, err)
		}
		existing, err := qtx.GetOrderByIdempotencyKey(ctx, dborder.GetOrderByIdempotencyKeyParams{
			IdempotencyKey: order.IdempotencyKey,
			UserID:         order.UserID,
		})
		if err != nil {
			return domain.Order{}, false, fmt.Errorf("%s: get existing order: %w", op, err)
		}
		order.ID = existing.ID
		order.Status = domain.OrderStatus(existing.Status)
		return order, false, nil
	}

	order.ID = row.ID
	order.Status = domain.OrderStatus(row.Status)

	items := make([]dborder.CreateOrderItemsParams, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, dborder.CreateOrderItemsParams{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice.Amount,
		})
	}

	if _, err := qtx.CreateOrderItems(ctx, items); err != nil {
		return domain.Order{}, false, fmt.Errorf("%s: insert items: %w", op, err)
	}

	payload, err := proto.Marshal(toOrderCreatedEvent(order))
	if err != nil {
		return domain.Order{}, false, fmt.Errorf("%s: marshal event: %w", op, err)
	}

	if err := qtx.InsertOutboxEvent(ctx, dborder.InsertOutboxEventParams{
		EventType:   "order.created",
		AggregateID: order.ID,
		Payload:     payload,
	}); err != nil {
		return domain.Order{}, false, fmt.Errorf("%s: insert outbox: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Order{}, false, fmt.Errorf("%s: commit tx: %w", op, err)
	}

	return order, true, nil
}

func toOrderCreatedEvent(order domain.Order) *orderv1.OrderCreated {
	items := make([]*orderv1.OrderItem, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, &orderv1.OrderItem{
			ProductId: item.ProductID.String(),
			Quantity:  item.Quantity,
			UnitPrice: &commonv1.Money{
				Amount:   item.UnitPrice.Amount,
				Currency: item.UnitPrice.Currency,
			},
		})
	}
	return &orderv1.OrderCreated{
		OrderId: order.ID.String(),
		UserId:  order.UserID.String(),
		Items:   items,
		TotalAmount: &commonv1.Money{
			Amount:   order.Total.Amount,
			Currency: order.Total.Currency,
		},
	}
}
