// Package `handler` implements HTTP request handlers that connect with API endpoints
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jakobsym/aura/internal/service"
)

// `TokenHandler` handles HTTP requests for token related business logic
type TokenHandler struct {
	s *service.TokenService
}

// `NewTokenHandler` creates new TokenHandler instance with dependency injection
func NewTokenHandler(s *service.TokenService) *TokenHandler {
	return &TokenHandler{s: s}
}

// `GetTokenDetails` handles GET requests for token information
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

// `DeleteToken` handles DELETE requests for tokens
func (th *TokenHandler) DeleteToken(w http.ResponseWriter, r *http.Request) {
	tokenAddress := chi.URLParam(r, "token_address")
	if tokenAddress == "" {
		http.Error(w, "must provide valid token address", http.StatusBadRequest)
		return
	}
	err := th.s.DeleteToken(r.Context(), tokenAddress)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("token deleted")
}
