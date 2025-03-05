package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jakobsym/aura/internal/repository"
	"github.com/jakobsym/aura/internal/repository/postgres"
)

type AccountService struct {
	solanaRepo repository.SolanaWebSocketRepo
	psqlRepo   repository.AccountRepo
	accounts   []string
}

// TODO: The accounts are currently supplied via memory, but future will pull from DB
func NewAccountService(sr repository.SolanaWebSocketRepo, acc []string, pr repository.AccountRepo) *AccountService {
	return &AccountService{solanaRepo: sr, accounts: acc, psqlRepo: pr}
}

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

// TODO: Above `MonitorAccountSubscription()` method will go here?
func (as *AccountService) TrackWallet(walletAddress, userId string) error {
	// check if subscription active for given walletAddress
	active, err := as.psqlRepo.CheckSubscription(walletAddress)
	if err != nil {
		if errors.Is(err, postgres.ErrWalletNotFound) {
			if err := as.psqlRepo.CreateSubscription(walletAddress, userId); err != nil {
				return err
			}
			// accountSubscribe() call here
		} else {
			return err
		}
	}

	// wallet exists, not active subscription
	if !active {
		if err := as.psqlRepo.SetSubscription(walletAddress, userId); err != nil {
			return err
		}
		// set subscription for wallet and associate with userId
	}

	// Wallet is actively subscribed to already
	return nil

	// if first time wallet is being tracked subscribe to Solana WS via
	// AccountSubscribe()
}
