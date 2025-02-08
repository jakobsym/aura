package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
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
	tokenAddress := chi.URLParam(r, "token_address")
	res := th.s.GetTokenData(r.Context(), tokenAddress)
	fmt.Println(res)
	// Call a function within 'service' package, and this function is responsible
	// for building a TokenResponse
	// encode the TokenResponse into JSON which gets sent as response
}
