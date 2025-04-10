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

type User struct {
	TelegramId int `json:"user_id"`
}

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

type UITokenAmount struct {
	UIAmount float64 `json:"uiAmount"`
}

type TokenBalance struct {
	AccountIndex  int           `json:"accountIndex"`
	Mint          string        `json:"mint"`
	Owner         string        `json:"owner"`
	UITokenAmount UITokenAmount `json:"uiTokenAmount"`
}

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
