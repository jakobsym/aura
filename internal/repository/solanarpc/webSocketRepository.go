package solana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"

	bin "github.com/gagliardetto/binary"
	token_metadata "github.com/gagliardetto/metaplex-go/clients/token-metadata"
	solanago "github.com/gagliardetto/solana-go"
	solanarpc "github.com/gagliardetto/solana-go/rpc"
	"github.com/gorilla/websocket"
	"github.com/jakobsym/aura/internal/domain"
	"github.com/jakobsym/aura/internal/repository"
	_ "github.com/joho/godotenv/autoload"
)

type solanaWebSocketRepo struct {
	Websocket *websocket.Conn
	mu        sync.Mutex
	pending   sync.Map                        // subscription responses
	subs      []chan domain.HeliusLogResponse // active subscriptions
	//Accounts  []string
}

const (
	pongWait   = 45 * time.Second
	pingPeriod = 30 * time.Second
	readWait   = 50 * time.Second
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
		ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	log.Println("WebSocket conneciton established.")
	return ws
}

func (sr *solanaWebSocketRepo) HandleWebSocketConnection(ctx context.Context) {
	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	go func() {
		for {
			select {
			case <-pingTicker.C:
				sr.mu.Lock()
				// write deadling for ping
				if err := sr.Websocket.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
					sr.mu.Unlock()
					continue
				}
				// send ping
				if err := sr.Websocket.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Printf("failed to send ping: %v", err)
				} else {
					log.Println("Ping sent")
					sr.Websocket.SetReadDeadline(time.Now().Add(pongWait))
				}
				sr.mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (sr *solanaWebSocketRepo) StartReader(ctx context.Context) {
	go func() {
		// initial read deadline and pong handler
		//sr.Websocket.SetReadDeadline(time.Now().Add(readWait))

		sr.Websocket.SetPongHandler(func(string) error {
			log.Println("Received pong from server")
			sr.Websocket.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})

		for {
			// read raw message
			var rawRes json.RawMessage
			if err := sr.Websocket.ReadJSON(&rawRes); err != nil {
				log.Printf("Websocket read error: %v", err)
				return
			}

			//sr.Websocket.SetReadDeadline(time.Now().Add(readWait))

			// try accountSubscribe() response
			var accountSubscribeRes domain.HeliusSubscriptionResponse
			if err := json.Unmarshal(rawRes, &accountSubscribeRes); err == nil && accountSubscribeRes.ID != 0 {
				if ch, ok := sr.pending.Load(accountSubscribeRes.ID); ok {
					ch.(chan domain.HeliusSubscriptionResponse) <- accountSubscribeRes
					sr.pending.Delete(accountSubscribeRes.ID)
				}
				continue
			}

			//			log.Printf("Unhandled message type: %s", string(rawRes))
			var logResponse domain.HeliusLogResponse
			if err := json.Unmarshal([]byte(rawRes), &logResponse); err != nil && logResponse.Method == "logsNotification" {
				sr.mu.Lock()
				for _, sub := range sr.subs {
					select {
					case sub <- logResponse:
					default:
						log.Println("Sub channel full, dropping notification")
					}
				}
				sr.mu.Unlock()
				continue
			}

			txnSignature := logResponse.Params.Result.Value.Signature
			payload, err := sr.GetTxnData(txnSignature)
			if err != nil {
				log.Printf("Error getting txn details from signature: %v", err)
				continue
			}
			swapData, err := sr.GetTxnSwapData(payload)
			if err != nil {
				log.Printf("Error getting txn swap data from payload: %v", err)
			}
			log.Println(swapData)

		}

	}()
}

func (sr *solanaWebSocketRepo) LogsSubscribe(ctx context.Context, walletAddress string, userId int) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	msg := domain.HeliusRequest{
		JsonRPC: "2.0",
		ID:      userId,
		Method:  "logsSubscribe",
		Params: []any{
			domain.LogsSubscribeParams{
				Mentions: []string{walletAddress},
			},
			domain.CommitmentConfig{
				Commitment: "finalized",
			},
		},
	}

	responseCh := make(chan domain.HeliusSubscriptionResponse, 1)
	sr.pending.Store(msg.ID, responseCh)
	defer sr.pending.Delete(msg.ID)
	if err := sr.Websocket.WriteJSON(msg); err != nil {
		return fmt.Errorf("failed to send logsSubscription request: %w", err)
	}
	select {
	case res := <-responseCh:
		if res.Error != nil {
			return fmt.Errorf("subscription error: %v", res.Error.Message)
		}
		log.Printf("Subscribed to %s | ID: %d\n", walletAddress, res.Result)
		return nil
	case <-time.After(30 * time.Second):
		return fmt.Errorf("subscription timeout")
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while awaiting subscription response")
	}

}

func (sr *solanaWebSocketRepo) GetTxnData(signature string) (domain.TransactionResult, error) {
	msg := domain.HeliusRequest{
		JsonRPC: "2.0",
		ID:      1,
		Method:  "getTransaction",
		Params: []any{
			signature,
			map[string]any{
				"encoding":                       "json",
				"maxSupportedTransactionVersion": 0,
			},
		},
	}
	reqMsg, err := json.Marshal(msg)
	if err != nil {
		return domain.TransactionResult{}, err
	}
	url := fmt.Sprintf("https://mainnet.helius-rpc.com/?api-key=%s", os.Getenv("HELIUS_API_KEY"))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqMsg))
	if err != nil {
		return domain.TransactionResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return domain.TransactionResult{}, err
	}
	defer res.Body.Close()
	var payload domain.TransactionResult
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return domain.TransactionResult{}, err
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return domain.TransactionResult{}, err
	}

	return payload, nil
}

