package solana

import (
	solanarpc "github.com/gagliardetto/solana-go/rpc"
	"github.com/jakobsym/aura/internal/repository"
)

type solanaTokenRepo struct {
	rpcClient *solanarpc.Client
}

func NewSolanaTokenRepo(c *solanarpc.Client) repository.SolanaTokenRepo {
	return &solanaTokenRepo{rpcClient: c}
}

func SolanaRpcConnection() *solanarpc.Client {
	return solanarpc.New("https://api.mainnet-beta.solana.com")
}
