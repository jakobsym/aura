package repository

import (
	"context"
	"time"

	token_metadata "github.com/gagliardetto/metaplex-go/clients/token-metadata"
)

// import domain models

type TokenRepo interface {
	// methods
}

type WalletRepo interface {
	// methods
}

type SolanaTokenRepo interface {
	GetTokenAge(ctx context.Context, tokenAddress string) (time.Time, error)                     // RPC
	GetTokenMetadata(ctx context.Context, tokenAddress string) (*token_metadata.Metadata, error) // RPC
	GetTokenSupply(ctx context.Context, tokenAddress string) (uint64, error)                     // RPC
	GetTokenPrice(ctx context.Context, tokenAddress string) (float64, error)                     // Jupiter                                                                // Jupiter
	GetTokenFDV(ctx context.Context, price float64, supply uint64) (float64, error)
	//GetTokenHolders(ctx context.Context, tokenAddress string) (uint64, error) // Helius
}
