package v1

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rahul0tripathi/framecoiner/entity"
)

type AccountService interface {
	GetTradingAccount(
		ctx context.Context,
		address common.Address,
	) (*entity.TradingAccount, error)
	PlaceTradeRequest(
		ctx context.Context,
		address common.Address,
		tokenAddress common.Address,
		ethIn string,
	) error
	LatestTrade(
		ctx context.Context,
		address common.Address,
	) (*entity.Trade, error)
}
