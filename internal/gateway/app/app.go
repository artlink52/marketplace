package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	userclient "github.com/artlink52/marketplace/internal/gateway/client/user"
	"github.com/artlink52/marketplace/internal/gateway/config"
	userhandler "github.com/artlink52/marketplace/internal/gateway/handler/user"
	"github.com/artlink52/marketplace/internal/gateway/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type App struct {
	server      *http.Server
	userClient  *userclient.Client
	httpTimeout time.Duration
}

func New(cfg config.Config) (*App, error) {
	userClient, err := userclient.New(cfg.UserServiceAddr)
	if err != nil {
		return nil, fmt.Errorf("create user client: %w", err)
	}

	userHandler := userhandler.New(
		userClient,
		cfg.JWTSecret,
		cfg.JWTTTL,
	)

	auth := middleware.Auth(cfg.JWTSecret)

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Mount("/", userhandler.NewRouter(userHandler, auth))

	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      r,
		ReadTimeout:  cfg.HTTPTimeout,
		WriteTimeout: cfg.HTTPTimeout,
	}

	return &App{
		server:      server,
		userClient:  userClient,
		httpTimeout: cfg.HTTPTimeout,
	}, nil
}

func (a *App) Run() error {
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server run error: %w", err)
	}
	return nil
}

func (a *App) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), a.httpTimeout)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}
	return a.userClient.Close()
}
