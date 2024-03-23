package entity

import "errors"

var (
	ErrEmpty = errors.New("nil value")

	ErrNoAccountFound = errors.New("no account found")
)
