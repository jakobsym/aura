// Package `routes` defines HTTP routes for internal API
package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jakobsym/aura/internal/handler"
)

// `Router` aggregates all API handlers
type Router struct {
	tokenHandler   *handler.TokenHandler
	accountHandler *handler.AccountHandler
}

// `NewRouter` creates a new Router instance with its handlers being injected
func NewRouter(th *handler.TokenHandler, ah *handler.AccountHandler) *Router {
	return &Router{tokenHandler: th, accountHandler: ah}
}

// `LoadRoutes` initalizes and returns configured chi.Mux router
// with versioned route groups
func (r *Router) LoadRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Route("/v0/token", r.tokenRoutes)
	router.Route("/v0/track", r.accountRoutes)

	return router
}

// `tokenRoutes` defines routes for token operations under /v0/token path
func (r *Router) tokenRoutes(router chi.Router) {
	// GET /v0/token/...
	router.Get("/{token_address}", r.tokenHandler.GetTokenDetails)
	// DELETE /v0/token/...
	router.Delete("/{token_address}", r.tokenHandler.DeleteToken)
}

// `accountRoutes` defines routes for wallet tracking under /v0/track path
func (r *Router) accountRoutes(router chi.Router) {
	// POST /v0/track/...
	router.Post("/", r.accountHandler.CreateUserEntry)
	// POST /v0/track/...
	router.Post("/{wallet_address}", r.accountHandler.TrackWallet)
	// PUT /v0/track/...
	router.Put("/{wallet_address}", r.accountHandler.UntrackWallet)
}
