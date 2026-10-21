package handler

import (
	"github.com/artlink52/marketplace/internal/gateway/config"
	"github.com/artlink52/marketplace/internal/gateway/handler/catalog"
	"github.com/artlink52/marketplace/internal/gateway/handler/order"
	"github.com/artlink52/marketplace/internal/gateway/handler/user"
	"github.com/artlink52/marketplace/internal/gateway/middleware"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	User    *user.Handler
	Catalog *catalog.Handler
	Order   *order.Handler
}

func NewRouter(h Handlers, cfg config.Config) chi.Router {
	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/login", h.User.Login)
		r.Post("/auth/register", h.User.Register)
		r.Get("/categories", h.Catalog.ListCategories)
		r.Get("/products", h.Catalog.ListProducts)
		r.Get("/products/{id}", h.Catalog.GetProduct)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.JWTSecret))
			r.Get("/users/{id}", h.User.GetUser)
			r.Post("/orders", h.Order.CreateOrder)

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))
				r.Post("/categories", h.Catalog.CreateCategory)
			})

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("seller"))
				r.Post("/products", h.Catalog.CreateProduct)
			})
		})
	})

	return r
}
