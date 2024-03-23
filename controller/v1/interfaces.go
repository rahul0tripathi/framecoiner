package v1

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

type AccountService interface {
	GetTradingAccount(ctx context.Context, address common.Address) (string, error)
}
