package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	cataloggrpc "github.com/artlink52/marketplace/internal/catalog/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	port       string
	gRPCServer *grpc.Server
}

func New(
	catalogService cataloggrpc.CatalogService,
	port string,
	timeout time.Duration,
) *App {
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(timeoutInterceptor(timeout)))
	cataloggrpc.RegisterCatalogServer(grpcServer, catalogService)
	reflection.Register(grpcServer)
	return &App{
		port:       port,
		gRPCServer: grpcServer,
	}
}

func (a *App) Run() error {
	const op = "cataloggrpc.Run"
	lis, err := net.Listen("tcp", ":"+a.port)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return a.gRPCServer.Serve(lis)
}

func (a *App) Stop() {
	a.gRPCServer.GracefulStop()
}

func timeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return handler(ctx, req)
	}
}
