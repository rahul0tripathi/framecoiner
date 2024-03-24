package entity

type TradingAccount struct {
	Balance string `json:"balance"`
	Account string `json:"account"`
	Owner   string `json:"owner"`
}
