package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jakobsym/aura/internal/handler"
)

type Router struct {
	tokenHandler   *handler.TokenHandler
	accountHandler *handler.AccountHandler
	// add other handlers here (I.E: walletHandler)
}

func NewRouter(th *handler.TokenHandler, ah *handler.AccountHandler) *Router {
	return &Router{tokenHandler: th, accountHandler: ah}
}

func (r *Router) LoadRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Route("/v0/token", r.tokenRoutes)
	router.Route("/v0/track", r.accountRoutes)

	return router
}

func (r *Router) tokenRoutes(router chi.Router) {
	router.Get("/{token_address}", r.tokenHandler.GetTokenDetails)
}

func (r *Router) accountRoutes(router chi.Router) {
	router.Post("/", r.accountHandler.CreateUserEntry)
	router.Post("/{wallet_address}", r.accountHandler.TrackWallet)
	router.Put("/{wallet_address}", r.accountHandler.UntrackWallet)
}
