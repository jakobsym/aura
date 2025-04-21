// Package `repository` defines the interfaces for various data access operations
package repository

import (
	"context"
	"time"

	"github.com/jakobsym/aura/internal/domain"
)

type PostgresTokenRepo interface {
	// `DeleteToken` deletes a token entry from DB based on given tokenAddress
	DeleteToken(tokenAddress string) error
	// `CreateToken` creates token entry within DB after transforming token data
	CreateToken(token domain.TokenResponse) error
}

// `SolanaTokenRepo` defines operations for extracting token related data via RPC nodes.
type SolanaTokenRepo interface {
	// `GetTokenAge` retrieves the time of creation for a given tokenAddress
	GetTokenAge(ctx context.Context, tokenAddress string) (time.Time, error) // RPC

	// `GetTokenNameAndSymbol` retrieves the name and symbol for a given tokenAddress
	// returning as a slice [name, symbol]
	GetTokenNameAndSymbol(ctx context.Context, tokenAddress string) ([]string, error) // RPC

	// `GetTokenSupply` retrieves the total token supply for a given tokenAddress
	GetTokenSupply(ctx context.Context, tokenAddress string) (float64, error) // RPC

	// `GetTokenPrice`retrieves the token price for a given tokenAddress
	GetTokenPrice(ctx context.Context, tokenAddress string) (float64, error) // Jupiter

	// `GetTokenFDV` retrieves the Fully Diluted Value (FDV) for a given tokenAddress
	GetTokenFDV(ctx context.Context, price float64, supply float64) float64
}

// `SolanaWebSocketRepo` defines websocket based operations for extracting real-time transaction data
// via Helius Websocket RPC.
type SolanaWebSocketRepo interface {
	// `AccountListen` starts listening for updates across the websocket connection and returns a channel
	// that recieves a HeliusLogResponse
	AccountListen(ctx context.Context) (<-chan domain.HeliusLogResponse, error)

	// `StopAccountListen` terminates AccountListen process
	StopAccountListen(<-chan domain.HeliusLogResponse)

	// `LogsSubscribe` subscribe to transaction logs for a given walletAddress
	LogsSubscribe(ctx context.Context, walletAddress string, userId int) error

	// `AccountSubscribe` subscribes to an Account for a given walletAddress
	AccountSubscribe(ctx context.Context, walletAddress string, userId int) error

	// `AccountUnsubscribe` terminates AccountSubscribe process
	AccountUnsubscribe(ctx context.Context, walletAddress string, userId int) (bool, error)

	// `HandleWebSocketConnection` mangages Websocket connection
	HandleWebSocketConnection(ctx context.Context)

	// `StartReader` reads incoming messages sent via Websocket
	StartReader(ctx context.Context)

	// `GetTxnData` retrieves transaction details for a given transaction signature
	GetTxnData(signature string) (domain.TransactionResult, error)

	// `GetTxnSwapData` extracts swap information from a TransactionResult
	GetTxnSwapData(payload domain.TransactionResult) ([]domain.SwapResult, error)
}

// `AccountRepo` defines operations for managing user, wallet, and subscriptions
// witihn a PostgreSQL database.
type AccountRepo interface {
	// `CheckSubscription` check if a subscription exists for a given walletId
	CheckSubscription(walletId int) (bool, error)

	// `CreateSubscription` creates a new subscription entry for a given userId and walletId
	CreateSubscription(walletAddress string, userId, walletId int) error

	// `CreateWallet` creates a new wallet entry for a given walletAddress
	CreateWallet(walletAddress string) (int, error)

	// `RemoveSubscription` removes a subscription entry based on the given walletAddress and userId
	RemoveSubscription(walletAddress string, userId int) (bool, error)

	// `CreateUser` creates a new user entry based on telegramId
	CreateUser(telegramId int) error

	// `GetUserID` fetches a userId based on a given telegramId
	GetUserID(telegramId int) (int, error)

	// `GetWalletID` fetches a walletId based on a given walletAddress
	GetWalletID(walletAddress string) (int, error)

	// `SetWalletActive` marks a given `walletId` as active in the database
	SetWalletActive(walletId int) error
}
