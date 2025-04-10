package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jakobsym/aura/internal/domain"
	"github.com/jakobsym/aura/internal/service"
)

type AccountHandler struct {
	as *service.AccountService
}

func NewAccountHandler(as *service.AccountService) *AccountHandler {
	return &AccountHandler{as: as}
}

// I.E: Using TG API to send live updates back to TG
func (ah *AccountHandler) TrackWallet(w http.ResponseWriter, r *http.Request) {
	walletAddress := chi.URLParam(r, "wallet_address")
	if walletAddress == "" {
		http.Error(w, "must provide valid wallet address", http.StatusBadRequest)
		return
	}
	var user domain.User
	// read TG ID from body
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "error decoding response body", http.StatusBadRequest)
		return
	}

	if err := ah.as.TrackWallet(walletAddress, user.TelegramId); err != nil {
		log.Printf("failed to track wallet: %v", err)
		http.Error(w, "error TrackWallet()", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("success")
}

func (ah *AccountHandler) CreateUserEntry(w http.ResponseWriter, r *http.Request) {
	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "error decoding req body", http.StatusBadRequest)
		return
	}
	if err := ah.as.CreateUser(user.TelegramId); err != nil {
		http.Error(w, "error creating user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("success")
}

// TODO: `UntrackWallet() not implemented`
func (ah *AccountHandler) UntrackWallet(w http.ResponseWriter, r *http.Request) {
	walletAddress := chi.URLParam(r, "wallet_address")
	if walletAddress == "" {
		http.Error(w, "must provide valid wallet address", http.StatusBadRequest)
		return
	}
	var user domain.User
	// read TG ID from body
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "error decoding req body", http.StatusBadRequest)
		return
	}
	if err := ah.as.UntrackWallet(walletAddress, user.TelegramId); err != nil {
		http.Error(w, "failed to untrack wallet", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("success")
}
