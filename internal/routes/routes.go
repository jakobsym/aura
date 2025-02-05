package routes

import "github.com/jakobsym/aura/internal/handler"

type Router struct {
	hdlr *handler.TokenHandler
}

func NewRouter(h *handler.TokenHandler) *Router {
	return &Router{hdlr: h}
}
