package solana

import (
	"context"
	"fmt"
	"regexp"
	"time"

	bin "github.com/gagliardetto/binary"
	token_metadata "github.com/gagliardetto/metaplex-go/clients/token-metadata"
	solanago "github.com/gagliardetto/solana-go"
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
	return 0.0, nil
}

func (sr *solanaTokenRepo) GetTokenSupply(ctx context.Context, tokenAddress string) (uint64, error) {
	return 0, nil
}

func (sr *solanaTokenRepo) GetTokenFDV(ctx context.Context, price float64, supply uint64) (float64, error) {
	return 0, nil
}

// TODO:  Need to find the correct metaplex methods to call
func (sr *solanaTokenRepo) GetTokenNameAndSymbol(ctx context.Context, tokenAddress string) ([]string, error) {
	mint := solanago.MustPublicKeyFromBase58(tokenAddress)

	seeds := [][]byte{
		[]byte("metadata"),
		token_metadata.ProgramID.Bytes(),
		mint.Bytes(),
	}
	mdAddr, _, err := solanago.FindProgramAddress(seeds, token_metadata.ProgramID)
	if err != nil {
		return []string{}, fmt.Errorf("unable to find metadata address: %w", err)
	}
	acc, err := sr.rpcClient.GetAccountInfo(context.Background(), mdAddr)
	if err != nil {
		return []string{}, fmt.Errorf("unable to find account info: %w", err)
	}
	data := acc.Value.Data.GetBinary()

	var metadata token_metadata.Metadata
	decoder := bin.NewBorshDecoder(data)
	if err := metadata.UnmarshalWithDecoder(decoder); err != nil {
		return []string{}, fmt.Errorf("unable to deserialize data: %w", err)
	}

	r, _ := regexp.Compile(`\x00+`)
	name := r.ReplaceAllString(metadata.Data.Name, "")
	symbol := r.ReplaceAllString(metadata.Data.Symbol, "")
	return []string{name, symbol}, nil
}

func (sr *solanaTokenRepo) GetTokenAge(ctx context.Context, tokenAddress string) (time.Time, error) {
	return time.Now(), nil
}
