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

type postgresAccountRepo struct {
	db *pgxpool.Pool
}

var (
	ErrWalletNotFound = errors.New("wallet not found in db")
)

func NewPostgresAccountRepo(db *pgxpool.Pool) repository.AccountRepo {
	return &postgresAccountRepo{db: db}
}

// TODO: All methods here perform SQL query which get passed to the AccountService()
func (ar *postgresAccountRepo) CheckSubscription(walletAddress string) (bool, error) {
	query := `SELECT subscription_active FROM wallets WHERE wallet_address = $1;`
	var isActive bool
	err := ar.db.QueryRow(context.TODO(), query, walletAddress).Scan(&isActive)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, fmt.Errorf("%w: %v", ErrWalletNotFound, err)
		}
		return false, fmt.Errorf("db error: %w", err)
	}
	return isActive, nil
}

// TODO: This needs a wallet_id to fill to successfully insert into subscriptsions
// This no longer needs to be a txn just normal insert
func (ar *postgresAccountRepo) CreateSubscription(walletAddress string, userId, walletId int) error {
	tx, err := ar.db.BeginTx(context.TODO(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.TODO())
	// maybe insert into wallet here?
	_, err = tx.Exec(context.TODO(), `INSERT into subscriptions(user_id, wallet_id, wallet_address) VALUES($1, $2, $3);`, userId, walletId, walletAddress)
	if err != nil {
		return fmt.Errorf("error inserting into join table: %w", err)
	}
	if err := tx.Commit(context.TODO()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	log.Printf("subscription created for: %d", userId)
	return nil
}

func (ar *postgresAccountRepo) SetSubscription(walletAddress string, userId int) error {
	tx, err := ar.db.BeginTx(context.TODO(), pgx.TxOptions{})
	defer tx.Rollback(context.TODO())
	if err != nil {
		return err
	}

	_, err = tx.Exec(context.TODO(), `UPDATE wallets SET subcription_active = TRUE WHERE wallet_address=$1;`, walletAddress)
	if err != nil {
		return fmt.Errorf("error updating wallets state: %w", err)
	}
	_, err = tx.Exec(context.TODO(), `INSERT into subscriptions(user_id, wallet_address) VALUES($1, $2);`, userId, walletAddress)
	if err != nil {
		return fmt.Errorf("error inserting into join table: %w", err)
	}
	if err := tx.Commit(context.TODO()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	log.Printf("subscription set for: %d", userId)
	return nil
}

func (ar *postgresAccountRepo) CreateWallet(walletAddress string) (int, error) {
	query := `INSERT into wallets(wallet_address, subscription_active) VALUES($1, TRUE) RETURNING id;`
	var walletId int
	err := ar.db.QueryRow(context.TODO(), query, walletAddress).Scan(&walletId)
	if err != nil {
		return -1, fmt.Errorf("error inseting into wallet table: %w", err)
	}
	return walletId, nil
}

// Removes the entry from subscription table
// checks if other users are tracking respective wallet address
// returns (true, nil) on success where true == wallet still being tracked by someone
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

// Note: you can only use ON CONFLICT if your column has a UNIQUE constraint
func (ar *postgresAccountRepo) CreateUser(userId int) error {
	query := `INSERT into users(telegram_id) VALUES($1);`
	_, err := ar.db.Exec(context.TODO(), query, userId)
	if err != nil {
		return err
	}
	return nil
}

// Handle no rows error?
func (ar *postgresAccountRepo) GetUserID(telegramId int) (int, error) {
	query := `SELECT id FROM users WHERE telegram_id = $1`
	var userId int
	err := ar.db.QueryRow(context.TODO(), query, telegramId).Scan(&userId)
	if err != nil {
		return -1, err
	}
	return userId, nil
}
