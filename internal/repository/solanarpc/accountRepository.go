package solana

import (
	"context"

	solanarpc "github.com/gagliardetto/solana-go/rpc"
	"github.com/jakobsym/aura/internal/repository"
)

// TODO: This is actually a websocket connection NOT an RPC connection
type solanaAccountRepo struct {
	rpcClient *solanarpc.Client
}

func NewSolanaAccountRepo(c *solanarpc.Client) repository.SolanaAccountRepo {
	return &solanaAccountRepo{rpcClient: c}
}

// TODO: This needs your API_KEY attached
func SolanaWebSocketConnection() *solanarpc.Client {
	return solanarpc.New("wss://mainnet.helius-rpc.com")
}

func (sr *solanaAccountRepo) AccountSubscribe(ctx context.Context, walletAddress string) (interface{}, error) {
	return nil, nil
}

/*TODO:
- Code here will send request(s) to the Helius websocket RPC specifically accountSubscribe()

// subscribe to account notifications

// prase transaction data

// filter for data
*/
