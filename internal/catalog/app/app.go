package app

import (
	"context"
	"fmt"

	grpcapp "github.com/artlink52/marketplace/internal/catalog/app/grpc"
	"github.com/artlink52/marketplace/internal/catalog/config"
	"github.com/artlink52/marketplace/internal/catalog/repository"
	dbcatalog "github.com/artlink52/marketplace/internal/catalog/repository/sqlc"
	"github.com/artlink52/marketplace/internal/catalog/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	GrpcApp *grpcapp.App
	Pool    *pgxpool.Pool
}

func New(cfg config.Config) (*App, error) {
	const op = "catalog.New"

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	catalogRepo := repository.New(dbcatalog.New(pool))
	userService := service.New(catalogRepo)
	grpcApp := grpcapp.New(userService, cfg.Port, cfg.Timeout)

	return &App{
		GrpcApp: grpcApp,
		Pool:    pool,
	}, nil
}

func (a *App) Stop() {
	a.GrpcApp.Stop()
	a.Pool.Close()
}
