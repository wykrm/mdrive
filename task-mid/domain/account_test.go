package domain

import (
	"errors"
	"testing"
)

func TestNewAccount(t *testing.T) {
	account := NewAccount("a1", 100, AccountStatusActive)

	if got, want := account.ID(), AccountID("a1"); got != want {
		t.Errorf("ID() = %q, want %q", got, want)
	}
	if got, want := account.Balance(), int64(100); got != want {
		t.Errorf("Balance() = %d, want %d", got, want)
	}
	if got, want := account.Status(), AccountStatusActive; got != want {
		t.Errorf("Status() = %q, want %q", got, want)
	}
	if account.Changes.IsDirty(FieldBalance) || account.Changes.IsDirty(FieldStatus) {
		t.Error("new account must have no dirty fields")
	}
}

func TestAccount_Withdraw(t *testing.T) {
	tests := []struct {
		name        string
		status      AccountStatus
		amount      int64
		wantBalance int64
		wantErr     error
		wantDirty   bool
	}{
		{name: "withdraws funds", status: AccountStatusActive, amount: 40, wantBalance: 60, wantDirty: true},
		{name: "rejects zero", status: AccountStatusActive, amount: 0, wantBalance: 100, wantErr: ErrInvalidAmount},
		{name: "rejects negative amount", status: AccountStatusActive, amount: -1, wantBalance: 100, wantErr: ErrInvalidAmount},
		{name: "validates before state", status: AccountStatusFrozen, amount: -1, wantBalance: 100, wantErr: ErrInvalidAmount},
		{name: "rejects frozen account", status: AccountStatusFrozen, amount: 40, wantBalance: 100, wantErr: ErrAccountFrozen},
		{name: "rejects insufficient balance", status: AccountStatusActive, amount: 101, wantBalance: 100, wantErr: ErrInsufficientBalance},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := NewAccount("a1", 100, tt.status)
			err := account.Withdraw(tt.amount)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Withdraw() error = %v, want %v", err, tt.wantErr)
			}
			if got := account.Balance(); got != tt.wantBalance {
				t.Errorf("Balance() = %d, want %d", got, tt.wantBalance)
			}
			if got := account.Changes.IsDirty(FieldBalance); got != tt.wantDirty {
				t.Errorf("balance dirty = %t, want %t", got, tt.wantDirty)
			}
			if account.Changes.IsDirty(FieldStatus) {
				t.Error("Withdraw() marked status dirty")
			}
		})
	}
}

func TestAccount_Deposit(t *testing.T) {
	tests := []struct {
		name        string
		status      AccountStatus
		amount      int64
		wantBalance int64
		wantErr     error
		wantDirty   bool
	}{
		{name: "deposits funds", status: AccountStatusActive, amount: 40, wantBalance: 140, wantDirty: true},
		{name: "rejects zero", status: AccountStatusActive, amount: 0, wantBalance: 100, wantErr: ErrInvalidAmount},
		{name: "rejects negative amount", status: AccountStatusActive, amount: -1, wantBalance: 100, wantErr: ErrInvalidAmount},
		{name: "validates before state", status: AccountStatusFrozen, amount: -1, wantBalance: 100, wantErr: ErrInvalidAmount},
		{name: "rejects frozen account", status: AccountStatusFrozen, amount: 40, wantBalance: 100, wantErr: ErrAccountFrozen},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := NewAccount("a1", 100, tt.status)
			err := account.Deposit(tt.amount)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Deposit() error = %v, want %v", err, tt.wantErr)
			}
			if got := account.Balance(); got != tt.wantBalance {
				t.Errorf("Balance() = %d, want %d", got, tt.wantBalance)
			}
			if got := account.Changes.IsDirty(FieldBalance); got != tt.wantDirty {
				t.Errorf("balance dirty = %t, want %t", got, tt.wantDirty)
			}
		})
	}
}

func TestAccount_Freeze(t *testing.T) {
	account := NewAccount("a1", 100, AccountStatusActive)

	if err := account.Freeze(); err != nil {
		t.Fatalf("Freeze() error = %v", err)
	}
	if got := account.Status(); got != AccountStatusFrozen {
		t.Errorf("Status() = %q, want %q", got, AccountStatusFrozen)
	}
	if !account.Changes.IsDirty(FieldStatus) {
		t.Error("Freeze() did not mark status dirty")
	}
	if account.Changes.IsDirty(FieldBalance) {
		t.Error("Freeze() marked balance dirty")
	}
	if err := account.Freeze(); !errors.Is(err, ErrAccountFrozen) {
		t.Errorf("second Freeze() error = %v, want %v", err, ErrAccountFrozen)
	}
}

func TestChangeTracker_NilReceiverIsClean(t *testing.T) {
	var tracker *ChangeTracker
	if tracker.IsDirty(FieldBalance) {
		t.Error("nil ChangeTracker reported a dirty field")
	}
}
