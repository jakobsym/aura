package main

import (
	"net/http"

	"github.com/jakobsym/aura/internal/handler"
	"github.com/jakobsym/aura/internal/repository/postgres"
	solana "github.com/jakobsym/aura/internal/repository/solanarpc"
	"github.com/jakobsym/aura/internal/routes"
	"github.com/jakobsym/aura/internal/service"
)

func main() {
	db := postgres.PostgresConnectionPool()
	rpcConnection := solana.SolanaRpcConnection()
	defer db.Close()

	psqlRepo := postgres.NewPostgresTokenRepo(db)
	solanaRepo := solana.NewSolanaTokenRepo(rpcConnection)
	service := service.NewTokenService(psqlRepo, solanaRepo)
	handler := handler.NewTokenHandler(service)

	router := routes.NewRouter(handler)
	if err := http.ListenAndServe(":3000", router.LoadRoutes()); err != nil {
		panic(err)
	}
}
