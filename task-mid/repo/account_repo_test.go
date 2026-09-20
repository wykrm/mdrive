package repo

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"moneytransfer/domain"
)

func TestAccountRepo_Retrieve(t *testing.T) {
	account := domain.NewAccount("a1", 100, domain.AccountStatusActive)
	repository := NewAccountRepo(nil, account)

	got, err := repository.Retrieve(context.Background(), "a1")
	if err != nil {
		t.Fatalf("Retrieve() error = %v", err)
	}
	if got != account {
		t.Fatalf("Retrieve() account = %p, want %p", got, account)
	}

	_, err = repository.Retrieve(context.Background(), "missing")
	if !errors.Is(err, ErrAccountNotFound) {
		t.Errorf("Retrieve(missing) error = %v, want %v", err, ErrAccountNotFound)
	}
}

func TestAccountRepo_RetrieveHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewAccountRepo().Retrieve(ctx, "a1")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Retrieve() error = %v, want %v", err, context.Canceled)
	}
}

func TestAccountRepo_UpdateMut(t *testing.T) {
	tests := []struct {
		name        string
		arrange     func(*domain.Account)
		wantUpdates map[string]interface{}
		wantNil     bool
	}{
		{
			name:    "returns nil for nil account",
			wantNil: true,
		},
		{
			name:    "returns nil for clean account",
			arrange: func(*domain.Account) {},
			wantNil: true,
		},
		{
			name: "includes only dirty balance",
			arrange: func(account *domain.Account) {
				if err := account.Deposit(25); err != nil {
					t.Fatalf("Deposit() error = %v", err)
				}
			},
			wantUpdates: map[string]interface{}{domain.FieldBalance: int64(125)},
		},
		{
			name: "includes only dirty status",
			arrange: func(account *domain.Account) {
				if err := account.Freeze(); err != nil {
					t.Fatalf("Freeze() error = %v", err)
				}
			},
			wantUpdates: map[string]interface{}{domain.FieldStatus: domain.AccountStatusFrozen},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var account *domain.Account
			if tt.arrange != nil {
				account = domain.NewAccount("a1", 100, domain.AccountStatusActive)
				tt.arrange(account)
			}

			mutation := NewAccountRepo().UpdateMut(account)
			if tt.wantNil {
				if mutation != nil {
					t.Fatalf("UpdateMut() = %#v, want nil", mutation)
				}
				return
			}
			if mutation == nil {
				t.Fatal("UpdateMut() = nil, want mutation")
			}
			if mutation.Table != "accounts" || mutation.ID != "a1" {
				t.Errorf("mutation identity = %s/%s, want accounts/a1", mutation.Table, mutation.ID)
			}
			if !reflect.DeepEqual(mutation.Updates, tt.wantUpdates) {
				t.Errorf("Updates = %#v, want %#v", mutation.Updates, tt.wantUpdates)
			}
		})
	}
}
