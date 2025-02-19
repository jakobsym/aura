package service

import (
	"context"
	"fmt"
	"log"

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
	if err := as.solanaRepo.AccountSubscribe(ctx, as.accounts); err != nil {
		return fmt.Errorf("service subscription error: %v", err)
	}
	// websocket data
	updates, err := as.solanaRepo.AccountListen(ctx)
	if err != nil {
		return fmt.Errorf("service listen error: %v", err)
	}
	go func() {
		for update := range updates {
			log.Printf("transaction detected: %+v", update)
			// currently just letting the repo print updates
			// later you'll process these updates here
		}
	}()

	return nil
}
