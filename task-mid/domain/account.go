package domain

import "errors"

var (
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrAccountFrozen       = errors.New("account is frozen")
)

type AccountID string

type AccountStatus string

const (
	AccountStatusActive AccountStatus = "ACTIVE"
	AccountStatusFrozen AccountStatus = "FROZEN"
)

const (
	FieldBalance = "balance"
	FieldStatus  = "status"
)

type ChangeTracker struct {
	dirty map[string]struct{}
}

func (t *ChangeTracker) MarkDirty(field string) {
	if t.dirty == nil {
		t.dirty = make(map[string]struct{})
	}
	t.dirty[field] = struct{}{}
}

func (t *ChangeTracker) IsDirty(field string) bool {
	if t == nil {
		return false
	}
	_, ok := t.dirty[field]
	return ok
}

type Account struct {
	id      AccountID
	balance int64
	status  AccountStatus

	Changes ChangeTracker
}

func NewAccount(id AccountID, balance int64, status AccountStatus) *Account {
	return &Account{
		id:      id,
		balance: balance,
		status:  status,
	}
}

func (a *Account) ID() AccountID {
	return a.id
}

func (a *Account) Balance() int64 {
	return a.balance
}

func (a *Account) Status() AccountStatus {
	return a.status
}

func (a *Account) Withdraw(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if a.status == AccountStatusFrozen {
		return ErrAccountFrozen
	}
	if a.balance < amount {
		return ErrInsufficientBalance
	}

	a.balance -= amount
	a.Changes.MarkDirty(FieldBalance)
	return nil
}

func (a *Account) Deposit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if a.status == AccountStatusFrozen {
		return ErrAccountFrozen
	}

	a.balance += amount
	a.Changes.MarkDirty(FieldBalance)
	return nil
}

func (a *Account) Freeze() error {
	if a.status == AccountStatusFrozen {
		return ErrAccountFrozen
	}

	a.status = AccountStatusFrozen
	a.Changes.MarkDirty(FieldStatus)
	return nil
}
