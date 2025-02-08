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
	//defer db.Close()

	//psqlRepo := postgres.NewPostgresTokenRepo(db)
	solanaRepo := solana.NewSolanaTokenRepo(rpcConnection)
	service := service.NewTokenService( /*psqlRepo,*/ solanaRepo)
	handler := handler.NewTokenHandler(service)

	router := routes.NewRouter(handler)
	log.Println("service running on 3000")
	if err := http.ListenAndServe(":3000", router.LoadRoutes()); err != nil {
		panic(err)
	}
}
