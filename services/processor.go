package services

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rahul0tripathi/framecoiner/entity"
	"github.com/rahul0tripathi/framecoiner/pkg/log"
	"go.uber.org/zap"
)

const (
	_tradesQueueBuffer    = 10
	_tradeJobExpiry       = time.Minute * 10
	_queueTimeout         = time.Second * 4
	_maxReceiptFetch      = 3
	_receiptFetchInterval = time.Second * 2
)

type TradeProcessor struct {
	backend    *ethclient.Client
	manager    keyManager
	repo       tradesRepo
	swapQuoter quoter
	jobs       chan *entity.TradeRequest
	logger     log.Logger
	chainID    *big.Int
	erc20ABI   *abi.ABI
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

	erc20ABI, err := abi.JSON(strings.NewReader(entity.Erc20BindingMetaData.ABI))
	if err != nil {
		return nil, err
	}

	return &TradeProcessor{
		manager:    manager,
		repo:       repo,
		swapQuoter: swapQuoter,
		jobs:       make(chan *entity.TradeRequest, _tradesQueueBuffer),
		backend:    client,
		logger:     logger,
		chainID:    chainIDInt,
		erc20ABI:   &erc20ABI,
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
			Error:   fmt.Sprintf("failed to relay: %s", err.Error()),
			Expiry:  time.Now().Add(_tradeJobExpiry),
			Request: *quote,
		})
		return err
	}

	if err = t.waitForTransactionReceipt(ctx, *hash); err != nil {
		err = t.repo.UpdateTrade(ctx, owner, &entity.Trade{
			Owner:   job.Owner,
			TxnHash: "",
			Error:   fmt.Sprintf("failed to fetch receipt: %s", err.Error()),
			Expiry:  time.Now().Add(_tradeJobExpiry),
			Request: *quote,
		})
		return err
	}

	flushHash, err := t.flush(ctx, job)
	if err != nil {
		err = t.repo.UpdateTrade(ctx, owner, &entity.Trade{
			Owner:   job.Owner,
			TxnHash: "",
			Error:   fmt.Sprintf("failed to flush: %s", err.Error()),
			Expiry:  time.Now().Add(_tradeJobExpiry),
			Request: *quote,
		})
		return err
	}

	if flushHash == nil {
		flushHash = &entity.ZeroHash
	}

	return t.repo.UpdateTrade(ctx, owner, &entity.Trade{
		Owner:   job.Owner,
		TxnHash: fmt.Sprintf("trade: %s, flush: %s", hash.Hex(), flushHash),
		Error:   "",
		Expiry:  time.Now().Add(_tradeJobExpiry),
		Request: *quote,
	})
}

func (t *TradeProcessor) waitForTransactionReceipt(ctx context.Context, txHash common.Hash) error {
	checkStatus := func(receipt *types.Receipt) error {
		if receipt.Status == 0 {
			return errors.New("transaction failed")
		}

		return nil
	}

	for i := 0; i < _maxReceiptFetch; i++ {
		receipt, err := t.backend.TransactionReceipt(ctx, txHash)
		switch {
		case err == nil:
			return checkStatus(receipt)
		case errors.Is(err, ethereum.NotFound):
			<-time.After(_receiptFetchInterval)
			continue
		default:
			return err
		}
	}

	return errors.New("failed to fetch receipt")
}

func (t *TradeProcessor) flush(
	ctx context.Context,
	job *entity.TradeRequest,
) (*common.Hash, error) {
	signer, err := t.manager.SigningAddress(ctx, common.HexToAddress(job.Owner))
	if err != nil {
		return nil, err
	}

	token, err := entity.NewErc20Binding(common.HexToAddress(job.ToToken), t.backend)
	if err != nil {
		return nil, err
	}

	balance, err := token.BalanceOf(&bind.CallOpts{Context: ctx}, signer)
	if err != nil {
		return nil, err
	}

	if balance.String() == "0" {
		return nil, nil
	}

	callData, err := t.erc20ABI.Pack("transfer", common.HexToAddress(job.Owner), balance)
	if err != nil {
		return nil, err
	}

	return t.relay(ctx, common.HexToAddress(job.Owner), &entity.Quote{
		To:                job.ToToken,
		Value:             "0",
		CallData:          hexutil.Encode(callData),
		BuyTokenToEthRate: "0",
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
