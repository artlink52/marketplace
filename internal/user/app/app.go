package app

import (
	"context"
	"fmt"

	grpcapp "github.com/artlink52/marketplace/internal/user/app/grpc"
	"github.com/artlink52/marketplace/internal/user/config"
	"github.com/artlink52/marketplace/internal/user/repository"
	dbuser "github.com/artlink52/marketplace/internal/user/repository/sqlc"
	"github.com/artlink52/marketplace/internal/user/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	GRPCServer *grpcapp.App
	Pool       *pgxpool.Pool
}

func New(cfg config.Config) (*App, error) {
	const op = "app.New"

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	repo := repository.New(dbuser.New(pool))
	userService := service.New(repo)
	grpcApp := grpcapp.New(userService, cfg.GRPC.Port, cfg.GRPC.Timeout)

	return &App{
		GRPCServer: grpcApp,
		Pool:       pool,
	}, nil
}

func (a *App) Stop() {
	a.GRPCServer.Stop()
	a.Pool.Close()
}
