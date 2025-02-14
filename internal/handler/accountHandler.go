package handler

import "github.com/jakobsym/aura/internal/service"

type AccountHandler struct {
	as *service.AccountService
}

func NewAccountHandler(as *service.AccountService) *AccountHandler {
	return &AccountHandler{as: as}
}

// Websocket related things with response/req here
