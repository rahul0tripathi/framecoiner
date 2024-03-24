package entity

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

type Signature struct {
	R [32]byte
	S [32]byte
	V uint8
}

func KeyAccount(account common.Address) string {
	return fmt.Sprintf("ACCOUNT:%s", account.Hex())
}
