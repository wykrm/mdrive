# Answers

## Q1: `Withdraw` succeeds and `Deposit` fails

Case: source active (100), destination frozen (40), transfer 25.

`source.Withdraw(25)` succeeds, so the in-memory source is 75 with `balance` dirty. `destination.Deposit(25)` returns `ErrAccountFrozen`, the destination stays at 40. `Execute` returns `domain.ErrAccountFrozen` and a nil plan, `UpdateMut` is never called, nothing is persisted.

The DB is fine, but the in-memory source isn't rolled back. After a failure the caller must discard and reload the aggregates before retrying.

## Q2: applying mutations separately

Balances 100 / 40, transfer 25. Source mutation applied first stores 75. If the destination mutation then fails (say the connection drops), it stays at 40, and 25 cents are lost even though the transfer returned an error.

One plan with both mutations lets the committer apply them in a single DB transaction: both stored or neither.

## Q3: dirty fields and concurrent updates

Writing only dirty fields stops stale values from overwriting unrelated data. Example: one request freezes an account while a transfer that loaded the old `ACTIVE` state changes the balance. A balance-only mutation keeps `FROZEN`. If it also wrote its stale `ACTIVE`, it would silently undo the freeze (lost update).

Narrow mutations also reduce write contention and show exactly what changed.

## Q4: always including every field

That's a full-row update from a possibly stale snapshot. Anything changed concurrently gets overwritten, even though this operation never touched it. A balance-only transfer would revert `FROZEN` back to stale `ACTIVE`.

Dirty fields avoid this. Same-field conflicts still need optimistic locking or transactions, but unrelated fields aren't at risk.