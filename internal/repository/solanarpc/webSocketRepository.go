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
	_ "github.com/joho/godotenv/autoload"
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

// TODO:  Make lifecycle/time for this connection?
func SolanaWebSocketConnection() *websocket.Conn {
	url := fmt.Sprintf("wss://mainnet.helius-rpc.com/?api-key=%s", os.Getenv("HELIUS_API_KEY"))
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalf("unable to create ws connection: %v", err)
	}
	log.Println("WebSocket conneciton established.")
	return ws
}

// TODO: To unsubscribe the ID field must be passed to unsubscribe given account
func (sr *solanaWebSocketRepo) AccountSubscribe(ctx context.Context, accounts []string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	for _, acc := range accounts {
		msg := domain.HeliusRequest{
			JsonRPC: "2.0",
			ID:      1,
			Method:  "accountSubscribe",
			Params: []interface{}{
				acc,
				map[string]interface{}{
					"commitment": "finalized",
					"encoding":   "jsonParsed",
				},
			},
		}
		fmt.Println(msg)
		if err := sr.Websocket.WriteJSON(msg); err != nil {
			return fmt.Errorf("failed to subscribe: %w", err)
		}
		var subscriptionRes domain.HeliusSubscriptionResponse
		if err := sr.Websocket.ReadJSON(&subscriptionRes); err != nil {
			return fmt.Errorf("failed to read subscription response: %w", err)
		}

		if subscriptionRes.Error != nil {
			return fmt.Errorf("error in subscription")
		}

		log.Printf("Subscribed to: %s || Subscription ID: %d", acc, subscriptionRes.Result)

	}
	//sr.Accounts = append(sr.Accounts, accounts...)
	return nil
}

func (sr *solanaWebSocketRepo) AccountUnsubscribe(ctx context.Context, account string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()
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
				var notif domain.AccountNotification
				if err := sr.Websocket.ReadJSON(&notif); err != nil {
					log.Printf("error reading message: %v", err)
					continue
				}
				if notif.Method == "accountNotification" {
					res := domain.AccountResponse{
						Context: notif.Params.Result.Context,
						Value:   notif.Params.Result.Value,
					}
					updates <- res
					//log.Printf("account update: %+v", res)
				}
			}
		}
	}()
	return updates, nil
}
