package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jakobsym/aura/internal/handler"
)

type Router struct {
	tokenHandler *handler.TokenHandler
	// add other handlers here (I.E: walletHandler)
}

func NewRouter(th *handler.TokenHandler) *Router {
	return &Router{tokenHandler: th}
}

func (r *Router) LoadRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Route("/v0/token", r.tokenRoutes)

	return router
}

func (r *Router) tokenRoutes(router chi.Router) {
	router.Get("/{token_address}", r.tokenHandler.GetTokenDetails)
	router.Post("/{token_address}", r.tokenHandler.CreateToken)
	// rest of token handler functions
}
