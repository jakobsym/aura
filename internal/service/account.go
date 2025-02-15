package service

import (
	"context"

	"github.com/jakobsym/aura/internal/repository"
)

type AccountService struct {
	solanaRepo repository.SolanaWebSocketRepo
	accounts   []string
}

// TODO: The accounts are currently supplied via memory, but future will pull from DB
func NewAccountService(sr repository.SolanaWebSocketRepo, acc []string) *AccountService {
	return &AccountService{solanaRepo: sr, accounts: acc}
}

// call repo methods
func (as *AccountService) MonitorAccountSubsription(ctx context.Context) error {
	// websocket data
	return nil
}
