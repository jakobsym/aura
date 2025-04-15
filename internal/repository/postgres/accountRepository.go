// Package `postgres` provides implementations of respository interfaces using PostgreSQL.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jakobsym/aura/internal/repository"
)

// `postgresAccountRepo` implements the respository.AccountRepo interface using PostgreSQL
type postgresAccountRepo struct {
	db *pgxpool.Pool
}

var (
	// `ErrWalletNotFound` returned when requested wallet is not found in the DB
	ErrWalletNotFound = errors.New("wallet not found in db")
)

// `NewPostgresAccountRepo` creates and returns a new PostgreSQL implementation
// of the AccountRepo interface.
func NewPostgresAccountRepo(db *pgxpool.Pool) repository.AccountRepo {
	return &postgresAccountRepo{db: db}
}

// `CheckSubscription` verifies if a subscription is active for a given walletId
// Returns True if subscription is active, False otherwise
func (ar *postgresAccountRepo) CheckSubscription(walletId int) (bool, error) {
	query := `SELECT subscription_active FROM wallets WHERE id = $1;`
	var isActive bool
	err := ar.db.QueryRow(context.TODO(), query, walletId).Scan(&isActive)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, fmt.Errorf("%w: %v", ErrWalletNotFound, err)
		}
		return false, fmt.Errorf("db error: %w", err)
	}
	return isActive, nil
}

// `CreateSubsciption` adds a new subscription record creating a (user - wallet) connection
func (ar *postgresAccountRepo) CreateSubscription(walletAddress string, userId, walletId int) error {
	query := `INSERT into subscriptions(user_id, wallet_id, wallet_address) VALUES ($1, $2, $3);`
	_, err := ar.db.Exec(context.TODO(), query, userId, walletId, walletAddress)
	if err != nil {
		return fmt.Errorf("error inserting into join table: %v", err)
	}
	log.Printf("subscription set for userID: %d | walletID: %d", userId, walletId)
	return nil
}

// `RemoveSubsciption` deletes a subscription for a given walletAddress and userId
// Checks if other Users are tracking the wallet
// Returns True if wallet is still tracked by other users, false otherwise
// Transaction is used to ensure data consistency.
func (ar *postgresAccountRepo) RemoveSubscription(walletAddress string, userId int) (bool, error) {
	tx, err := ar.db.BeginTx(context.TODO(), pgx.TxOptions{})
	if err != nil {
		return false, err
	}
	defer tx.Rollback(context.TODO())
	_, err = tx.Exec(context.TODO(), `DELETE FROM subscriptions WHERE wallet_address=$1 and user_id=$2;`, walletAddress, userId)
	if err != nil {
		return false, fmt.Errorf("failed to perform operation: %w", err)
	}

	// find how many users are tracking respective wallet
	var userCount int
	err = tx.QueryRow(context.TODO(), `SELECT COUNT(*) FROM subscriptions where wallet_address=$1;`, walletAddress).Scan(&userCount)
	if err != nil {
		return false, fmt.Errorf("failed to perform operation: %w", err)
	}

	// set inactive if no-one is tracking
	if userCount == 0 {
		_, err = tx.Exec(context.TODO(), `UPDATE wallet SET subcription_active = FALSE WHERE wallet_address=$1;`, walletAddress)
		if err != nil {
			return false, fmt.Errorf("failed to perform operation: %w", err)
		}
	}

	if err := tx.Commit(context.TODO()); err != nil {
		return false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return userCount > 0, nil
}

// `CreateWallet` adds a new wallet record with an active subscription status
// Returns the ID of the newly created wallet.
func (ar *postgresAccountRepo) CreateWallet(walletAddress string) (int, error) {
	query := `INSERT into wallets(wallet_address, subscription_active) VALUES($1, TRUE) RETURNING id;`
	var walletId int
	err := ar.db.QueryRow(context.TODO(), query, walletAddress).Scan(&walletId)
	if err != nil {
		return -1, fmt.Errorf("error inseting into wallet table: %w", err)
	}
	return walletId, nil
}

// Using Upsert
// if not found then create
func (ar *postgresAccountRepo) GetWalletID(walletAddress string) (int, error) {
	tx, err := ar.db.BeginTx(context.TODO(), pgx.TxOptions{})
	defer tx.Rollback(context.TODO())
	if err != nil {
		return -1, fmt.Errorf("error creating transaction -> GetWalletID(): %v", err)
	}
	var walletId int
	err = tx.QueryRow(context.TODO(), `SELECT id from wallets where wallet_address=$1;`, walletAddress).Scan(&walletId)
	if err != nil {
		if err == pgx.ErrNoRows {
			err = tx.QueryRow(context.TODO(), `INSERT into wallets(wallet_address) VALUES($1) RETURNING id;`, walletAddress).Scan(&walletId)
			if err != nil {
				return -1, fmt.Errorf("\nerror inserting wallet")
			}
		} else {
			return -1, fmt.Errorf("\nerror querying wallets table")
		}
	}
	if err := tx.Commit(context.TODO()); err != nil {
		return -1, fmt.Errorf("failed to execute transaction -> GetWalletID(): %v", err)
	}

	return walletId, nil
}

func (ar *postgresAccountRepo) SetWalletActive(walletId int) error {
	query := `UPDATE wallets SET subscription_active = TRUE WHERE id = $1;`
	_, err := ar.db.Exec(context.TODO(), query, walletId)
	if err != nil {
		return fmt.Errorf("error executing update -> SetWalletActive(): %v", err)
	}
	return nil
}

func (ar *postgresAccountRepo) CreateUser(userId int) error {
	query := `INSERT into users(telegram_id) VALUES($1);`
	_, err := ar.db.Exec(context.TODO(), query, userId)
	if err != nil {
		return err
	}
	return nil
}

func (ar *postgresAccountRepo) GetUserID(telegramId int) (int, error) {
	query := `SELECT id FROM users WHERE telegram_id = $1`
	var userId int
	err := ar.db.QueryRow(context.TODO(), query, telegramId).Scan(&userId)
	if err != nil {
		return -1, err
	}
	return userId, nil
}
