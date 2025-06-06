// Package `solana` provides implementations of repository interfaces using Solana RPC methods,
// and external API calls
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

// `solanaWebSocketRepo` implements SolanaWebSocketRepo interface
// for interacting with real-time data via Helius RPC websockets
type solanaWebSocketRepo struct {
	Websocket *websocket.Conn
	mu        sync.Mutex
	pending   sync.Map                        // subscription responses
	subs      []chan domain.HeliusLogResponse // active subscriptions
	//Accounts  []string
}

// websocket connection logic constants
const (
	pongWait   = 45 * time.Second
	pingPeriod = 30 * time.Second
	readWait   = 50 * time.Second
	writeWait  = 10 * time.Second
)

// `NewSolanaWebSocketRepo` creates a new Solana websocket repository intstance
func NewSolanaWebSocketRepo(ws *websocket.Conn) repository.SolanaWebSocketRepo {
	return &solanaWebSocketRepo{Websocket: ws, mu: sync.Mutex{}}
}

// `SolanaWebSocketConnection` establishes a websocket connection to a Helius RPC endpoint
// Configures a pong handler to establish ping/pong connection between websocket and local server.
// returns the connection
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

// `HandleWebSocketConnection` manages websocket connection by implementing
// a ping/pong mechanism to keep the connection alive
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

// `StartReader` continuously reads messages from the websocket
// processing and dispatching them to the appropriate handlers.
// Handles subscription responses, log notifications, and
// extracts txn data when available.
func (sr *solanaWebSocketRepo) StartReader(ctx context.Context) {
	go func() {
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

			// try accountSubscribe() response
			var accountSubscribeRes domain.HeliusSubscriptionResponse
			if err := json.Unmarshal(rawRes, &accountSubscribeRes); err == nil && accountSubscribeRes.ID != 0 {
				if ch, ok := sr.pending.Load(accountSubscribeRes.ID); ok {
					ch.(chan domain.HeliusSubscriptionResponse) <- accountSubscribeRes
					sr.pending.Delete(accountSubscribeRes.ID)
				}
				continue
			}

			// try logResponse
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
			// process txn data
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

// `LogsSubscribe` subscribes to logs for a specific wallet address
// sends a subscription request and awaits for confirmation.
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
	// create channel for response and store in pending map
	responseCh := make(chan domain.HeliusSubscriptionResponse, 1)
	sr.pending.Store(msg.ID, responseCh)
	defer sr.pending.Delete(msg.ID)

	// send sub request
	if err := sr.Websocket.WriteJSON(msg); err != nil {
		return fmt.Errorf("failed to send logsSubscription request: %w", err)
	}

	// await for response w/ timeout
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

// `GetTxnData` retrieves detailed txn data for a given txn signature
// by making an RPC call to Helius RPC
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

	// send request
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

	// process data
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

// `GetTxnSwapData` analyzes txn data to identify token swaps
// extracting details regarding sent/recieved tokens to determine
// balance changes for a tracked wallet.
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

		// calculate token amount changes
		delta := post.UITokenAmount.UIAmount - pre.UITokenAmount.UIAmount
		if delta < 0 {
			// token was sent
			sent = append(sent, domain.TokenBalance{
				Mint: post.Mint,
				UITokenAmount: domain.UITokenAmount{
					UIAmount: -delta,
				},
			})
		} else if delta > 0 {
			// token was received
			received = append(received, domain.TokenBalance{
				Mint: post.Mint,
				UITokenAmount: domain.UITokenAmount{
					UIAmount: delta,
				},
			})
		}
	}

	// pair sent and received tokens to identify swaps
	for i := 0; i < len(sent) && i < len(received); i++ {
		// get metadata for sent/received tokens
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

// `GetTokenNameAndSymbol` retrieves the name and symbol for a Solana token
// by fetching and decoding its metadata account
// returns a string slice [name, symbol]
func (sr *solanaWebSocketRepo) GetTokenNameAndSymbol(ctx context.Context, tokenAddress string) ([]string, error) {
	mint := solanago.MustPublicKeyFromBase58(tokenAddress)
	rpcClient := solanarpc.New("https://api.mainnet-beta.solana.com")

	// find where metadata is stored
	// using token mint, and token programID
	seeds := [][]byte{
		[]byte("metadata"),
		token_metadata.ProgramID.Bytes(),
		mint.Bytes(),
	}
	/* Extraction */
	mdAddr, _, err := solanago.FindProgramAddress(seeds, token_metadata.ProgramID)
	if err != nil {
		return []string{}, fmt.Errorf("unable to find metadata address: %w", err)
	}
	acc, err := rpcClient.GetAccountInfo(context.Background(), mdAddr)
	if err != nil {
		return []string{}, fmt.Errorf("unable to find account info: %w", err)
	}

	/* Transformation */
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

// `AccountListen` creates a channel for recieving account notifications
// returns a read only channel that receives a HeliusLogResponse
func (sr *solanaWebSocketRepo) AccountListen(ctx context.Context) (<-chan domain.HeliusLogResponse, error) {
	updates := make(chan domain.HeliusLogResponse, 10)
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.subs = append(sr.subs, updates)
	return updates, nil
}

// `StopAccountListen` unsubscribes from account notifications
// by removing specified channel from subscription list and closing it
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

/*
 ** Deprecated **
 */
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

/*
 ** Deprecated **
 */
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
