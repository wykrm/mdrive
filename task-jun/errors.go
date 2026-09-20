package wallet

import (
	"errors"
	"fmt"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrWalletFrozen        = errors.New("wallet is frozen")
)

type InsufficientBalanceError struct {
	Required  int64
	Available int64
}

func (e *InsufficientBalanceError) Error() string {
	return fmt.Sprintf(
		"%s: required %d cents, available %d cents",
		ErrInsufficientBalance,
		e.Required,
		e.Available,
	)
}

func (e *InsufficientBalanceError) Is(target error) bool {
	return target == ErrInsufficientBalance
}
