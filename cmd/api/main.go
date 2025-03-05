package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jakobsym/aura/internal/handler"
	"github.com/jakobsym/aura/internal/repository/postgres"
	solana "github.com/jakobsym/aura/internal/repository/solanarpc"
	"github.com/jakobsym/aura/internal/routes"
	"github.com/jakobsym/aura/internal/service"
)

func main() {
	db := postgres.PostgresConnectionPool()
	defer db.Close()

	accounts := []string{os.Getenv("WALLET_ADDRESS")}
	rpcConnection := solana.SolanaRpcConnection()
	wsConnection := solana.SolanaWebSocketConnection()

	solanaAccountRepo := solana.NewSolanaWebSocketRepo(wsConnection) // TODO: Rename `solanaWSRepo`
	accountPsqlRepo := postgres.NewPostgresAccountRepo(db)
	solanaAccountService := service.NewAccountService(solanaAccountRepo, accounts, accountPsqlRepo)
	accountHandler := handler.NewAccountHandler(solanaAccountService)

	tokenPsqlRepo := postgres.NewPostgresTokenRepo(db)
	solanaTokenRepo := solana.NewSolanaTokenRepo(rpcConnection)
	tokenService := service.NewTokenService(tokenPsqlRepo, solanaTokenRepo)
	tokenHandler := handler.NewTokenHandler(tokenService)

	router := routes.NewRouter(tokenHandler, accountHandler)

	ctx := context.Background()
	log.Println("service running on 3000")
	if err := solanaAccountService.MonitorAccountSubsription(ctx); err != nil {
		log.Fatalf("failed to start monitoring: %v", err)
	}
	if err := http.ListenAndServe(":3000", router.LoadRoutes()); err != nil {
		panic(err)
	}

}
