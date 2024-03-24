package integrations

import (
	"context"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-resty/resty/v2"
	"github.com/rahul0tripathi/framecoiner/entity"
)

const (
	_zeroXURL = "https://api.0x.org/swap/v1"
)

type zeroXQuoteResponse struct {
	Code                 int    `json:"code"`
	ChainId              int    `json:"chainId"`
	Price                string `json:"price"`
	GrossPrice           string `json:"grossPrice"`
	EstimatedPriceImpact string `json:"estimatedPriceImpact"`
	Value                string `json:"value"`
	BuyTokenAddress      string `json:"buyTokenAddress"`
	BuyAmount            string `json:"buyAmount"`
	GrossBuyAmount       string `json:"grossBuyAmount"`
	SellTokenAddress     string `json:"sellTokenAddress"`
	SellAmount           string `json:"sellAmount"`
	GrossSellAmount      string `json:"grossSellAmount"`
	SellTokenToEthRate   string `json:"sellTokenToEthRate"`
	BuyTokenToEthRate    string `json:"buyTokenToEthRate"`
	To                   string `json:"to"`
	From                 string `json:"from"`
	Data                 string `json:"data"`
}

type ZeroXConfig struct {
	ApiKey  string
	ChainID string
}

type ZeroXSwapper struct {
	client *resty.Client
	cfg    ZeroXConfig
}

func NewZeroXSwapper(cfg ZeroXConfig) (*ZeroXSwapper, error) {
	return &ZeroXSwapper{client: resty.New().SetBaseURL(_zeroXURL), cfg: cfg}, nil
}

func (z *ZeroXSwapper) GetQuote(ctx context.Context, token common.Address, ethIn string) (*entity.Quote, error) {
	response := &zeroXQuoteResponse{}
	query := map[string]string{
		"buyToken":                        token.Hex(),
		"sellAmount":                      ethIn,
		"sellToken":                       "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		"priceImpactProtectionPercentage": "0.4",
	}

	resp, err := z.client.R().SetContext(ctx).SetHeader("0x-api-key", z.cfg.ApiKey).SetHeader("0x-chain-id", z.cfg.ChainID).SetQueryParams(query).Get("/quote")
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, err
	}

	if response.Code != 0 {
		return nil, entity.ErrNoQuoteFound
	}

	return &entity.Quote{
		To:                response.To,
		Value:             response.Value,
		CallData:          response.Data,
		BuyTokenToEthRate: response.BuyTokenToEthRate,
	}, nil
}
