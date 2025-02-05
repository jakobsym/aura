package main

import (
	"net/http"

	"github.com/jakobsym/aura/internal/handler"
	"github.com/jakobsym/aura/internal/repository/postgres"
	"github.com/jakobsym/aura/internal/routes"
	"github.com/jakobsym/aura/internal/service"
)

func main() {
	db := "todo"

	// token
	repo := postgres.NewPostgresTokenRepo(db)
	service := service.NewTokenService(repo)
	handler := handler.NewTokenHandler(service)

	router := routes.NewRouter(handler)
	if err := http.ListenAndServe(":3000", router.LoadRoutes()); err != nil {
		panic(err)
	}

}
