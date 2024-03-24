package services

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rahul0tripathi/framecoiner/entity"
)

type keyManager interface {
	SigningAddress(ctx context.Context, owner common.Address) (common.Address, error)
	SignTx(
		ctx context.Context,
		owner common.Address,
		transaction *types.Transaction,
		chainID *big.Int,
	) (*types.Transaction, error)
}

type quoter interface {
	GetQuote(ctx context.Context, token common.Address, ethIn string) (*entity.Quote, error)
}

type tradesRepo interface {
	LatestTrade(ctx context.Context, owner common.Address) (*entity.Trade, error)
	UpdateTrade(ctx context.Context, owner common.Address, trade *entity.Trade) error
}

type tradeProcessor interface {
	Submit(ctx context.Context, job *entity.TradeRequest) error
}
