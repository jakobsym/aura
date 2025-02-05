package handler

import (
	"github.com/jakobsym/aura/internal/service"
)

type TokenHandler struct {
	s *service.TokenService
}

func NewTokenHandler(s *service.TokenService) *TokenHandler {
	return &TokenHandler{s: s}
}

// req/res related methods
