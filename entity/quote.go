package entity

import "github.com/ethereum/go-ethereum/common"

var (
	ZeroHash = common.HexToHash("")
)

type Quote struct {
	To                string `json:"to"`
	Value             string `json:"value"`
	CallData          string `json:"callData"`
	BuyTokenToEthRate string `json:"buyTokenToEthRate"`
}
