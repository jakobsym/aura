package solana

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/jakobsym/aura/internal/domain"
	"github.com/jakobsym/aura/internal/repository"
)

type solanaWebSocketRepo struct {
	Websocket *websocket.Conn
	mu        sync.Mutex
	Accounts  []string
}

// TODO: Accounts will have to come from DB (future issue)
func NewSolanaWebSocketRepo(ws *websocket.Conn) repository.SolanaWebSocketRepo {
	return &solanaWebSocketRepo{Websocket: ws, mu: sync.Mutex{}, Accounts: make([]string, 0)}
}

func SolanaWebSocketConnection() *websocket.Conn {
	url := fmt.Sprintf("wss://api.helius.xyz/v0/ws?api-key=%s", os.Getenv("HELIUS_API_KEY"))
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalf("unable to create ws connection: %v", err)
	}
	return ws
}

// TODO: Not sure if an array of accounts can be passed in as a parameter
func (sr *solanaWebSocketRepo) AccountSubscribe(ctx context.Context, accounts []string) error {

	sr.mu.Lock()
	defer sr.mu.Unlock()

	msg := domain.HeliusRequest{
		JsonRPC: "2.0",
		Method:  "accountSubscribe",
		Params: []interface{}{
			accounts,
			map[string]interface{}{
				"commitment": "confirmed",
				"encoding":   "jsonParsed",
			},
		},
	}

	if err := sr.Websocket.WriteJSON(msg); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}
	sr.Accounts = append(sr.Accounts, accounts...)
	return nil
}

func (sr *solanaWebSocketRepo) AccountListen(ctx context.Context) (<-chan domain.AccountResponse, error) {
	updates := make(chan domain.AccountResponse)

	go func() {
		defer close(updates)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, message, err := sr.Websocket.ReadMessage()
				if err != nil {
					log.Printf("error reading message: %v", err)
					continue
				}
				log.Printf("Received message: %s", string(message))
			}
		}
	}()
	return updates, nil
}
