package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jakobsym/aura/internal/repository"
)

type TokenService struct {
	//psqlRepo   repository.TokenRepo
	solanaRepo repository.SolanaTokenRepo
}

func NewTokenService( /* r repository.TokenRepo, */ sr repository.SolanaTokenRepo) *TokenService {
	return &TokenService{ /*psqlRepo: r, */ solanaRepo: sr}
}

// call methods from TokenRepo && SolanaTokenRepo interface
func (ts *TokenService) GetTokenPrice(ctx context.Context, tokenAddress string) (float64, error) {
	return ts.solanaRepo.GetTokenPrice(ctx, tokenAddress)
}

// This method will build a token which gets sent to the handler
// Call rest of methods (using goroutines)
func (ts *TokenService) GetTokenData(ctx context.Context, tokenAddress string) (time.Time, error) {
	res, err := ts.solanaRepo.GetTokenAge(ctx, tokenAddress)
	if err != nil {
		return time.Time{}, fmt.Errorf("error retrieving token metadata")
	}
	return res, nil
	/*
		res, err := ts.solanaRepo.GetTokenNameAndSymbol(ctx, tokenAddress)
		if err != nil {
			return nil, fmt.Errorf("error retrieving token metadata")
		}
		return res, nil
	*/
}
