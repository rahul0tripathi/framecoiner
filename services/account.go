package services

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rahul0tripathi/framecoiner/entity"
)

var (
	gwei = new(big.Float).SetInt64(1e9)
)

type AccountService struct {
	backend    *ethclient.Client
	keyManager keyManager
	processor  tradeProcessor
	repo       tradesRepo
}

func NewAccountService(
	manager keyManager,
	repo tradesRepo,
	processor tradeProcessor,
	backend *ethclient.Client,
) *AccountService {
	return &AccountService{
		keyManager: manager,
		repo:       repo,
		processor:  processor,
		backend:    backend,
	}
}

func (a *AccountService) GetTradingAccount(
	ctx context.Context,
	address common.Address,
) (*entity.TradingAccount, error) {
	account, err := a.keyManager.SigningAddress(ctx, address)
	if err != nil {
		return nil, err
	}

	resp := &entity.TradingAccount{
		Owner:   address.Hex(),
		Account: account.Hex(),
	}

	balance, err := a.backend.BalanceAt(ctx, account, nil)
	if err != nil {
		return nil, err
	}

	balanceETH := new(big.Float).Quo(new(big.Float).Quo(new(big.Float).SetInt(balance), gwei), gwei)
	resp.Balance = fmt.Sprintf("Ξ %s", balanceETH.String())
	return resp, nil
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
