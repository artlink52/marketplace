package client

import (
	"context"
	"fmt"
	"time"

	catalogv1 "github.com/artlink52/marketplace/gen/go/catalog/v1"
	"github.com/artlink52/marketplace/internal/order/domain"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CatalogClient struct {
	client catalogv1.CatalogServiceClient
	conn   *grpc.ClientConn
}

func New(addr string, timeout time.Duration) (*CatalogClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(timeoutInterceptor(timeout)),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to catalog-service: %w", err)
	}

	return &CatalogClient{
		client: catalogv1.NewCatalogServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *CatalogClient) Close() error {
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("close catalog-service: %w", err)
	}
	return nil
}

func (c *CatalogClient) BatchGetProducts(ctx context.Context, ids []uuid.UUID) ([]domain.ProductInfo, error) {
	const op = "CatalogClient.BatchGetProducts"

	strIDs := make([]string, len(ids))
	for i, id := range ids {
		strIDs[i] = id.String()
	}

	resp, err := c.client.BatchGetProducts(ctx, &catalogv1.BatchGetProductsRequest{
		ProductIds: strIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	result := make([]domain.ProductInfo, 0, len(resp.Products))
	for _, p := range resp.Products {
		id, err := uuid.Parse(p.Id)
		if err != nil {
			return nil, fmt.Errorf("%s: parse product id: %w", op, err)
		}
		result = append(result, domain.ProductInfo{
			ID: id,
			Price: domain.Money{
				Amount:   p.Price.Amount,
				Currency: p.Price.Currency,
			},
			Status: toProductStatus(p.Status),
		})
	}

	return result, nil
}

func toProductStatus(s catalogv1.ProductStatus) string {
	switch s {
	case catalogv1.ProductStatus_PRODUCT_STATUS_ACTIVE:
		return "active"
	case catalogv1.ProductStatus_PRODUCT_STATUS_DRAFT:
		return "draft"
	case catalogv1.ProductStatus_PRODUCT_STATUS_ARCHIVED:
		return "archived"
	default:
		return "unspecified"
	}
}

func timeoutInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req any,
		reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
