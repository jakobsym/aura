package handler

import (
	"net/http"

	"github.com/jakobsym/aura/internal/service"
)

type TokenHandler struct {
	s *service.TokenService
}

func NewTokenHandler(s *service.TokenService) *TokenHandler {
	return &TokenHandler{s: s}
}

// req/res related methods
func (th *TokenHandler) GetTokenDetails(w http.ResponseWriter, r *http.Request) {
	// Call a function within 'service' package, and this function is responsible
	// for building a TokenResponse
}
