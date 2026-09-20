package transfer

import (
	"context"
	"errors"

	"moneytransfer/contracts"
	"moneytransfer/domain"
)

var (
	ErrNilRequest       = errors.New("transfer request is nil")
	ErrInvalidAccountID = errors.New("invalid account ID")
	ErrSameAccount      = errors.New("source and destination accounts must differ")
)

type TransferRequest struct {
	FromAccountID domain.AccountID
	ToAccountID   domain.AccountID
	Amount        int64
}

type Interactor struct {
	repo contracts.AccountRepository
}

func NewInteractor(repository contracts.AccountRepository) *Interactor {
	return &Interactor{repo: repository}
}

func (uc *Interactor) Execute(ctx context.Context, req *TransferRequest) (*contracts.Plan, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	if req.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	if req.FromAccountID == "" || req.ToAccountID == "" {
		return nil, ErrInvalidAccountID
	}
	if req.FromAccountID == req.ToAccountID {
		return nil, ErrSameAccount
	}

	source, err := uc.repo.Retrieve(ctx, req.FromAccountID)
	if err != nil {
		return nil, err
	}
	destination, err := uc.repo.Retrieve(ctx, req.ToAccountID)
	if err != nil {
		return nil, err
	}

	if err = source.Withdraw(req.Amount); err != nil {
		return nil, err
	}
	if err = destination.Deposit(req.Amount); err != nil {
		return nil, err
	}

	plan := contracts.NewPlan()
	plan.Add(uc.repo.UpdateMut(source))
	plan.Add(uc.repo.UpdateMut(destination))
	return plan, nil
}
