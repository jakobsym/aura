package postgres

import (
	"github.com/jakobsym/aura/internal/repository"
)

type postgresTokenRepo struct {
	//db *sql.DB
	db string
}

func NewPostgresTokenRepo(db string /*db *sql.DB*/) repository.TokenRepo {
	return &postgresTokenRepo{db: db}
}

// PSQL Queries for a Token table
