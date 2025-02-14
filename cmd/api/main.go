package main

import (
	"log"
	"net/http"

	"github.com/jakobsym/aura/internal/handler"
	solana "github.com/jakobsym/aura/internal/repository/solanarpc"
	"github.com/jakobsym/aura/internal/routes"
	"github.com/jakobsym/aura/internal/service"
)

func main() {
	//db := postgres.PostgresConnectionPool()
	rpcConnection := solana.SolanaRpcConnection()
	wsConnection := solana.SolanaWebSocketConnection()
	//defer db.Close()

	solanaAccountRepo := solana.NewSolanaAccountRepo(wsConnection)
	accountService := service.NewAccountService(solanaAccountRepo)
	accountHandler := handler.NewAccountHandler(accountService)

	// TODO: Pass this account handler into internal websocket

	//psqlRepo := postgres.NewPostgresTokenRepo(db)
	solanaTokenRepo := solana.NewSolanaTokenRepo(rpcConnection)
	tokenService := service.NewTokenService( /*psqlRepo,*/ solanaTokenRepo)
	tokenHandler := handler.NewTokenHandler(tokenService)

	router := routes.NewRouter(tokenHandler)
	log.Println("service running on 3000")
	if err := http.ListenAndServe(":3000", router.LoadRoutes()); err != nil {
		panic(err)
	}
}
