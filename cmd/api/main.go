package main

import (
	"github.com/jakobsym/aura/internal/handler"
	"github.com/jakobsym/aura/internal/repository/postgres"
	"github.com/jakobsym/aura/internal/routes"
	"github.com/jakobsym/aura/internal/service"
)

func main() {
	db := "todo"
	repo := postgres.NewPostgresTokenRepo(db)
	service := service.NewTokenService(repo)
	handler := handler.NewTokenHandler(service)
	router := routes.NewRouter(handler)
	// listen and serve
}
