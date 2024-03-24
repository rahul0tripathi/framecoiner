package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-resty/resty/v2"
	"github.com/rahul0tripathi/framecoiner/entity"
	"github.com/rahul0tripathi/framecoiner/integrations"
)

type MetadataService struct {
	client  *resty.Client
	backend *ethclient.Client
	cfg     integrations.ZeroXConfig
}

type priceResponse struct {
	Coins map[string]struct {
		Decimals  int     `json:"decimals"`
		Price     float64 `json:"price"`
		Symbol    string  `json:"symbol"`
		Timestamp float64 `json:"timestamp"`
	} `json:"coins"`
}

func NewTokenMetadataService(backend *ethclient.Client, cfg integrations.ZeroXConfig) *MetadataService {
	return &MetadataService{
		client:  resty.New().SetBaseURL("https://coins.llama.fi"),
		backend: backend,
		cfg:     cfg,
	}
}

func (m *MetadataService) GetTokenMetadata(ctx context.Context, token common.Address) (*entity.TokenMetadata, error) {
	response := &priceResponse{}
	resp, err := m.client.R().SetContext(ctx).Get(fmt.Sprintf("prices/current/base:%s?searchWidth=4h", token.Hex()))
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, err
	}

	binding, err := entity.NewErc20Binding(token, m.backend)
	if err != nil {
		return nil, err
	}

	symbol, err := binding.Symbol(&bind.CallOpts{
		Context: ctx,
	})
	if err != nil {
		return nil, err
	}

	metadata, ok := response.Coins[fmt.Sprintf("base:%s", token.String())]
	if !ok {
		return nil, errors.New("no price found")
	}

	return &entity.TokenMetadata{
		Ticker: symbol,
		Price:  fmt.Sprintf("%.5f", metadata.Price),
		Logo:   fmt.Sprintf("https://token-registry.s3.amazonaws.com/icons/tokens/base/128/%s.png", token.String()),
	}, nil
}
