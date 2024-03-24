package services

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rahul0tripathi/framecoiner/entity"
)

type AccountService struct {
	keyManager keyManager
	processor  tradeProcessor
	repo       tradesRepo
}

func NewAccountService(manager keyManager, repo tradesRepo, processor tradeProcessor) *AccountService {
	return &AccountService{
		keyManager: manager,
		repo:       repo,
		processor:  processor,
	}
}

func (a *AccountService) GetTradingAccount(ctx context.Context, address common.Address) (string, error) {
	account, err := a.keyManager.SigningAddress(ctx, address)
	if err != nil {
		return "", err
	}

	return account.Hex(), nil
}

func (a *AccountService) PlaceTradeRequest(
	ctx context.Context,
	address common.Address,
	tokenAddress common.Address,
	ethIn string,
) error {
	return a.processor.Submit(ctx, &entity.TradeRequest{
		Owner:   address.Hex(),
		EthIn:   ethIn,
		ToToken: tokenAddress.Hex(),
	})
}

func (a *AccountService) LatestTrade(
	ctx context.Context,
	address common.Address,
) (*entity.Trade, error) {
	return a.repo.LatestTrade(ctx, address)
}
