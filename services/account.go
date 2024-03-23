package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rahul0tripathi/framecoiner/entity"
)

type AccountService struct {
	storage Storage
}

func NewAccountService(storage Storage) *AccountService {
	return &AccountService{
		storage: storage,
	}
}

func (a *AccountService) GetTradingAccount(ctx context.Context, address common.Address) (string, error) {
	account, err := a.storage.Read(ctx, fmt.Sprintf("ACCOUNT:%s", address.Hex()))
	switch {
	case err == nil:
	case errors.Is(err, entity.ErrEmpty):
		return "", entity.ErrNoAccountFound
	case err != nil:
		return "", err
	}

	return account, nil
}
