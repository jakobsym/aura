package handler

import (
	"encoding/json"
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
	if tokenAddress == "" {
		http.Error(w, "must provide valid token address", http.StatusBadRequest)
		return
	}
	res, err := th.s.GetTokenData(r.Context(), tokenAddress)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
