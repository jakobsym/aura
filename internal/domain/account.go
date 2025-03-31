package domain

type HeliusRequest struct {
	JsonRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

type LogsSubscribeParams struct {
	Mentions []string `json:"mentions"`
}

type CommitmentConfig struct {
	Commitment string `json:"commitment"`
}

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

type HeliusSubscriptionResponse struct {
	JsonRPC string `json:"jsonrpc"`
	Result  int    `json:"result"`
	ID      int    `json:"id"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

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

type HeliusTransaction struct {
	Description string `json:"description"`
	Type        string `json:"type"`
	Source      string `json:"source"`
	Fee         int64  `json:"fee"`
	Timestamp   int64  `json:"timestamp"`
	Events      struct {
		Swap []SwapEvent `json:"swap"`
	} `json:"events"`
}

type SwapEvent struct {
	TokenIn     TokenInfo `json:"tokenIn"`
	TokenOut    TokenInfo `json:"tokenOut"`
	AmountIn    string    `json:"amountIn"`
	AmountOut   string    `json:"amountOut"`
	Source      string    `json:"source"`
	SourceLabel string    `json:"sourceLabel"`
}

type TokenInfo struct {
	Symbol    string `json:"symbol"`
	Address   string `json:"address"`
	Amount    string `json:"amount"`
	Decimals  int    `json:"decimals"`
	TokenName string `json:"tokenName"`
}

type AccountResponse struct {
	Context Context `json:"context"`
	Value   Value   `json:"value"`
}

type Context struct {
	Slot uint64 `json:"slot"`
}

type Value struct {
	Data       interface{} `json:"data"`
	Executable bool        `json:"executable"`
	Lamports   uint64      `json:"lamports"`
	Owner      string      `json:"owner"`
	RentEpoch  uint64      `json:"rentEpoch"`
}

type User struct {
	TelegramId int `json:"user_id"`
}

type HeliusUnsubscribeResponse struct {
	JsonRPC string `json:"jsonrpc"`
	Result  bool   `json:"result"`
	ID      int    `json:"id"`
}
