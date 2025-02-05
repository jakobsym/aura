package postgres

import (
	"database/sql"

	"github.com/jakobsym/aura/internal/repository"
)

type postgresTokenRepo struct {
	db *sql.DB
}

func NewPostgresTokenRepo(db *sql.DB) repository.TokenRepo {
	return &postgresTokenRepo{db: db}
}

// PSQL Queries for a Token table
