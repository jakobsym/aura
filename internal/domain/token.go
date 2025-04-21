// Package `domain` contains structs and types used throughout application
package domain

import "time"

// `Token` represents basic token information with optional JSON fields
type Token struct {
	TokenAddress string  `json:"token_address"`
	Name         string  `json:"name,omitempty"`
	Symbol       string  `json:"symbol,omitempty"`
	Price        float64 `json:"price,omitempty"`
	Socials      string  `json:"socials,omitempty"`
}

// `TokenResponse` represents transformed token data for API responses
// as well as DB token entries
type TokenResponse struct {
	Address   string    `json:"token_address"`
	Name      string    `json:"name"`
	Symbol    string    `json:"symbol"`
	CreatedAt time.Time `json:"created_at"`
	Supply    float64   `json:"supply"`
	Price     float64   `json:"price"`
	FDV       float64   `json:"fdv"`
	Socials   string    `json:"socials,omitempty"`
}
