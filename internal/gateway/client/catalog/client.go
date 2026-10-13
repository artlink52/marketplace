package catalog

import (
	"fmt"

	catalogv1 "github.com/artlink52/marketplace/gen/go/catalog/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	catalog catalogv1.CatalogServiceClient
	conn    *grpc.ClientConn
}

func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connect to catalog-service: %w", err)
	}

	return &Client{
		catalog: catalogv1.NewCatalogServiceClient(conn),
		conn:    conn,
	}, nil
}

func (c *Client) Close() error {
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("close catalog-service: %w", err)
	}
	return nil
}
