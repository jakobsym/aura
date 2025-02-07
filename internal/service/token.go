package service

import "github.com/jakobsym/aura/internal/repository"

type TokenService struct {
	psqlRepo   repository.TokenRepo
	solanaRepo repository.SolanaTokenRepo
}

func NewTokenService(r repository.TokenRepo, sr repository.SolanaTokenRepo) *TokenService {
	return &TokenService{psqlRepo: r, solanaRepo: sr}
}

// call methods from TokenRepo && SolanaTokenRepo interface
