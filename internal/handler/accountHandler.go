package handler

import (
	"encoding/json"
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

// req/res related methods
// get results from account service and send them out

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

	if err := ah.as.TrackWallet(walletAddress, user.UserId); err != nil {
		http.Error(w, "failed to track wallet", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("success")
	// checks DB if wallet is already being tracked
	// if not create new entry for this wallet
}

func (ah *AccountHandler) GetWalletUpdates(w http.ResponseWriter, r *http.Request) {
	// as updates arrive, dispatch them to telegram
}
