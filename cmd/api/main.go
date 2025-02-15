package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jakobsym/aura/internal/handler"
	solana "github.com/jakobsym/aura/internal/repository/solanarpc"
	"github.com/jakobsym/aura/internal/routes"
	"github.com/jakobsym/aura/internal/service"
)

func main() {
	//db := postgres.PostgresConnectionPool()
	accounts := []string{os.Getenv("WALLET_ADDRESS")}
	rpcConnection := solana.SolanaRpcConnection()
	wsConnection := solana.SolanaWebSocketConnection()
	//defer db.Close()

	solanaAccountRepo := solana.NewSolanaWebSocketRepo(wsConnection)
	solanaAccountService := service.NewAccountService(solanaAccountRepo, accounts)
	ctx := context.Background()

	//psqlRepo := postgres.NewPostgresTokenRepo(db)
	solanaTokenRepo := solana.NewSolanaTokenRepo(rpcConnection)
	tokenService := service.NewTokenService( /*psqlRepo,*/ solanaTokenRepo)
	tokenHandler := handler.NewTokenHandler(tokenService)

	router := routes.NewRouter(tokenHandler)

	log.Println("service running on 3000")
	if err := solanaAccountService.MonitorAccountSubsription(ctx); err != nil {
		log.Fatalf("failed to start monitoring: %v", err)
	}
	if err := http.ListenAndServe(":3000", router.LoadRoutes()); err != nil {
		panic(err)
	}

}
