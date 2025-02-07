package solana

import (
	"context"
	"time"

	token_metadata "github.com/gagliardetto/metaplex-go/clients/token-metadata"
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

// These will interact with SOlana RPC directly
func (sr *solanaTokenRepo) GetTokenPrice(ctx context.Context, tokenAddress string) (float64, error) {
	return 0, nil
}

func (sr *solanaTokenRepo) GetTokenSupply(ctx context.Context, tokenAddress string) (uint64, error) {
	return 0, nil
}

func (sr *solanaTokenRepo) GetTokenFDV(ctx context.Context, price float64, supply uint64) (float64, error) {
	return 0, nil
}

func (sr *solanaTokenRepo) GetTokenMetadata(ctx context.Context, tokenAddress string) (*token_metadata.Metadata, error) {
	return nil, nil
}

func (sr *solanaTokenRepo) GetTokenAge(ctx context.Context, tokenAddress string) (time.Time, error) {
	return time.Now(), nil
}
