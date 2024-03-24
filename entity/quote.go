package entity

type Quote struct {
	To                string `json:"to"`
	Value             string `json:"value"`
	CallData          string `json:"callData"`
	BuyTokenToEthRate string `json:"buyTokenToEthRate"`
}
