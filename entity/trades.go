package entity

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type TradeRequest struct {
	Owner   string `json:"owner"`
	EthIn   string `json:"ethIn"`
	ToToken string `json:"toToken"`
}

type Trade struct {
	Owner   string    `json:"owner"`
	TxnHash string    `json:"txnHash"`
	Error   string    `json:"error"`
	Expiry  time.Time `json:"expiry"`
	Request Quote     `json:"request"`
}

func KeyTrades(owner common.Address) string {
	return fmt.Sprintf("TRADES:%s", owner.Hex())
}
