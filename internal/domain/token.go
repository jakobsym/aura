package domain

type Token struct {
	TokenAddress string  `json:"token_address"`
	Name         string  `json:"name,omitempty"`
	Symbol       string  `json:"symbol,omitempty"`
	Price        float64 `json:"price,omitempty"`
	Socials      string  `json:"socials,omitempty"`
}
