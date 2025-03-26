package postgres

import (
	"context"
	"database/sql"
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
	ErrWalletNotFound = errors.New("wallet not found")
)

func NewPostgresAccountRepo(db *pgxpool.Pool) repository.AccountRepo {
	return &postgresAccountRepo{db: db}
}

// TODO: All methods here perform SQL query which get passed to the AccountService()
func (ar *postgresAccountRepo) CheckSubscription(walletAddress string) (bool, error) {
	query := `SELECT subscription_active FROM wallet WHERE wallet_address = $1;`
	var isActive bool
	err := ar.db.QueryRow(context.TODO(), query, walletAddress).Scan(&isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("%w: %v", ErrWalletNotFound, err)
		}
		return false, fmt.Errorf("db error: %w", err)
	}
	return isActive, nil
}

func (ar *postgresAccountRepo) CreateSubscription(walletAddress string, userId int) error {
	tx, err := ar.db.BeginTx(context.TODO(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.TODO())

	_, err = tx.Exec(context.TODO(), `INSERT into wallet(wallet_address, subscription_active) VALUES($1, TRUE);`, walletAddress)
	if err != nil {
		return fmt.Errorf("error inserting into wallet table: %w", err)
	}
	_, err = tx.Exec(context.TODO(), `INSERT into subscriptions(user_id, wallet_address) VALUES($1, $2) ON CONFLICT (user_id, wallet_address) DO NOTHING;`, userId, walletAddress)
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

	_, err = tx.Exec(context.TODO(), `UPDATE wallet SET subcription_active = TRUE WHERE wallet_address=$1;`, walletAddress)
	if err != nil {
		return fmt.Errorf("error updating wallet state: %w", err)
	}
	_, err = tx.Exec(context.TODO(), `INSERT into subscriptions(user_id, wallet_address) VALUES($1, $2) ON CONFLICT (user_id, wallet_address) DO NOTHING;`, userId, walletAddress)
	if err != nil {
		return fmt.Errorf("error inserting into join table: %w", err)
	}
	if err := tx.Commit(context.TODO()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	log.Printf("subscription set for: %d", userId)
	return nil
}

func (ar *postgresAccountRepo) CreateWallet(walletAddress string) error {
	query := `INSERT into wallet(wallet_address) VALUES($1) ON CONFLICT (wallet_address) DO NOTHING;`
	_, err := ar.db.Exec(context.TODO(), query, walletAddress)
	if err != nil {
		return fmt.Errorf("error inseting into wallet table: %w", err)
	}
	return nil
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
