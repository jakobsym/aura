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
	AccountListen(ctx context.Context) (<-chan domain.HeliusLogResponse, error)
	StopAccountListen(<-chan domain.HeliusLogResponse)
	LogsSubscribe(ctx context.Context, walletAddress string, userId int) error
	AccountSubscribe(ctx context.Context, walletAddress string, userId int) error
	AccountUnsubscribe(ctx context.Context, walletAddress string, userId int) (bool, error)
	HandleWebSocketConnection(ctx context.Context)
	StartReader(ctx context.Context)
}

// PSQL based repo
type AccountRepo interface {
	CheckSubscription(walletId int) (bool, error)
	CreateSubscription(walletAddress string, userId, walletId int) error
	SetSubscription(walletAddress string, userId, walletId int) error
	CreateWallet(walletAddress string) (int, error)
	RemoveSubscription(walletAddress string, userId int) (bool, error)
	CreateUser(telegramId int) error
	GetUserID(telegramId int) (int, error)
	GetWalletID(walletAddress string) (int, error)
	SetWalletActive(walletId int) error
	// DB methods
}
