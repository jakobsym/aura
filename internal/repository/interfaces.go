package repository

import (
	"context"
	"time"

	"github.com/jakobsym/aura/internal/domain"
)

type TokenRepo interface{}

// RPC based repo
type SolanaTokenRepo interface {
	GetTokenAge(ctx context.Context, tokenAddress string) (time.Time, error)          // RPC
	GetTokenNameAndSymbol(ctx context.Context, tokenAddress string) ([]string, error) // RPC
	GetTokenSupply(ctx context.Context, tokenAddress string) (float64, error)         // RPC
	GetTokenPrice(ctx context.Context, tokenAddress string) (float64, error)          // Jupiter
	GetTokenFDV(ctx context.Context, price float64, supply float64) float64
	//GetTokenHolders(ctx context.Context, tokenAddress string) (uint64, error) // Helius
}

// RPC WS based repo
type SolanaWebSocketRepo interface {
	AccountListen(ctx context.Context) (<-chan domain.AccountResponse, error)
	AccountSubscribe(ctx context.Context, walletAddress, userId string) error
	AccountUnsubscribe(ctx context.Context, walletAddress, userId string) (bool, error)
	HandleWebSocketConnection(ctx context.Context)
}

// PSQL based repo
type AccountRepo interface {
	CheckSubscription(walletAddress string) (bool, error)
	CreateSubscription(walletAddress, userId string) error
	SetSubscription(walletAddress, userId string) error
	CreateWallet(walletAddress string) error
	RemoveSubscription(walletAddress, userId string) (bool, error)
	CreateUser(userId string) error
	// DB methods
}
