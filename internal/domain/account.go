package domain

type HeliusRequest struct {
	JsonRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type AccountResponse struct {
	Result struct {
		Context interface{}   `json:"context"`
		Value   AccountUpdate `json:"value"`
	} `json:"result"`
}

type AccountUpdate struct {
	AccountData struct {
		Lamports   int64  `json:"lamports"`
		Owner      string `json:"owner"`
		Data       string `json:"data"`
		Executable bool   `json:"executable"`
		RentEpoch  int64  `json:"rentEpoch"`
	} `json:"accountData"`
	Slot    int64  `json:"slot"`
	Account string `json:"account"`
}
