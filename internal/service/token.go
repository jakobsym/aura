// Package `service` calls repository methods to implement business logic
package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jakobsym/aura/internal/domain"
	"github.com/jakobsym/aura/internal/repository"
)

// `TokenSerivce` provides business logic for token operations by receiving data
// from PostgresTokenRepo and SolanaTokenRepo
type TokenService struct {
	psqlRepo   repository.PostgresTokenRepo
	solanaRepo repository.SolanaTokenRepo
}

// `NewTokenService` creates and returns a new TokenService with required dependencies
func NewTokenService(r repository.PostgresTokenRepo, sr repository.SolanaTokenRepo) *TokenService {
	return &TokenService{psqlRepo: r, solanaRepo: sr}
}

// `GetTokenData` retrieves token metadata from multiple sources
// concurrently fetches metadata, supply, price, and age data from Solana
// calculates FDV and persists the data
func (ts *TokenService) GetTokenData(ctx context.Context, tokenAddress string) (*domain.TokenResponse, error) {
	var (
		age struct {
			age time.Time
			err error
		}
		price struct {
			price float64
			err   error
		}
		supply struct {
			supply float64
			err    error
		}
		md struct {
			metadata []string
			err      error
		}
		hasPrice  bool
		hasSupply bool
		recieved  int
		fdv       float64
	)

	socials := fmt.Sprintf("https://x.com/search?q=%s", tokenAddress)

	// buffered channels for concurrent data retrieval
	supplyCh := make(chan struct {
		supply float64
		err    error
	}, 1)
	priceCh := make(chan struct {
		price float64
		err   error
	}, 1)
	ageCh := make(chan struct {
		age time.Time
		err error
	}, 1)
	metadataCh := make(chan struct {
		metadata []string
		err      error
	}, 1)

	// concurrent data retrieval
	go func() {
		md, err := ts.solanaRepo.GetTokenNameAndSymbol(ctx, tokenAddress)
		metadataCh <- struct {
			metadata []string
			err      error
		}{md, err}
		recieved++
	}()
	go func() {
		supply, err := ts.solanaRepo.GetTokenSupply(ctx, tokenAddress)
		supplyCh <- struct {
			supply float64
			err    error
		}{supply, err}
		recieved++
		hasSupply = true
	}()

	go func() {
		price, err := ts.solanaRepo.GetTokenPrice(ctx, tokenAddress)
		priceCh <- struct {
			price float64
			err   error
		}{price, err}
		recieved++
	}()

	go func() {
		age, err := ts.solanaRepo.GetTokenAge(ctx, tokenAddress)
		ageCh <- struct {
			age time.Time
			err error
		}{age, err}
		recieved++
		hasPrice = true
	}()

	// process channels as they are filled
	for {
		select {
		case age = <-ageCh:
			if age.err != nil {
				return nil, fmt.Errorf("failed to fetch age: %w", age.err)
			}
		case supply = <-supplyCh:
			if supply.err != nil {
				return nil, fmt.Errorf("failed to supply: %w", supply.err)
			}
		case price = <-priceCh:
			if price.err != nil {
				return nil, fmt.Errorf("failed to fetch price: %w", price.err)
			}
		case md = <-metadataCh:
			if md.err != nil {
				return nil, fmt.Errorf("failed to fetch metadata: %w", md.err)
			}
		}

		if hasPrice && hasSupply {
			fdv = ts.solanaRepo.GetTokenFDV(ctx, price.price, supply.supply)
		}

		if recieved == 4 {
			break
		}
	}

	token := &domain.TokenResponse{
		Address:   tokenAddress,
		Name:      md.metadata[0],
		Symbol:    md.metadata[1],
		CreatedAt: age.age,
		Supply:    supply.supply,
		Price:     price.price,
		FDV:       fdv,
		Socials:   socials,
	}

	// insert this token into DB
	err := ts.psqlRepo.CreateToken(*token)
	if err != nil {
		log.Printf("unable to store token in db")
	}

	return token, nil
}

// `DeleteToken` removes a token entry from the DB
func (ts *TokenService) DeleteToken(ctx context.Context, tokenAddress string) error {
	err := ts.psqlRepo.DeleteToken(tokenAddress)
	if err != nil {
		return err
	}
	return nil
}
