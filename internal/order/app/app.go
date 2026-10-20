package app

import (
	"context"
	"fmt"

	grpcapp "github.com/artlink52/marketplace/internal/order/app/grpc"
	"github.com/artlink52/marketplace/internal/order/client"
	"github.com/artlink52/marketplace/internal/order/config"
	"github.com/artlink52/marketplace/internal/order/repository"
	"github.com/artlink52/marketplace/internal/order/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	GrpcApp *grpcapp.App
	Pool    *pgxpool.Pool
}

func New(cfg config.Config) (*App, error) {
	const op = "order.New"
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	catalogClient, err := client.New(cfg.CatalogServiceAddr, cfg.CatalogServiceTimeout)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	orderRepo := repository.New(pool)
	orderService := service.New(orderRepo, catalogClient)
	grpcApp := grpcapp.New(orderService, cfg.Port, cfg.Timeout)

	return &App{
		GrpcApp: grpcApp,
		Pool:    pool,
	}, nil
}

func (a *App) Stop() {
	a.GrpcApp.Stop()
	a.Pool.Close()
}
