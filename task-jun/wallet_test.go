package wallet

import (
	"errors"
	"testing"
)

func TestNewWallet(t *testing.T) {
	w := NewWallet("w1", "owner1", 100)

	if got, want := w.ID(), WalletID("w1"); got != want {
		t.Errorf("ID() = %q, want %q", got, want)
	}
	if got, want := w.OwnerID(), OwnerID("owner1"); got != want {
		t.Errorf("OwnerID() = %q, want %q", got, want)
	}
	if got, want := w.Balance(), int64(100); got != want {
		t.Errorf("Balance() = %d, want %d", got, want)
	}
	if got, want := w.Status(), StatusActive; got != want {
		t.Errorf("Status() = %q, want %q", got, want)
	}
}

func TestWallet_Deposit(t *testing.T) {
	tests := []struct {
		name        string
		initial     int64
		amount      int64
		freeze      bool
		wantBalance int64
		wantErr     error
	}{
		{
			name:        "adds amount",
			initial:     100,
			amount:      50,
			wantBalance: 150,
		},
		{
			name:        "rejects zero",
			initial:     100,
			amount:      0,
			wantBalance: 100,
			wantErr:     ErrInvalidAmount,
		},
		{
			name:        "rejects negative amount",
			initial:     100,
			amount:      -50,
			wantBalance: 100,
			wantErr:     ErrInvalidAmount,
		},
		{
			name:        "rejects operation on frozen wallet",
			initial:     100,
			amount:      50,
			freeze:      true,
			wantBalance: 100,
			wantErr:     ErrWalletFrozen,
		},
		{
			name:        "validates amount before frozen state",
			initial:     100,
			amount:      -50,
			freeze:      true,
			wantBalance: 100,
			wantErr:     ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWallet("w1", "owner1", tt.initial)
			if tt.freeze {
				if err := w.Freeze(); err != nil {
					t.Fatalf("Freeze() error = %v", err)
				}
			}

			err := w.Deposit(tt.amount)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Deposit() error = %v, want %v", err, tt.wantErr)
			}
			if got := w.Balance(); got != tt.wantBalance {
				t.Errorf("Balance() = %d, want %d", got, tt.wantBalance)
			}
		})
	}
}

func TestWallet_Withdraw(t *testing.T) {
	tests := []struct {
		name        string
		initial     int64
		amount      int64
		freeze      bool
		wantBalance int64
		wantErr     error
	}{
		{
			name:        "subtracts amount",
			initial:     100,
			amount:      50,
			wantBalance: 50,
		},
		{
			name:        "allows entire balance",
			initial:     100,
			amount:      100,
			wantBalance: 0,
		},
		{
			name:        "rejects zero",
			initial:     100,
			amount:      0,
			wantBalance: 100,
			wantErr:     ErrInvalidAmount,
		},
		{
			name:        "rejects negative amount",
			initial:     100,
			amount:      -50,
			wantBalance: 100,
			wantErr:     ErrInvalidAmount,
		},
		{
			name:        "rejects operation on frozen wallet",
			initial:     100,
			amount:      50,
			freeze:      true,
			wantBalance: 100,
			wantErr:     ErrWalletFrozen,
		},
		{
			name:        "validates amount before frozen state",
			initial:     100,
			amount:      -50,
			freeze:      true,
			wantBalance: 100,
			wantErr:     ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWallet("w1", "owner1", tt.initial)
			if tt.freeze {
				if err := w.Freeze(); err != nil {
					t.Fatalf("Freeze() error = %v", err)
				}
			}

			err := w.Withdraw(tt.amount)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Withdraw() error = %v, want %v", err, tt.wantErr)
			}
			if got := w.Balance(); got != tt.wantBalance {
				t.Errorf("Balance() = %d, want %d", got, tt.wantBalance)
			}
		})
	}
}

func TestWallet_WithdrawInsufficientBalance(t *testing.T) {
	w := NewWallet("w1", "owner1", 100)

	err := w.Withdraw(150)

	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("Withdraw() error = %v, want errors.Is(_, ErrInsufficientBalance)", err)
	}

	var balanceErr *InsufficientBalanceError
	if !errors.As(err, &balanceErr) {
		t.Fatalf("Withdraw() error type = %T, want *InsufficientBalanceError", err)
	}
	if got, want := balanceErr.Required, int64(150); got != want {
		t.Errorf("Required = %d, want %d", got, want)
	}
	if got, want := balanceErr.Available, int64(100); got != want {
		t.Errorf("Available = %d, want %d", got, want)
	}
	if got, want := err.Error(), "insufficient balance: required 150 cents, available 100 cents"; got != want {
		t.Errorf("error message = %q, want %q", got, want)
	}
	if got, want := w.Balance(), int64(100); got != want {
		t.Errorf("Balance() = %d after failed withdrawal, want %d", got, want)
	}
}

func TestWallet_Freeze(t *testing.T) {
	w := NewWallet("w1", "owner1", 100)

	if err := w.Freeze(); err != nil {
		t.Fatalf("first Freeze() error = %v", err)
	}
	if got, want := w.Status(), StatusFrozen; got != want {
		t.Errorf("Status() = %q, want %q", got, want)
	}
	if err := w.Freeze(); !errors.Is(err, ErrWalletFrozen) {
		t.Errorf("second Freeze() error = %v, want %v", err, ErrWalletFrozen)
	}
}