// TODO: Not working as intended, currently returns empty array
func (sr *solanaWebSocketRepo) GetTxnSwapData(payload domain.TransactionResult) ([]domain.SwapResult, error) {
	userWalletAddress := payload.Result.Transaction.Message.AccountKeys[0]
	balanceMap := make(map[int]map[string]domain.TokenBalance)
	var swaps []domain.SwapResult

	// build balanceMap
	for _, pre := range payload.Result.Meta.PreTokenBalances {
		if balanceMap[pre.AccountIndex] == nil {
			balanceMap[pre.AccountIndex] = make(map[string]domain.TokenBalance)
		}
		balanceMap[pre.AccountIndex][pre.Mint] = pre
	}

	// process post balance i.e: find swaps
	var sent, received []domain.TokenBalance
	for _, post := range payload.Result.Meta.PostTokenBalances {
		if post.Owner != userWalletAddress {
			continue
		}

		pre, exists := balanceMap[post.AccountIndex][post.Mint]
		if !exists {
			continue
		}

		delta := post.UITokenAmount.UIAmount - pre.UITokenAmount.UIAmount
		if delta < 0 {
			sent = append(sent, domain.TokenBalance{
				Mint: post.Mint,
				UITokenAmount: domain.UITokenAmount{
					UIAmount: -delta,
				},
			})
		} else if delta > 0 {
			received = append(received, domain.TokenBalance{
				Mint: post.Mint,
				UITokenAmount: domain.UITokenAmount{
					UIAmount: delta,
				},
			})
		}
	}

	// Pair sent and received tokens
	for i := 0; i < len(sent) && i < len(received); i++ {

		sentTokenDetail, err := sr.GetTokenNameAndSymbol(context.TODO(), sent[i].Mint)
		if err != nil {
			return []domain.SwapResult{}, err
		}
		receivedTokenDetail, err := sr.GetTokenNameAndSymbol(context.TODO(), received[i].Mint)
		if err != nil {
			return []domain.SwapResult{}, err
		}

		swaps = append(swaps, domain.SwapResult{
			SentAmount:      sent[i].UITokenAmount.UIAmount,
			SentSymbol:      sentTokenDetail[1],
			SentAddress:     sent[i].Mint,
			ReceivedAddress: received[i].Mint,
			ReceivedAmount:  received[i].UITokenAmount.UIAmount,
			ReceivedSymbol:  receivedTokenDetail[1],
		})
	}
	return swaps, nil
}

func (sr *solanaWebSocketRepo) GetTokenNameAndSymbol(ctx context.Context, tokenAddress string) ([]string, error) {
	mint := solanago.MustPublicKeyFromBase58(tokenAddress)
	rpcClient := solanarpc.New("https://api.mainnet-beta.solana.com")
	seeds := [][]byte{
		[]byte("metadata"),
		token_metadata.ProgramID.Bytes(),
		mint.Bytes(),
	}
	mdAddr, _, err := solanago.FindProgramAddress(seeds, token_metadata.ProgramID)
	if err != nil {
		return []string{}, fmt.Errorf("unable to find metadata address: %w", err)
	}
	acc, err := rpcClient.GetAccountInfo(context.Background(), mdAddr)
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

// TODO: Fix to follow AccountSubscribe()
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
func (sr *solanaWebSocketRepo) AccountListen(ctx context.Context) (<-chan domain.HeliusLogResponse, error) {
	updates := make(chan domain.HeliusLogResponse, 10)
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.subs = append(sr.subs, updates)
	return updates, nil
}

func (sr *solanaWebSocketRepo) StopAccountListen(ch <-chan domain.HeliusLogResponse) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	for i, sub := range sr.subs {
		if sub == ch {
			sr.subs = append(sr.subs[:i], sr.subs[i+1:]...)
			close(sub)
			break
		}
	}
}

func (sr *solanaWebSocketRepo) AccountSubscribe(ctx context.Context, walletAddress string, userId int) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()
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
	responseCh := make(chan domain.HeliusSubscriptionResponse, 1)
	sr.pending.Store(msg.ID, responseCh)
	defer sr.pending.Delete(msg.ID)

	if err := sr.Websocket.WriteJSON(msg); err != nil {
		return fmt.Errorf("failed to send subscription request: %w", err)
	}

	select {
	case res := <-responseCh:
		if res.Error != nil {
			return fmt.Errorf("subscription error: %v", res.Error.Message)
		}
		log.Printf("Subscribed to %s | ID: %d\n", walletAddress, res.Result)
		return nil
	case <-time.After(30 * time.Second):
		return fmt.Errorf("subscription timeout")
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while awaiting subscription response")
	}
}
