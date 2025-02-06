package handler

import (
	"fmt"
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
func (th *TokenHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	fmt.Println("CreateToken World")
}

func (th *TokenHandler) GetTokenDetails(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetTokenDetails World")
}
