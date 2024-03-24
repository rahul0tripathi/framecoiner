package repo

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rahul0tripathi/framecoiner/entity"
)

const (
	_tradeExpiry = time.Minute * 10
)

type TradesRepo struct {
	storage Storage
}

func NewTradesRepo(storage Storage) *TradesRepo {
	return &TradesRepo{storage: storage}
}

func (t *TradesRepo) LatestTrade(ctx context.Context, owner common.Address) (*entity.Trade, error) {
	value, err := t.storage.Read(ctx, entity.KeyTrades(owner))
	switch {
	case err == nil:
	case errors.Is(err, entity.ErrEmpty):
		return nil, entity.ErrNoTradesFound
	default:
		return nil, err
	}

	trade := &entity.Trade{}
	if err = json.Unmarshal([]byte(value), trade); err != nil {
		return nil, err
	}

	return trade, nil
}

func (t *TradesRepo) UpdateTrade(ctx context.Context, owner common.Address, trade *entity.Trade) error {
	value, err := json.Marshal(trade)
	if err != nil {
		return err
	}

	return t.storage.Write(ctx, entity.KeyTrades(owner), string(value), _tradeExpiry)
}
