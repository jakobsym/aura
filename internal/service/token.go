package service

import (
	"context"

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

// This method will build a token which gets sent to the handler
// Call rest of methods (using goroutines)
func (ts *TokenService) GetTokenData(ctx context.Context, tokenAddress string) (float64, error) {
	price, err := ts.solanaRepo.GetTokenPrice(ctx, tokenAddress)
	if err != nil {
		return 0, nil
	}
	supply, err := ts.solanaRepo.GetTokenSupply(ctx, tokenAddress)
	if err != nil {
		return 0, nil
	}
	res := ts.solanaRepo.GetTokenFDV(ctx, price, supply)
	return res, nil
	/*
		res, err := ts.solanaRepo.GetTokenNameAndSymbol(ctx, tokenAddress)
		if err != nil {
			return nil, fmt.Errorf("error retrieving token metadata")
		}
		return res, nil
	*/
}
