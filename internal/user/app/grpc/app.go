package grpcapp

import (
	"context"
	"fmt"
	"net"
	"time"

	usergrpc "github.com/artlink52/marketplace/internal/user/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	port       string
	gRPCServer *grpc.Server
}

func New(
	userService usergrpc.UserService,
	port string,
	timeout time.Duration,
) *App {
	gRPCServer := grpc.NewServer(grpc.UnaryInterceptor(timeoutInterceptor(timeout)))
	usergrpc.RegisterUserServer(gRPCServer, userService)
	reflection.Register(gRPCServer)
	return &App{gRPCServer: gRPCServer, port: port}
}

func timeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return handler(ctx, req)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	lis, err := net.Listen("tcp", ":"+a.port)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return a.gRPCServer.Serve(lis)
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Stop() {
	a.gRPCServer.GracefulStop()
}
