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
		for update := range updates {
			log.Printf("transaction detected: %+v", update)
			// currently just letting the repo print updates
			// later you'll process these updates here
		}
	}()

	return nil
}

func (as *AccountService) TrackWallet(walletAddress string, telegramId int) error {
	// TODO: Get id for user based on the passed in telegramId
	userId, err := as.psqlRepo.GetUserID(telegramId)
	if err != nil {
		return err
	}
	// check if subscription active for given walletAddress
	active, err := as.psqlRepo.CheckSubscription(walletAddress)
	// log.Printf("subscription active: %t", active)
	// wallet !exist
	if err != nil {
		if errors.Is(err, postgres.ErrWalletNotFound) {
			walletId, err := as.psqlRepo.CreateWallet(walletAddress)
			if err != nil {
				return err
			}

			if err := as.psqlRepo.CreateSubscription(walletAddress, userId, walletId); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	// wallet exists, !active subscription
	if !active {
		if err := as.psqlRepo.SetSubscription(walletAddress, userId); err != nil {
			log.Printf("error setting subscription: %v", err)
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
