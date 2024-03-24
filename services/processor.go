package services

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rahul0tripathi/framecoiner/entity"
	"github.com/rahul0tripathi/framecoiner/pkg/log"
	"go.uber.org/zap"
)

const (
	_tradesQueueBuffer = 10
	_tradeJobExpiry    = time.Minute * 10
	_queueTimeout      = time.Second * 4
)

type TradeProcessor struct {
	backend    *ethclient.Client
	manager    keyManager
	repo       tradesRepo
	swapQuoter quoter
	jobs       chan *entity.TradeRequest
	logger     log.Logger
	chainID    *big.Int
}

func NewTradeProcessor(
	manager keyManager,
	repo tradesRepo,
	swapQuoter quoter,
	client *ethclient.Client,
	logger log.Logger,
	chainID string,
) (*TradeProcessor, error) {
	chainIDInt, ok := new(big.Int).SetString(chainID, 10)
	if !ok {
		return nil, errors.New("failed to parse chainID")
	}

	return &TradeProcessor{
		manager:    manager,
		repo:       repo,
		swapQuoter: swapQuoter,
		jobs:       make(chan *entity.TradeRequest, _tradesQueueBuffer),
		backend:    client,
		logger:     logger,
		chainID:    chainIDInt,
	}, nil
}

func (t *TradeProcessor) Run(ctx context.Context, workers int) {
	for i := 0; i < workers; i++ {
		go t.worker(ctx)
	}
}

func (t *TradeProcessor) Submit(ctx context.Context, job *entity.TradeRequest) error {
	select {
	case <-time.After(_queueTimeout):
		return errors.New("failed to queue job")
	case t.jobs <- job:
		return nil
	}
}

func (t *TradeProcessor) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-t.jobs:
			err := t.trade(ctx, job)
			if err != nil {
				t.logger.Error("failed to execute job", zap.Any("job", job), zap.Error(err))
			}
		}
	}
}

func (t *TradeProcessor) trade(ctx context.Context, job *entity.TradeRequest) error {
	owner := common.HexToAddress(job.Owner)
	if err := t.repo.UpdateTrade(ctx, owner, &entity.Trade{
		Owner:   job.Owner,
		TxnHash: "",
		Error:   "",
		Expiry:  time.Now().Add(_tradeJobExpiry),
		Request: entity.Quote{},
	}); err != nil {
		return err
	}

	quote, err := t.swapQuoter.GetQuote(ctx, common.HexToAddress(job.ToToken), job.EthIn)
	if err != nil {
		err = t.repo.UpdateTrade(ctx, owner, &entity.Trade{
			Owner:   job.Owner,
			TxnHash: "",
			Error:   fmt.Sprintf("failed to get quote: %w", err),
			Expiry:  time.Now().Add(_tradeJobExpiry),
			Request: entity.Quote{},
		})
		return err
	}

	hash, err := t.relay(ctx, owner, quote)
	if err != nil {
		err = t.repo.UpdateTrade(ctx, owner, &entity.Trade{
			Owner:   job.Owner,
			TxnHash: "",
			Error:   fmt.Sprintf("failed to relay: %w", err),
			Expiry:  time.Now().Add(_tradeJobExpiry),
			Request: *quote,
		})
		return err
	}

	return t.repo.UpdateTrade(ctx, owner, &entity.Trade{
		Owner:   job.Owner,
		TxnHash: hash.Hex(),
		Error:   "",
		Expiry:  time.Now().Add(_tradeJobExpiry),
		Request: *quote,
	})
}

func (t *TradeProcessor) relay(ctx context.Context, owner common.Address, quote *entity.Quote) (*common.Hash, error) {
	signer, err := t.manager.SigningAddress(ctx, owner)
	if err != nil {
		return nil, err
	}

	nonce, err := t.backend.PendingNonceAt(ctx, signer)
	if err != nil {
		return nil, err
	}

	gasPrice, err := t.backend.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	target := common.HexToAddress(quote.To)
	data, err := hexutil.Decode(quote.CallData)
	if err != nil {
		return nil, err
	}

	value, ok := new(big.Int).SetString(quote.Value, 10)
	if !ok {
		return nil, errors.New("failed to parse value ")
	}

	gasLimit, err := t.backend.EstimateGas(ctx, ethereum.CallMsg{
		From:  signer,
		To:    &target,
		Value: value,
		Data:  data,
	})
	if err != nil {
		return nil, err
	}

	signed, err := t.manager.SignTx(ctx, owner, types.NewTx(
		&types.LegacyTx{
			Nonce:    nonce,
			GasPrice: gasPrice,
			Gas:      gasLimit,
			To:       &target,
			Value:    value,
			Data:     data,
		},
	), t.chainID)
	if err != nil {
		return nil, err
	}

	if err = t.backend.SendTransaction(ctx, signed); err != nil {
		return nil, err
	}

	hash := signed.Hash()
	return &hash, nil
}
