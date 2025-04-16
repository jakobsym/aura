// Package `postgres` provides implementations of respository interfaces using PostgreSQL.
package postgres

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jakobsym/aura/internal/repository"
)

// `postgresTokenRepo` implements the respository.TokenRepo interface using PostgreSQL
type postgresTokenRepo struct {
	db *pgxpool.Pool
	//db *sql.DB
}

// `NewPostgresTokenRepo` creates and returns a new PostgreSQL implementation
// of the TokenRepo interface.
func NewPostgresTokenRepo(db *pgxpool.Pool) repository.TokenRepo {
	return &postgresTokenRepo{db: db}
}

// `PostgresConnectionPool` creates a pgxpool.Pool connection to the provided
// DB_URL
func PostgresConnectionPool() *pgxpool.Pool {
	config, err := pgxpool.ParseConfig(os.Getenv("DB_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 15 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute
	config.MaxConnLifetimeJitter = 5 * time.Minute

	dbpool, err := pgxpool.NewWithConfig(context.TODO(), config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	log.Printf("connected to db")
	return dbpool
}
