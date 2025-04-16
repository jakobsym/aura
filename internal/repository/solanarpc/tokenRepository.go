// Package `solana` provides implementations of repository interfaces using Solana RPC methods,
// and external API calls
package solana

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	bin "github.com/gagliardetto/binary"
	token_metadata "github.com/gagliardetto/metaplex-go/clients/token-metadata"
	solanago "github.com/gagliardetto/solana-go"
	solanarpc "github.com/gagliardetto/solana-go/rpc"
	"github.com/jakobsym/aura/internal/repository"
	"github.com/tidwall/gjson"
)

// `solanaTokenRepo` implements the solanarpc.SolanaTokenRepo interface using a solanarpc.Client
type solanaTokenRepo struct {
	rpcClient *solanarpc.Client
}

// `NewSolanaTokenRepo` creates and returns a new solanarpc.Client implementation
// of the SolanaTokenRepo interface.
func NewSolanaTokenRepo(c *solanarpc.Client) repository.SolanaTokenRepo {
	return &solanaTokenRepo{rpcClient: c}
}

// `SolanaRpcConnection` creates a new connection to Solana mainnet
// using the created solanarpc.Client
func SolanaRpcConnection() *solanarpc.Client {
	return solanarpc.New("https://api.mainnet-beta.solana.com")
}

// `GetTokenPrice` retrieves current price of a token in USD from the jupiter API
// TODO: Make such that denominating asset included in req body
// Currently only USD
func (sr *solanaTokenRepo) GetTokenPrice(ctx context.Context, tokenAddress string) (float64, error) {
	client := &http.Client{}
	url := fmt.Sprintf("https://api.jup.ag/price/v2?ids=%s", tokenAddress)
	req, err := http.NewRequest("GET", url, nil)
	//	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("error building req: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error receiving response: %w", err)
	}
	defer res.Body.Close()
	/*
		if res.StatusCode != http.StatusOK {
			return 0, fmt.Errorf("unexpected status code: %w", err)
		}
	*/
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading res body: %w", err)
	}
	price := gjson.Get(string(body), "data."+tokenAddress+".price")
	if !price.Exists() {
		return 0, fmt.Errorf("price not found: %s", tokenAddress)
	}
	return price.Float(), nil
}

// `GetTokenSupply` retrieves current ciculating supply for a given Solana tokenAddress
// returns supply as float64 for easier calculations
func (sr *solanaTokenRepo) GetTokenSupply(ctx context.Context, tokenAddress string) (float64, error) {
	mint := solanago.MustPublicKeyFromBase58(tokenAddress)
	out, err := sr.rpcClient.GetTokenSupply(ctx, mint, solanarpc.CommitmentFinalized)
	if err != nil {
		return 0, fmt.Errorf("error fetching token supply: %w", err)
	}
	supply := out.Value.UiAmountString
	supplyInt, err := strconv.ParseFloat(supply, 64)
	if err != nil {
		return 0, fmt.Errorf("error converting token supply: %w", err)
	}
	return supplyInt, nil
}

// `GetTokenFDV` calculates the fully dilued valuation of a given token
func (sr *solanaTokenRepo) GetTokenFDV(ctx context.Context, price float64, supply float64) float64 {
	return (price * supply)
}

// `GetTokenNameAndSymbol` retrieves the name and symbol for a Solana token
// by fetching and decoding its metadata account
// returns a string slice [name, symbol]
func (sr *solanaTokenRepo) GetTokenNameAndSymbol(ctx context.Context, tokenAddress string) ([]string, error) {
	mint := solanago.MustPublicKeyFromBase58(tokenAddress)

	// find where metadata is stored
	// using token mint, and token programID
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

	r, _ := regexp.Compile(`\x00+`) // remove null bytes
	name := r.ReplaceAllString(metadata.Data.Name, "")
	symbol := r.ReplaceAllString(metadata.Data.Symbol, "")
	return []string{name, symbol}, nil
}

// `GetTokenAge` determines when a token is created by finding its earliest transaction
// involving the metadata account
// returns creation time as UTC timestamp.
func (sr *solanaTokenRepo) GetTokenAge(ctx context.Context, tokenAddress string) (time.Time, error) {
	mint := solanago.MustPublicKeyFromBase58(tokenAddress)

	// find where metadata is stored
	// using token mint, and token programID
	seeds := [][]byte{
		[]byte("metadata"),
		token_metadata.ProgramID.Bytes(),
		mint.Bytes(),
	}
	mdAddr, _, err := solanago.FindProgramAddress(seeds, token_metadata.ProgramID)
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to find metadata address: %w", err)
	}
	sig, err := sr.rpcClient.GetSignaturesForAddress(ctx, mdAddr)
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to find account info: %w", err)
	}
	time := sig[len(sig)-1].BlockTime.Time().UTC()
	return time, nil
}
