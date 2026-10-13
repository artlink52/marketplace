package handler

import (
	"net/http"

	"github.com/artlink52/marketplace/internal/gateway/handler/user"
	"github.com/go-chi/chi/v5"
)

func NewRouter(h *user.Handler, auth func(handler http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Post("/auth/login", h.Login)
	r.Post("/auth/register", h.Register)

	r.Group(func(r chi.Router) {
		r.Use(auth)
		r.Get("/users/{id}", h.GetUser)

		r.Group(func(r chi.Router) {

		})
	})

	return r
}
