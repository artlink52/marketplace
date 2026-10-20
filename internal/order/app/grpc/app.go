package grpcapp

import (
	"context"
	"fmt"
	"net"
	"time"

	ordergrpc "github.com/artlink52/marketplace/internal/order/grpc"
	"google.golang.org/grpc"
)

type App struct {
	port       string
	grpcServer *grpc.Server
}

func New(
	orderService ordergrpc.OrderService,
	port string,
	timeout time.Duration,
) *App {
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(timeoutInterceptor(timeout)))
	ordergrpc.RegisterOrderServiceServer(grpcServer, orderService)
	return &App{
		port:       port,
		grpcServer: grpcServer,
	}
}

func (a *App) Run() error {
	const op = "orderApp.Run"
	lis, err := net.Listen("tcp", ":"+a.port)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return a.grpcServer.Serve(lis)
}

func (a *App) Stop() {
	a.grpcServer.GracefulStop()
}

func timeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return handler(ctx, req)
	}
}
