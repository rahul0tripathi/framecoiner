package integrations

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rahul0tripathi/framecoiner/entity"
)

type key struct {
	Account    string `json:"account"`
	Owner      string `json:"owner"`
	SigningKey string `json:"signingKey"`
}

func (k *key) sign(transaction *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	signingKey, err := crypto.HexToECDSA(k.SigningKey)
	if err != nil {
		return nil, err
	}

	return types.SignTx(transaction, types.NewEIP155Signer(chainID), signingKey)
}

type KeyManager struct {
	storage Storage
}

func NewKeyManager(storage Storage) *KeyManager {
	return &KeyManager{storage: storage}
}

func (m *KeyManager) getAccount(ctx context.Context, owner common.Address) (*key, error) {
	value, err := m.storage.Read(ctx, entity.KeyAccount(owner))
	if err != nil {
		return nil, err
	}

	response := &key{}
	if err = json.Unmarshal([]byte(value), response); err != nil {
		return nil, err
	}

	return response, nil
}

func (m *KeyManager) createNewAccount(ctx context.Context, owner common.Address) (*key, error) {
	signingKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	publicKeyECDSA, ok := signingKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	metadata := &key{
		Account:    crypto.PubkeyToAddress(*publicKeyECDSA).Hex(),
		Owner:      owner.Hex(),
		SigningKey: hexutil.Encode(crypto.FromECDSA(signingKey))[2:],
	}

	seralized, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	if err = m.storage.Write(ctx, entity.KeyAccount(owner), string(seralized), 0); err != nil {
		return nil, err
	}

	return metadata, nil
}

func (m *KeyManager) SigningAddress(ctx context.Context, owner common.Address) (common.Address, error) {
	account, err := m.getAccount(ctx, owner)
	switch {
	case err == nil:
		return common.HexToAddress(account.Account), nil
	case !errors.Is(err, entity.ErrEmpty):
		return common.HexToAddress(""), err
	}

	created, err := m.createNewAccount(ctx, owner)
	if err != nil {
		return common.HexToAddress(""), nil
	}

	return common.HexToAddress(created.Account), nil
}

func (m *KeyManager) SignTx(
	ctx context.Context,
	owner common.Address,
	transaction *types.Transaction,
	chainID *big.Int,
) (*types.Transaction, error) {
	signingKey, err := m.getAccount(ctx, owner)
	if err != nil {
		return nil, err
	}

	return signingKey.sign(transaction, chainID)
}
