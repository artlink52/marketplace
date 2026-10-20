package grpc

import (
	"context"
	"errors"

	orderv1 "github.com/artlink52/marketplace/gen/go/order/v1"
	"github.com/artlink52/marketplace/internal/order/domain"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderService interface {
	CreateOrder(ctx context.Context, order domain.Order) (domain.Order, bool, error)
	GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error)
}

type OrderServer struct {
	orderv1.UnimplementedOrderServiceServer
	orderService OrderService
}

func RegisterOrderServiceServer(server *grpc.Server, orderService OrderService) {
	orderv1.RegisterOrderServiceServer(server, &OrderServer{orderService: orderService})
}

func (s *OrderServer) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	idempotencyKey, err := uuid.Parse(req.GetIdempotencyKey())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid idempotency key")
	}

	items := make([]domain.OrderItem, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		productID, err := uuid.Parse(item.GetProductId())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid product id")
		}
		items = append(items, domain.OrderItem{
			ProductID: productID,
			Quantity:  item.GetQuantity(),
		})
	}

	order, _, err := s.orderService.CreateOrder(ctx, domain.Order{
		UserID:         userID,
		IdempotencyKey: idempotencyKey,
		Items:          items,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &orderv1.CreateOrderResponse{
		OrderId: order.ID.String(),
		Status:  toProtoStatus(order.Status),
	}, nil
}

func (s *OrderServer) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	orderID, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order id")
	}

	order, err := s.orderService.GetOrder(ctx, orderID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &orderv1.GetOrderResponse{
		Order: toProtoOrder(order),
	}, nil
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, domain.ErrOrderNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrEmptyOrder),
		errors.Is(err, domain.ErrInvalidQuantity),
		errors.Is(err, domain.ErrProductNotFound),
		errors.Is(err, domain.ErrProductUnavailable):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func toProtoStatus(s domain.OrderStatus) orderv1.OrderStatus {
	switch s {
	case domain.StatusPending:
		return orderv1.OrderStatus_ORDER_STATUS_PENDING
	case domain.StatusReserved:
		return orderv1.OrderStatus_ORDER_STATUS_RESERVED
	case domain.StatusConfirmed:
		return orderv1.OrderStatus_ORDER_STATUS_CONFIRMED
	case domain.StatusCancelled:
		return orderv1.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func toProtoOrder(o *domain.Order) *orderv1.Order {
	items := make([]*orderv1.OrderItem, 0, len(o.Items))
	for _, item := range o.Items {
		items = append(items, &orderv1.OrderItem{
			ProductId: item.ProductID.String(),
			Quantity:  item.Quantity,
		})
	}
	return &orderv1.Order{
		Id:     o.ID.String(),
		UserId: o.UserID.String(),
		Status: toProtoStatus(o.Status),
		Items:  items,
	}
}
