package handler

import "github.com/jakobsym/aura/internal/service"

type AccountHandler struct {
	as *service.AccountService
}

func NewAccountHandler(as *service.AccountService) *AccountHandler {
	return &AccountHandler{as: as}
}

// TODO: This is where internal websocket will go
// 		This is where AccountService will send its responses to in the future
