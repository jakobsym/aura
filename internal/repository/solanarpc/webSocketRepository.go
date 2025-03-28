package solana

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jakobsym/aura/internal/domain"
	"github.com/jakobsym/aura/internal/repository"
	_ "github.com/joho/godotenv/autoload"
)

type solanaWebSocketRepo struct {
	Websocket *websocket.Conn
	mu        sync.Mutex
	pending   sync.Map                      // subscription responses
	subs      []chan domain.AccountResponse // active subscriptions
	//Accounts  []string
}

const (
	pongWait   = 60 * time.Second    // server waits 60s for a ping
	pingPeriod = (pongWait * 9) / 10 // server sends ping every 54s
	readWait   = 60 * time.Second
	writeWait  = 10 * time.Second
)

func NewSolanaWebSocketRepo(ws *websocket.Conn) repository.SolanaWebSocketRepo {
	return &solanaWebSocketRepo{Websocket: ws, mu: sync.Mutex{}}
}

func SolanaWebSocketConnection() *websocket.Conn {
	url := fmt.Sprintf("wss://mainnet.helius-rpc.com/?api-key=%s", os.Getenv("HELIUS_API_KEY"))
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalf("unable to create ws connection: %v", err)
	}
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(readWait))
		return nil
	})
	log.Println("WebSocket conneciton established.")
	return ws
}

func (sr *solanaWebSocketRepo) StartReader(ctx context.Context) {
	go func() {
		for {
			// read raw message
			var rawRes json.RawMessage
			if err := sr.Websocket.ReadJSON(&rawRes); err != nil {
				log.Printf("Websocket read error: %v", err)
				return
			}

			// try accountSubscribe() response
			var accountSubscribeRes domain.HeliusSubscriptionResponse
			if err := json.Unmarshal(rawRes, &accountSubscribeRes); err != nil && accountSubscribeRes.ID != 0 {
				if ch, ok := sr.pending.Load(accountSubscribeRes.ID); ok {
					ch.(chan domain.HeliusSubscriptionResponse) <- accountSubscribeRes
					sr.pending.Delete(accountSubscribeRes.ID)
				}
				continue
			}

			// try account notification response
			var notif domain.AccountNotification
			if err := json.Unmarshal(rawRes, &notif); err != nil && notif.Method == "accountNotification" {
				res := domain.AccountResponse{
					Context: notif.Params.Result.Context,
					Value:   notif.Params.Result.Value,
				}
				sr.mu.Lock()
				for _, sub := range sr.subs {
					select {
					case sub <- res:
					default:
						log.Println("Sub channel full, dropping notification")
					}
				}
				sr.mu.Unlock()
				continue
			}
			log.Printf("Unhandled message type: %s", string(rawRes))
		}

	}()
}

// TODO: Use context to handle graceful shutodwn
// setting read deadlines here to handle idle periods
func (sr *solanaWebSocketRepo) HandleWebSocketConnection(ctx context.Context) {
	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	sr.Websocket.SetReadDeadline(time.Now().Add(readWait))

	go func() {
		for range pingTicker.C {
			sr.Websocket.SetWriteDeadline(time.Now().Add(writeWait))
			if err := sr.Websocket.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
			sr.Websocket.SetReadDeadline(time.Now().Add(readWait))
		}
	}()
}

// TODO: This has race conditions maybe revert to original version?
func (sr *solanaWebSocketRepo) AccountListen(ctx context.Context) (<-chan domain.AccountResponse, error) {
	updates := make(chan domain.AccountResponse)

	go func() {
		defer close(updates)
		for {
			select {
			case <-ctx.Done():
				sr.Websocket.Close()
				return
			default:
				var notif domain.AccountNotification
				err := sr.Websocket.ReadJSON(&notif)
				if err != nil {
					return
				}
				if notif.Method == "accountNotification" {
					res := domain.AccountResponse{
						Context: notif.Params.Result.Context,
						Value:   notif.Params.Result.Value,
					}
					updates <- res
				}
			}
		}
	}()
	return updates, nil
}

func (sr *solanaWebSocketRepo) AccountSubscribe(ctx context.Context, walletAddress string, userId int) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	log.Printf("walletAddress=%s\nuserId=%d", walletAddress, userId)
	msg := domain.HeliusRequest{
		JsonRPC: "2.0",
		ID:      userId,
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
		log.Printf("websocket WriteJSON; AccountSubscribe(): %v", err)
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

func (sr *solanaWebSocketRepo) AccountUnsubscribe(ctx context.Context, walletAddress string, userId int) (bool, error) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	msg := domain.HeliusRequest{
		JsonRPC: "2.0",
		ID:      userId,
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
