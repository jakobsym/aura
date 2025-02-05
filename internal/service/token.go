package service

import "github.com/jakobsym/aura/internal/repository"

type TokenService struct {
	repo repository.TokenRepo
}

func NewTokenService(r repository.TokenRepo) *TokenService {
	return &TokenService{repo: r}
}

// call methods from TokenRepo interface
