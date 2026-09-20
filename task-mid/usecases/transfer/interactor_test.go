package transfer

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"moneytransfer/contracts"
	"moneytransfer/domain"
	accountrepo "moneytransfer/repo"
)

type fakeAccountRepository struct {
	accounts      map[domain.AccountID]*domain.Account
	retrieveError map[domain.AccountID]error
	retrieved     []domain.AccountID
	updated       []*domain.Account
	mutationRepo  *accountrepo.AccountRepo
}

func newFakeAccountRepository(accounts ...*domain.Account) *fakeAccountRepository {
	byID := make(map[domain.AccountID]*domain.Account, len(accounts))
	for _, account := range accounts {
		byID[account.ID()] = account
	}
	return &fakeAccountRepository{
		accounts:      byID,
		retrieveError: make(map[domain.AccountID]error),
		mutationRepo:  accountrepo.NewAccountRepo(),
	}
}

func (r *fakeAccountRepository) Retrieve(_ context.Context, id domain.AccountID) (*domain.Account, error) {
	r.retrieved = append(r.retrieved, id)
	if err := r.retrieveError[id]; err != nil {
		return nil, err
	}
	return r.accounts[id], nil
}

func (r *fakeAccountRepository) UpdateMut(account *domain.Account) *contracts.Mutation {
	r.updated = append(r.updated, account)
	return r.mutationRepo.UpdateMut(account)
}

func TestInteractor_Execute(t *testing.T) {
	source := domain.NewAccount("source", 100, domain.AccountStatusActive)
	destination := domain.NewAccount("destination", 40, domain.AccountStatusActive)
	repository := newFakeAccountRepository(source, destination)
	uc := NewInteractor(repository)

	plan, err := uc.Execute(context.Background(), &TransferRequest{
		FromAccountID: "source",
		ToAccountID:   "destination",
		Amount:        25,
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got, want := source.Balance(), int64(75); got != want {
		t.Errorf("source balance = %d, want %d", got, want)
	}
	if got, want := destination.Balance(), int64(65); got != want {
		t.Errorf("destination balance = %d, want %d", got, want)
	}
	if got, want := repository.retrieved, []domain.AccountID{"source", "destination"}; !reflect.DeepEqual(got, want) {
		t.Errorf("retrieved accounts = %v, want %v", got, want)
	}
	if got, want := repository.updated, []*domain.Account{source, destination}; !reflect.DeepEqual(got, want) {
		t.Errorf("UpdateMut calls = %v, want %v", got, want)
	}

	mutations := plan.Mutations()
	if got, want := len(mutations), 2; got != want {
		t.Fatalf("mutation count = %d, want %d", got, want)
	}
	assertMutation(t, mutations[0], "source", 75)
	assertMutation(t, mutations[1], "destination", 65)

	// Mutations returns a slice copy; replacing an element cannot alter the plan.
	mutations[0] = nil
	if plan.Mutations()[0] == nil {
		t.Error("Mutations() exposed the plan's internal slice")
	}
}

func TestInteractor_ExecuteValidatesBeforeRepositoryCalls(t *testing.T) {
	tests := []struct {
		name    string
		request *TransferRequest
		wantErr error
	}{
		{name: "nil request", wantErr: ErrNilRequest},
		{name: "zero amount", request: &TransferRequest{FromAccountID: "source", ToAccountID: "destination"}, wantErr: domain.ErrInvalidAmount},
		{name: "negative amount", request: &TransferRequest{FromAccountID: "source", ToAccountID: "destination", Amount: -1}, wantErr: domain.ErrInvalidAmount},
		{name: "empty source ID", request: &TransferRequest{ToAccountID: "destination", Amount: 1}, wantErr: ErrInvalidAccountID},
		{name: "empty destination ID", request: &TransferRequest{FromAccountID: "source", Amount: 1}, wantErr: ErrInvalidAccountID},
		{name: "same account", request: &TransferRequest{FromAccountID: "source", ToAccountID: "source", Amount: 1}, wantErr: ErrSameAccount},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := newFakeAccountRepository()
			plan, err := NewInteractor(repository).Execute(context.Background(), tt.request)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Execute() error = %v, want %v", err, tt.wantErr)
			}
			if plan != nil {
				t.Errorf("Execute() plan = %#v, want nil", plan)
			}
			if len(repository.retrieved) != 0 || len(repository.updated) != 0 {
				t.Errorf("repository calls: retrieved=%v updated=%v, want none", repository.retrieved, repository.updated)
			}
		})
	}
}

