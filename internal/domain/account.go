// Package `domain` contains structs and types used throughout application
package domain

// Represents standard JSON-RPC request format for Helius API calls
type HeliusRequest struct {
	JsonRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

// Parameters for subscribing to account log event(s)
type LogsSubscribeParams struct {
	Mentions []string `json:"mentions"`
}

// Specifies blockchain commitment level for queries
type CommitmentConfig struct {
	Commitment string `json:"commitment"`
}

// Represents response for subscription request(s)
type HeliusSubscriptionResponse struct {
	JsonRPC string `json:"jsonrpc"`
	Result  int    `json:"result"`
	ID      int    `json:"id"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Real time log data for subscribed accounts
type HeliusLogResponse struct {
	JsonRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		Result struct {
			Value struct {
				Signature string   `json:"signature"`
				Logs      []string `json:"logs"`
			} `json:"value"`
		} `json:"result"`
		Subscription int `json:"subscription"`
	} `json:"params"`
}

// Represents a User via TelegramId
type User struct {
	TelegramId int `json:"user_id"`
}

// Contains parsed transaction data w/ token balance changes
type TransactionResult struct {
	Result struct {
		Meta struct {
			PreTokenBalances  []TokenBalance `json:"preTokenBalances"`
			PostTokenBalances []TokenBalance `json:"postTokenBalances"`
		} `json:"meta"`
		Transaction struct {
			Message struct {
				AccountKeys []string `json:"accountKeys"`
			} `json:"message"`
		} `json:"transaction"`
	} `json:"result"`
}

// Represents token balance in float64 format
type UITokenAmount struct {
	UIAmount float64 `json:"uiAmount"`
}

// Contains token ownership information
type TokenBalance struct {
	AccountIndex  int           `json:"accountIndex"`
	Mint          string        `json:"mint"`
	Owner         string        `json:"owner"`
	UITokenAmount UITokenAmount `json:"uiTokenAmount"`
}

// Represents outcome of a token swap operation
type SwapResult struct {
	SentAmount      float64 `json:"sentAmount"`
	SentSymbol      string  `json:"sentSymbol"`
	SentAddress     string  `json:"sentAddress"`
	ReceivedAddress string  `json:"receivedAddress"`
	ReceivedAmount  float64 `json:"receivedAmount"`
	ReceivedSymbol  string  `json:"receivedSymbol"`
}

/*
** deprecated **
 */
type HeliusUnsubscribeResponse struct {
	JsonRPC string `json:"jsonrpc"`
	Result  bool   `json:"result"`
	ID      int    `json:"id"`
}

/*
** deprecated **
 */
type Value struct {
	Data       interface{} `json:"data"`
	Executable bool        `json:"executable"`
	Lamports   uint64      `json:"lamports"`
	Owner      string      `json:"owner"`
	RentEpoch  uint64      `json:"rentEpoch"`
}

/*
** deprecated **
 */
type Context struct {
	Slot uint64 `json:"slot"`
}

/*
** deprecated **
 */
type AccountNotification struct {
	JsonRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		Result struct {
			Context Context `json:"context"`
			Value   Value   `json:"value"`
		} `json:"result"`
		Subscription int `json:"subscription"`
	} `json:"params"`
}
