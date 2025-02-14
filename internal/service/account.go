package service

import "github.com/jakobsym/aura/internal/repository"

type AccountService struct {
	solanaRepo repository.SolanaAccountRepo
}

func NewAccountService(sr repository.SolanaAccountRepo) *AccountService {
	return &AccountService{solanaRepo: sr}
}

func (as *AccountService) AccountSubsription() (interface{}, error) {
	return nil, nil
}
