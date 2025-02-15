package repository

import (
	"context"
	"time"

	"github.com/jakobsym/aura/internal/domain"
)

// import domain models

type TokenRepo interface {
	// methods
}

type WalletRepo interface {
	// methods
}

type SolanaTokenRepo interface {
	GetTokenAge(ctx context.Context, tokenAddress string) (time.Time, error)          // RPC
	GetTokenNameAndSymbol(ctx context.Context, tokenAddress string) ([]string, error) // RPC
	GetTokenSupply(ctx context.Context, tokenAddress string) (float64, error)         // RPC
	GetTokenPrice(ctx context.Context, tokenAddress string) (float64, error)          // Jupiter
	GetTokenFDV(ctx context.Context, price float64, supply float64) float64
	//GetTokenHolders(ctx context.Context, tokenAddress string) (uint64, error) // Helius
}

// Websocket based RPC methods
type SolanaWebSocketRepo interface {
	AccountListen(ctx context.Context) (<-chan domain.AccountResponse, error) // channel that recieves AccountResponse
	AccountSubscribe(ctx context.Context, accounts []string) error
}
