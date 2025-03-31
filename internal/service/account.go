package service

import (
	"context"
	"fmt"
	"log"

	"github.com/jakobsym/aura/internal/repository"
)

type AccountService struct {
	solanaRepo repository.SolanaWebSocketRepo
	psqlRepo   repository.AccountRepo
}

func NewAccountService(sr repository.SolanaWebSocketRepo, pr repository.AccountRepo) *AccountService {
	return &AccountService{solanaRepo: sr, psqlRepo: pr}
}

// TODO: Here is where you will relay updates to users that have subscribed to the respective accounts/walletAddresses?
// using the TG api
// within the response the ID is returned which can be used to map the response to the respective user
func (as *AccountService) MonitorAccountSubsription(ctx context.Context) error {
	as.solanaRepo.HandleWebSocketConnection(ctx)
	// websocket data
	updates, err := as.solanaRepo.AccountListen(ctx)
	if err != nil {
		return fmt.Errorf("service listen error: %v", err)
	}
	go func() {
		defer as.solanaRepo.StopAccountListen(updates)
		for update := range updates {
			log.Printf("transaction detected: %+v", update)
			// currently just letting the repo print updates
			// later you'll process these updates here
		}
	}()

	return nil
}

// TODO: Logic here might be wrong?

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

	return as.solanaRepo.AccountSubscribe(context.TODO(), walletAddress, userId)
}

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

func (as *AccountService) CreateUser(userId int) error {
	err := as.psqlRepo.CreateUser(userId)
	if err != nil {
		log.Printf("error creating user: %v", err)
		return err
	}
	return nil
}
