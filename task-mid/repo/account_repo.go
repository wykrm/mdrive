package repo

import (
	"context"
	"errors"

	"moneytransfer/contracts"
	"moneytransfer/domain"
)

var ErrAccountNotFound = errors.New("account not found")

type AccountRepo struct {
	accounts map[domain.AccountID]*domain.Account
}

func NewAccountRepo(accounts ...*domain.Account) *AccountRepo {
	repository := &AccountRepo{
		accounts: make(map[domain.AccountID]*domain.Account, len(accounts)),
	}
	for _, account := range accounts {
		if account != nil {
			repository.accounts[account.ID()] = account
		}
	}
	return repository
}

func (r *AccountRepo) Retrieve(ctx context.Context, id domain.AccountID) (*domain.Account, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	account, ok := r.accounts[id]
	if !ok {
		return nil, ErrAccountNotFound
	}
	return account, nil
}

func (r *AccountRepo) UpdateMut(account *domain.Account) *contracts.Mutation {
	if account == nil {
		return nil
	}

	updates := make(map[string]interface{})
	if account.Changes.IsDirty(domain.FieldBalance) {
		updates[domain.FieldBalance] = account.Balance()
	}
	if account.Changes.IsDirty(domain.FieldStatus) {
		updates[domain.FieldStatus] = account.Status()
	}

	if len(updates) == 0 {
		return nil
	}

	return &contracts.Mutation{
		Table:   "accounts",
		ID:      string(account.ID()),
		Updates: updates,
	}
}

var _ contracts.AccountRepository = (*AccountRepo)(nil)
