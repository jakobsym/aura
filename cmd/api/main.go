// Package `main` is the entry point for backend
// inits all dependencies, connects external services
// and starts HTTP server
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jakobsym/aura/internal/handler"
	"github.com/jakobsym/aura/internal/repository/postgres"
	solana "github.com/jakobsym/aura/internal/repository/solanarpc"
	"github.com/jakobsym/aura/internal/routes"
	"github.com/jakobsym/aura/internal/service"
)

func main() {
	// Init Postgres connection pool
	db := postgres.PostgresConnectionPool()
	defer db.Close()

	// Init Solana RPC and Websocket connections
	rpcConnection := solana.SolanaRpcConnection()
	wsConnection := solana.SolanaWebSocketConnection()
	defer wsConnection.Close()

	// Init wallet tracking dependencies
	solanaAccountRepo := solana.NewSolanaWebSocketRepo(wsConnection)
	solanaAccountRepo.StartReader(context.Background()) // generalized reader for WS connection
	accountPsqlRepo := postgres.NewPostgresAccountRepo(db)
	solanaAccountService := service.NewAccountService(solanaAccountRepo, accountPsqlRepo)
	accountHandler := handler.NewAccountHandler(solanaAccountService)

	// Init token dependencies
	solanaTokenRepo := solana.NewSolanaTokenRepo(rpcConnection)
	psqlTokenRepo := postgres.NewPostgresTokenRepo(db)
	tokenService := service.NewTokenService(psqlTokenRepo, solanaTokenRepo)
	tokenHandler := handler.NewTokenHandler(tokenService)

	// Config HTTP routes
	router := routes.NewRouter(tokenHandler, accountHandler)
	ctx := context.Background()

	log.Println("service running on 3000")
	// Start monitoring actively tracked wallets
	if err := solanaAccountService.MonitorAccountSubsription(ctx); err != nil {
		log.Fatalf("failed to start monitoring: %v", err)
	}

	// Start HTTP server
	if err := http.ListenAndServe(":3000", router.LoadRoutes()); err != nil {
		panic(err)
	}

}
