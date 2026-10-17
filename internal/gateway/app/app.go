package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	catalogclient "github.com/artlink52/marketplace/internal/gateway/client/catalog"
	userclient "github.com/artlink52/marketplace/internal/gateway/client/user"
	"github.com/artlink52/marketplace/internal/gateway/config"
	"github.com/artlink52/marketplace/internal/gateway/handler"
	cataloghandler "github.com/artlink52/marketplace/internal/gateway/handler/catalog"
	userhandler "github.com/artlink52/marketplace/internal/gateway/handler/user"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type App struct {
	server        *http.Server
	userClient    *userclient.Client
	catalogClient *catalogclient.Client
	httpTimeout   time.Duration
}

func New(cfg config.Config) (*App, error) {
	userClient, err := userclient.New(cfg.UserServiceAddr, cfg.UserServiceTimeout)
	if err != nil {
		return nil, fmt.Errorf("create user client: %w", err)
	}

	catalogClient, err := catalogclient.New(cfg.CatalogServiceAddr, cfg.CatalogServiceTimeout)
	if err != nil {
		_ = userClient.Close()
		return nil, fmt.Errorf("create catalog client: %w", err)
	}

	handlers := handler.Handlers{
		User:    userhandler.New(userClient, cfg.JWTSecret, cfg.JWTTTL),
		Catalog: cataloghandler.New(catalogClient),
	}

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Mount("/", handler.NewRouter(handlers, cfg))

	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      r,
		ReadTimeout:  cfg.HTTPTimeout,
		WriteTimeout: cfg.HTTPTimeout,
	}

	return &App{
		server:        server,
		userClient:    userClient,
		catalogClient: catalogClient,
		httpTimeout:   cfg.HTTPTimeout,
	}, nil
}

func (a *App) Run() error {
	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server run error: %w", err)
	}
	return nil
}

func (a *App) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, a.httpTimeout)
	defer cancel()

	// Сначала дожидаемся текущих HTTP-запросов, потом закрываем gRPC-клиенты,
	// иначе запросы в процессе упадут на закрытом соединении.
	var errs []error
	if err := a.server.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("shutdown http server: %w", err))
	}
	if err := a.userClient.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := a.catalogClient.Close(); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
