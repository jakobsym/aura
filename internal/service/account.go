// Package `service` calls repository methods to implement business logic
package service

import (
	"context"
	"fmt"
	"log"

	"github.com/jakobsym/aura/internal/repository"
)

// `AccountService` provides wallet tracking business logic by receiving data
// from the SolanaWebSocketRepo, and Postgres AccountRepo
type AccountService struct {
	solanaRepo repository.SolanaWebSocketRepo
	psqlRepo   repository.AccountRepo
}

// `NewAccountService` creates and returns a new AccountService with required dependencies
func NewAccountService(sr repository.SolanaWebSocketRepo, pr repository.AccountRepo) *AccountService {
	return &AccountService{solanaRepo: sr, psqlRepo: pr}
}

// `MonitorAccountSubscription` initiates and manages wallet monitoring subscription(s).
// Establishes a websocket connection and processes incoming wallet updates
// Note: This method runs indefinitely until context cancellation, or connection failure
func (as *AccountService) MonitorAccountSubsription(ctx context.Context) error {
	as.solanaRepo.HandleWebSocketConnection(ctx)
	updates, err := as.solanaRepo.AccountListen(ctx)
	if err != nil {
		return fmt.Errorf("service listen error: %v", err)
	}
	go func() {
		defer as.solanaRepo.StopAccountListen(updates)
		for update := range updates {
			log.Printf("transaction detected: %+v", update)
		}
	}()
	return nil
}

// `TrackWallet` starts tracking a wallet for a specific telegram user.
// Creates necessary database records, and subscribes to Solana log events for updates.
func (as *AccountService) TrackWallet(walletAddress string, telegramId int) error {
	userId, err := as.psqlRepo.GetUserID(telegramId)
	if err != nil {
		return err
	}
	//log.Printf("userID: %d\n", userId)
	walletId, err := as.psqlRepo.GetWalletID(walletAddress)
	if err != nil {
		return err
	}
	//log.Printf("walletID: %d\n", walletId)
	active, err := as.psqlRepo.CheckSubscription(walletId)
	if err != nil {
		return err
	}
	//log.Printf("walletID: %d | activity_status: %t\n", walletId, active)
	if !active {
		err := as.psqlRepo.SetWalletActive(walletId)
		if err != nil {
			return err
		}
		if err := as.psqlRepo.CreateSubscription(walletAddress, userId, walletId); err != nil {
			return err
		}
	}

	return as.solanaRepo.LogsSubscribe(context.TODO(), walletAddress, userId)
}

// `UntrackWallet` stops tracking a wallet for a given telegram user.
// removes subscription and cleans up resources
func (as *AccountService) UntrackWallet(walletAddress string, userId int) error {
	isTracked, err := as.psqlRepo.RemoveSubscription(walletAddress, userId)
	if err != nil {
		return err
	}
	if !isTracked {
		_, err := as.solanaRepo.AccountUnsubscribe(context.TODO(), walletAddress, userId)
		if err != nil {
			return err
		}
	}
	return nil
}

// `CreateUser` creates a new user record in the database
// Once telegram user instantiates application/bot this method is called
// to create a user record in background.
func (as *AccountService) CreateUser(userId int) error {
	err := as.psqlRepo.CreateUser(userId)
	if err != nil {
		log.Printf("error creating user: %v", err)
		return err
	}
	return nil
}
