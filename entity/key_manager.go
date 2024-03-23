package entity

type KeyManager struct {
	Account    string `json:"account"`
	Owner      string `json:"owner"`
	SigningKey string `json:"signingKey"`
}
