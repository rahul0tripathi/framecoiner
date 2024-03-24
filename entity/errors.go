package entity

import "errors"

var (
	ErrEmpty = errors.New("nil value")

	ErrNoAccountFound = errors.New("no account found")
	ErrNoTradesFound  = errors.New("no trades found")

	ErrNoQuoteFound = errors.New("no quote found")
)
