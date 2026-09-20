package wallet

type WalletID string

type OwnerID string

type Status string

const (
	StatusActive Status = "ACTIVE"
	StatusFrozen Status = "FROZEN"
)

type Wallet struct {
	id      WalletID
	ownerID OwnerID
	balance int64
	status  Status
}

func NewWallet(id WalletID, ownerID OwnerID, balance int64) *Wallet {
	return &Wallet{
		id:      id,
		ownerID: ownerID,
		balance: balance,
		status:  StatusActive,
	}
}

func (w *Wallet) ID() WalletID {
	return w.id
}

func (w *Wallet) OwnerID() OwnerID {
	return w.ownerID
}

func (w *Wallet) Balance() int64 {
	return w.balance
}

func (w *Wallet) Status() Status {
	return w.status
}

func (w *Wallet) Deposit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if w.status == StatusFrozen {
		return ErrWalletFrozen
	}

	w.balance += amount
	return nil
}

func (w *Wallet) Withdraw(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if w.status == StatusFrozen {
		return ErrWalletFrozen
	}
	if w.balance < amount {
		return &InsufficientBalanceError{
			Required:  amount,
			Available: w.balance,
		}
	}

	w.balance -= amount
	return nil
}

func (w *Wallet) Freeze() error {
	if w.status == StatusFrozen {
		return ErrWalletFrozen
	}

	w.status = StatusFrozen
	return nil
}
