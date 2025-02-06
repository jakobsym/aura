package postgres

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jakobsym/aura/internal/repository"
)

type postgresTokenRepo struct {
	db *pgxpool.Pool
	//db *sql.DB
}

func NewPostgresTokenRepo(db *pgxpool.Pool) repository.TokenRepo {
	return &postgresTokenRepo{db: db}
}

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
	return dbpool
}

// PSQL Queries for a Token table
