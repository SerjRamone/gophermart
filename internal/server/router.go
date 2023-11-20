package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewRouter returns chi.Router
func NewRouter(secret []byte, tokenExpr int, storage Storage) chi.Router {
	baseHandler := NewBaseHandler(secret, tokenExpr, storage)
	mux := chi.NewRouter()
	mux.Use(RequestLogger)

	// api
	mux.Route("/api/user", func(r chi.Router) {
		r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
			baseHandler.Register(r.Context(), w, r)
		})

		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			baseHandler.Login(r.Context(), w, r)
		})
	})

	return mux
}