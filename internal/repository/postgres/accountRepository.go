package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

func (ar *postgresAccountRepo) CreateSubscription(walletAddress, userId string) error {
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
	return nil
}

func (ar *postgresAccountRepo) SetSubscription(walletAddress, userId string) error {
	tx, err := ar.db.BeginTx(context.TODO(), pgx.TxOptions{})
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

func (ar *postgresAccountRepo) UntrackWallet(walletAddress string) any {
	return nil
}

// TODO: When user starts bot, this will invoke
func (ar *postgresAccountRepo) CreateUser(userId string) any {
	return nil
}