func TestInteractor_ExecuteReturnsRetrieveErrorsAsIs(t *testing.T) {
	errSource := errors.New("source retrieval failed")
	errDestination := errors.New("destination retrieval failed")

	tests := []struct {
		name           string
		retrieveErrors map[domain.AccountID]error
		wantErr        error
		wantRetrieved  []domain.AccountID
	}{
		{
			name:           "source retrieval fails",
			retrieveErrors: map[domain.AccountID]error{"source": errSource},
			wantErr:        errSource,
			wantRetrieved:  []domain.AccountID{"source"},
		},
		{
			name:           "destination retrieval fails",
			retrieveErrors: map[domain.AccountID]error{"destination": errDestination},
			wantErr:        errDestination,
			wantRetrieved:  []domain.AccountID{"source", "destination"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := domain.NewAccount("source", 100, domain.AccountStatusActive)
			destination := domain.NewAccount("destination", 40, domain.AccountStatusActive)
			repository := newFakeAccountRepository(source, destination)
			repository.retrieveError = tt.retrieveErrors

			plan, err := NewInteractor(repository).Execute(context.Background(), &TransferRequest{
				FromAccountID: "source",
				ToAccountID:   "destination",
				Amount:        25,
			})

			if err != tt.wantErr {
				t.Fatalf("Execute() error = %v, want exact error %v", err, tt.wantErr)
			}
			if plan != nil {
				t.Errorf("Execute() plan = %#v, want nil", plan)
			}
			if got := repository.retrieved; !reflect.DeepEqual(got, tt.wantRetrieved) {
				t.Errorf("retrieved = %v, want %v", got, tt.wantRetrieved)
			}
			if source.Balance() != 100 || destination.Balance() != 40 {
				t.Errorf("balances changed: source=%d destination=%d", source.Balance(), destination.Balance())
			}
			if len(repository.updated) != 0 {
				t.Errorf("UpdateMut calls = %d, want 0", len(repository.updated))
			}
		})
	}
}

func TestInteractor_ExecuteReturnsWithdrawErrorAsIs(t *testing.T) {
	source := domain.NewAccount("source", 10, domain.AccountStatusActive)
	destination := domain.NewAccount("destination", 40, domain.AccountStatusActive)
	repository := newFakeAccountRepository(source, destination)

	plan, err := NewInteractor(repository).Execute(context.Background(), &TransferRequest{
		FromAccountID: "source",
		ToAccountID:   "destination",
		Amount:        25,
	})

	if err != domain.ErrInsufficientBalance {
		t.Fatalf("Execute() error = %v, want exact error %v", err, domain.ErrInsufficientBalance)
	}
	if plan != nil {
		t.Errorf("Execute() plan = %#v, want nil", plan)
	}
	if source.Balance() != 10 || destination.Balance() != 40 {
		t.Errorf("balances changed: source=%d destination=%d", source.Balance(), destination.Balance())
	}
	if len(repository.updated) != 0 {
		t.Errorf("UpdateMut calls = %d, want 0", len(repository.updated))
	}
}

func TestInteractor_ExecuteDepositFailureReturnsNoPlan(t *testing.T) {
	source := domain.NewAccount("source", 100, domain.AccountStatusActive)
	destination := domain.NewAccount("destination", 40, domain.AccountStatusFrozen)
	repository := newFakeAccountRepository(source, destination)

	plan, err := NewInteractor(repository).Execute(context.Background(), &TransferRequest{
		FromAccountID: "source",
		ToAccountID:   "destination",
		Amount:        25,
	})

	if err != domain.ErrAccountFrozen {
		t.Fatalf("Execute() error = %v, want exact error %v", err, domain.ErrAccountFrozen)
	}
	if plan != nil {
		t.Errorf("Execute() plan = %#v, want nil", plan)
	}
	if got, want := source.Balance(), int64(75); got != want {
		t.Errorf("source balance = %d, want %d", got, want)
	}
	if got, want := destination.Balance(), int64(40); got != want {
		t.Errorf("destination balance = %d, want %d", got, want)
	}
	if !source.Changes.IsDirty(domain.FieldBalance) {
		t.Error("source balance should remain dirty after successful Withdraw")
	}
	if destination.Changes.IsDirty(domain.FieldBalance) {
		t.Error("destination balance should remain clean after failed Deposit")
	}
	if len(repository.updated) != 0 {
		t.Errorf("UpdateMut calls = %d, want 0", len(repository.updated))
	}
}

func assertMutation(t *testing.T, mutation *contracts.Mutation, id string, balance int64) {
	t.Helper()
	if mutation == nil {
		t.Fatal("mutation is nil")
	}
	if mutation.Table != "accounts" || mutation.ID != id {
		t.Errorf("mutation identity = %s/%s, want accounts/%s", mutation.Table, mutation.ID, id)
	}
	wantUpdates := map[string]interface{}{domain.FieldBalance: balance}
	if !reflect.DeepEqual(mutation.Updates, wantUpdates) {
		t.Errorf("updates = %#v, want %#v", mutation.Updates, wantUpdates)
	}
}
