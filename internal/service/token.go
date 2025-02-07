package service

import (
	"context"

	"github.com/jakobsym/aura/internal/domain"
	"github.com/jakobsym/aura/internal/repository"
)

type TokenService struct {
	psqlRepo   repository.TokenRepo
	solanaRepo repository.SolanaTokenRepo
}

func NewTokenService(r repository.TokenRepo, sr repository.SolanaTokenRepo) *TokenService {
	return &TokenService{psqlRepo: r, solanaRepo: sr}
}

// call methods from TokenRepo && SolanaTokenRepo interface
func (ts *TokenService) GetTokenPrice(ctx context.Context, tokenAddress string) (float64, error) {
	return ts.solanaRepo.GetTokenPrice(ctx, tokenAddress)
}

func (ts *TokenService) GetTokenData(ctx context.Context, tokenAddress string) (*domain.TokenResponse, error) {
	return nil, nil
}
