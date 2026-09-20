# Bug analysis

## Why `errors.Is` returns `false`

`errors.Is` first checks if the returned error equals the target. Here they're two different `WalletError` values:

- the sentinel has message `"insufficient balance"`;
- the returned one has message `"need 500, have 100"`.

Both fields count in struct equality, so they aren't equal. `WalletError` also has no `Is(error) bool` and no `Unwrap() error`, so `errors.Is` has nothing else to try and returns `false`.

## Fix

Add `Is` and compare the stable code instead of the message:

```go
type WalletError struct {
    Code    string
    Message string
}

func (e WalletError) Error() string {
    return e.Message
}

func (e WalletError) Is(target error) bool {
    targetError, ok := target.(WalletError)
    return ok && e.Code == targetError.Code
}

var ErrInsufficientBalance = WalletError{
    Code:    "E001",
    Message: "insufficient balance",
}
```

`Withdraw` doesn't need changes. The returned error with code `E001` now matches the sentinel with the same code, even though the messages differ.

In the submitted wallet implementation I used a sentinel plus a dedicated `InsufficientBalanceError` instead. Its `Is` matches the sentinel, and its fields keep the context values for `errors.As`.