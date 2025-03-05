package solana

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/jakobsym/aura/internal/domain"
	"github.com/jakobsym/aura/internal/repository"
	_ "github.com/joho/godotenv/autoload"
)

type solanaWebSocketRepo struct {
	Websocket *websocket.Conn
	mu        sync.Mutex
	//Accounts  []string
}

// TODO: Accounts will have to come from DB (future issue)
func NewSolanaWebSocketRepo(ws *websocket.Conn) repository.SolanaWebSocketRepo {
	return &solanaWebSocketRepo{Websocket: ws, mu: sync.Mutex{}}
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

func (sr *solanaWebSocketRepo) AccountSubscribe(ctx context.Context, walletAddress, userId string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	subscriptionId, err := strconv.Atoi(userId)
	if err != nil {
		return fmt.Errorf("error converting userId to int: %w", err)
	}
	msg := domain.HeliusRequest{
		JsonRPC: "2.0",
		ID:      subscriptionId,
		Method:  "accountSubscribe",
		Params: []any{
			walletAddress,
			map[string]any{
				"commitment": "finalized",
				"encoding":   "jsonParsed",
			},
		},
	}
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

	log.Printf("Subscribed to: %s || Subscription ID: %d", walletAddress, subscriptionRes.Result)
	return nil
}

func (sr *solanaWebSocketRepo) AccountUnsubscribe(ctx context.Context, walletAddress, userId string) (bool, error) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	subscriptionId, err := strconv.Atoi(userId)
	if err != nil {
		return false, err
	}

	msg := domain.HeliusRequest{
		JsonRPC: "2.0",
		ID:      subscriptionId,
		Method:  "accountUnsubscribe",
	}
	if err := sr.Websocket.WriteJSON(msg); err != nil {
		return false, fmt.Errorf("failed to send msg to ws: %w", err)
	}
	var unsubscribeResponse domain.HeliusUnsubscribeResponse
	if err := sr.Websocket.ReadJSON(&unsubscribeResponse); err != nil {
		return false, fmt.Errorf("failed to read msg to ws: %w", err)
	}
	return unsubscribeResponse.Result, nil
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
